package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"go-service/internal/models"
)

type ProjectPostgres struct {
	db *sqlx.DB
}

func NewProjectPostgres(db *sqlx.DB) *ProjectPostgres {
	return &ProjectPostgres{db: db}
}

func (r *ProjectPostgres) Create(project models.Project) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (name) VALUES ($1) RETURNING id`, projectsTable)
	row := r.db.QueryRow(query, project.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ProjectPostgres) Update(projectID int, input models.UpdateProjects) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var exists bool
	err = tx.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", projectsTable), projectID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return ErrNotFound
	}

	_, err = tx.Exec(fmt.Sprintf("SELECT * FROM %s WHERE id = $1 FOR UPDATE", projectsTable), projectID)
	if err != nil {
		tx.Rollback()
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
	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectPostgres) Delete(projectID int) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, projectsTable)
	res, err := r.db.Exec(query, projectID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *ProjectPostgres) GetAll(limit, offset int) (models.GetAllProjects, error) {
	var projects []models.Project
	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p LIMIT $1 OFFSET $2`, projectsTable)
	if err := r.db.Select(&projects, query, limit, offset); err != nil {
		return models.GetAllProjects{}, err
	}

	var total int
	err := r.db.Get(&total, fmt.Sprintf(`SELECT COUNT(*) FROM %s p`, projectsTable))
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
	var goods models.Project
	query := fmt.Sprintf(`SELECT p.id, p.name, p.created_at FROM %s p WHERE p.id = $1`, projectsTable)
	if err := r.db.Get(&goods, query, projectID); err != nil {
		return goods, err
	}
	return goods, nil
}
