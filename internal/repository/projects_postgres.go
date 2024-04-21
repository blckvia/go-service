package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"go-service/internal/models"
)

type ProjectPostgres struct {
	ctx context.Context
	db  *pgx.Conn
}

func NewProjectPostgres(ctx context.Context, db *pgx.Conn) *ProjectPostgres {
	return &ProjectPostgres{
		ctx: ctx,
		db:  db,
	}
}

func (r *ProjectPostgres) Create(ctx context.Context, project models.Project) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) RETURNING id`, projectsTable)
	row := r.db.QueryRow(ctx, query, project.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProjectPostgres) Update(ctx context.Context, projectID int, input models.UpdateProjects) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	var exists bool
	err = tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", projectsTable), projectID).Scan(&exists)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	if !exists {
		tx.Rollback(ctx)
		return ErrNotFound
	}

	_, err = tx.Exec(ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = $1 FOR UPDATE", projectsTable), projectID)
	if err != nil {
		tx.Rollback(ctx)
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
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectPostgres) Delete(ctx context.Context, projectID int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, projectsTable)
	res, err := r.db.Exec(ctx, query, projectID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
func (r *ProjectPostgres) GetAll(ctx context.Context, limit, offset int) (models.GetAllProjects, error) {
	var projects []models.Project
	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p LIMIT $1 OFFSET $2`, projectsTable)
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	err = r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s p`, projectsTable)).Scan(&total)
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
func (r *ProjectPostgres) GetByID(ctx context.Context, projectID int) (models.Project, error) {
	var project models.Project
	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p WHERE p.id = $1`, projectsTable)
	row := r.db.QueryRow(ctx, query, projectID)
	if err := row.Scan(&project.ID, &project.Name, &project.CreatedAt); err != nil {
		return project, err
	}

	return project, nil
}
