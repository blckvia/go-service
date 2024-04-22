package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

type ProjectPostgres struct {
	ctx    context.Context
	db     *pgx.Conn
	cache  r.Cache
	logger *zap.Logger
}

func NewProjectPostgres(ctx context.Context, db *pgx.Conn, cache r.Cache, logger *zap.Logger) *ProjectPostgres {
	return &ProjectPostgres{
		ctx:    ctx,
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (r *ProjectPostgres) Create(project models.Project) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) RETURNING id`, projectsTable)
	_, err := r.db.Prepare(r.ctx, "createProject", query)
	if err != nil {
		return 0, err
	}

	row := r.db.QueryRow(r.ctx, "createProject", project.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProjectPostgres) Update(projectID int, input models.UpdateProjects) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	var exists bool
	_, err = r.db.Prepare(r.ctx, "projectExists", fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", projectsTable))
	if err != nil {
		return err
	}

	err = tx.QueryRow(r.ctx, "projectExists", projectID).Scan(&exists)
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

func (r *ProjectPostgres) Delete(projectID int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, projectsTable)
	res, err := r.db.Exec(r.ctx, query, projectID)
	if err != nil {
		return err
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	key := fmt.Sprintf("project:%d", projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}
func (r *ProjectPostgres) GetAll(limit, offset int) (models.GetAllProjects, error) {
	var projects []models.Project
	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p LIMIT $1 OFFSET $2`, projectsTable)
	_, err := r.db.Prepare(r.ctx, "getAllProjects", query)
	if err != nil {
		return models.GetAllProjects{}, err
	}

	rows, err := r.db.Query(r.ctx, "getAllProjects", limit, offset)
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
	_, err = r.db.Prepare(r.ctx, "countAllProjects", fmt.Sprintf(`SELECT COUNT(p.id) FROM %s p`, projectsTable))
	if err != nil {
		return models.GetAllProjects{}, err
	}

	err = r.db.QueryRow(r.ctx, "countAllProjects").Scan(&total)
	if err != nil {
		return models.GetAllProjects{}, err
	}

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
func (r *ProjectPostgres) GetByID(projectID int) (models.Project, error) {
	var project models.Project
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
	_, err = r.db.Prepare(r.ctx, "getProjectByID", query)
	if err != nil {
		return project, err
	}

	row := r.db.QueryRow(r.ctx, "getProjectByID", projectID)
	if err := row.Scan(&project.ID, &project.Name, &project.CreatedAt); err != nil {
		return project, err
	}

	projectJson, err := json.Marshal(project)
	if err != nil {
		r.logger.Error("Failed to marshal project: %v", zap.Error(err))
		return project, err
	}
	err = r.cache.Set(r.ctx, key, string(projectJson), 1*time.Minute)
	if err != nil {
		r.logger.Error("Failed to cache project: %v", zap.Error(err))
	}

	return project, nil
}
