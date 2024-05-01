package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

type ProjectPostgres struct {
	ctx    context.Context
	db     *pgxpool.Pool
	cache  r.Cache
	logger *zap.Logger
	tracer trace.Tracer
}

func NewProjectPostgres(ctx context.Context, db *pgxpool.Pool, cache r.Cache, logger *zap.Logger, tracer trace.Tracer) *ProjectPostgres {
	return &ProjectPostgres{
		ctx:    ctx,
		db:     db,
		cache:  cache,
		logger: logger,
		tracer: tracer,
	}
}

func (r *ProjectPostgres) Create(ctx context.Context, project models.Project) (int, error) {
	var id int

	_, span := r.tracer.Start(ctx, "CreateProject")
	defer span.End()

	query := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) RETURNING id`, projectsTable)

	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "createProject", query)
	if err != nil {
		return 0, err
	}

	span.AddEvent("createProject", trace.WithAttributes(attribute.String("query", query)))
	row := pgxConn.QueryRow(r.ctx, "createProject", project.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProjectPostgres) Update(ctx context.Context, projectID int, input models.UpdateProjects) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	var exists bool
	//_, err = r.db.Prepare(r.ctx, "projectExists", fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", projectsTable))
	//if err != nil {
	//	return err
	//}

	err = tx.QueryRow(r.ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", projectsTable), projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.Exec(r.ctx, fmt.Sprintf("SELECT p.id, p.name, p.created_at FROM %s p WHERE id = $1 FOR UPDATE", projectsTable), projectID)
	if err != nil {
		return err
	}

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if input.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name = $%d", argID))
		args = append(args, *input.Name)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, projectsTable, setQuery, argID)
	args = append(args, projectID)
	_, err = tx.Exec(r.ctx, query, args...)
	if err != nil {
		return err
	}

	err = tx.Commit(r.ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("project:%d", projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

func (r *ProjectPostgres) Delete(ctx context.Context, projectID int) error {
	_, span := r.tracer.Start(ctx, "DeleteProject")
	defer span.End()

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, projectsTable)
	span.AddEvent("delete project", trace.WithAttributes(attribute.String("query", query)))
	res, err := r.db.Exec(r.ctx, query, projectID)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	span.AddEvent("invalidate project in cache", trace.WithAttributes(attribute.String("key", fmt.Sprintf("project:%d", projectID))))
	key := fmt.Sprintf("project:%d", projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("Invalidate error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}
func (r *ProjectPostgres) GetAll(сtx context.Context, limit, offset int) (models.GetAllProjects, error) {
	var projects []models.Project

	_, span := r.tracer.Start(сtx, "GetAllProjects")
	defer span.End()

	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p LIMIT $1 OFFSET $2`, projectsTable)

	span.AddEvent("get all projects", trace.WithAttributes(attribute.String("query", query)))
	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return models.GetAllProjects{}, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "getAllProjects", query)
	if err != nil {
		return models.GetAllProjects{}, err
	}

	rows, err := pgxConn.Query(r.ctx, "getAllProjects", limit, offset)
	if err != nil {
		return models.GetAllProjects{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.CreatedAt); err != nil {
			return models.GetAllProjects{}, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return models.GetAllProjects{}, err
	}

	var total int
	_, err = pgxConn.Prepare(r.ctx, "countAllProjects", fmt.Sprintf(`SELECT COUNT(p.id) FROM %s p`, projectsTable))
	if err != nil {
		return models.GetAllProjects{}, err
	}

	err = pgxConn.QueryRow(r.ctx, "countAllProjects").Scan(&total)
	if err != nil {
		return models.GetAllProjects{}, err
	}
	span.AddEvent("count all projects", trace.WithAttributes(attribute.Int("total", total)))

	meta := models.MetaProjects{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	response := models.GetAllProjects{
		Meta:     meta,
		Projects: projects,
	}

	return response, nil
}
func (r *ProjectPostgres) GetByID(ctx context.Context, projectID int) (models.Project, error) {
	var project models.Project

	_, span := r.tracer.Start(ctx, "GetByID")
	defer span.End()

	span.AddEvent("redis get", trace.WithAttributes(attribute.String("key", fmt.Sprintf("project:%d", projectID))))
	key := fmt.Sprintf("project:%d", projectID)
	cachedProject, err := r.cache.Get(r.ctx, key)
	if err == nil {
		err := json.Unmarshal([]byte(cachedProject), &project)
		if err != nil {
			r.logger.Error("Failed to unmarshal cached project: %v", zap.Error(err))
		} else {
			return project, nil
		}
	}

	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p WHERE p.id = $1`, projectsTable)

	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return project, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "getProjectByID", query)
	if err != nil {
		return project, err
	}

	row := pgxConn.QueryRow(r.ctx, "getProjectByID", projectID)
	if err := row.Scan(&project.ID, &project.Name, &project.CreatedAt); err != nil {
		return project, err
	}

	span.AddEvent("Write In JSON")
	projectJson, err := json.Marshal(project)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to marshal project: %v", zap.Error(err))
		return project, err
	}

	span.AddEvent("redis set", trace.WithAttributes(attribute.String("key", fmt.Sprintf("project:%d", projectID))))
	err = r.cache.Set(r.ctx, key, string(projectJson), 1*time.Minute)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to cache project: %v", zap.Error(err))
	}

	return project, nil
}
