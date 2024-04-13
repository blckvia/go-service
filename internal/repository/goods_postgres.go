package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"go-service/internal/models"
)

var ErrNotFound = errors.New("record not found")

type GoodsPostgres struct {
	db *sqlx.DB
}

func NewGoodsPostgres(db *sqlx.DB) *GoodsPostgres {
	return &GoodsPostgres{db: db}
}

func (r *GoodsPostgres) GetAll(limit, offset int) (models.GetAllGoodsResponse, error) {
	var goods []models.Goods
	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp LIMIT $1 OFFSET $2`, goodsTable)
	if err := r.db.Select(&goods, query, limit, offset); err != nil {
		return models.GetAllGoodsResponse{}, err
	}

	var total int
	err := r.db.Get(&total, fmt.Sprintf(`SELECT COUNT(*) FROM %s gp`, goodsTable))
	if err != nil {
		return models.GetAllGoodsResponse{}, err
	}

	var removed int
	err = r.db.Get(&removed, fmt.Sprintf(`SELECT COUNT(id) FROM %s gp WHERE gp.removed = true`, goodsTable))
	if err != nil {
		return models.GetAllGoodsResponse{}, err
	}

	meta := models.Meta{
		Total:   total,
		Removed: removed,
		Limit:   limit,
		Offset:  offset,
	}

	response := models.GetAllGoodsResponse{
		Meta:  meta,
		Goods: goods,
	}
	return response, nil
}

func (r *GoodsPostgres) GetOne(goodsID, projectID int) (models.Goods, error) {
	var goods models.Goods
	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp WHERE gp.id = $1 AND gp.project_id = $2`, goodsTable)
	if err := r.db.Get(&goods, query, goodsID, projectID); err != nil {
		return goods, err
	}
	return goods, nil
}

func (r *GoodsPostgres) Create(projectID int, goods models.Goods) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (project_id, name, description, priority, removed) VALUES ($1, $2, $3, $4, $5) RETURNING id`, goodsTable)
	row := r.db.QueryRow(query, projectID, goods.Name, goods.Description, goods.Priority, goods.Removed)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *GoodsPostgres) Update(goodsID, projectID int, input models.UpdateGoods) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var exists bool
	err = tx.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable), goodsID, projectID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return ErrNotFound
	}

	_, err = tx.Exec(fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND project_id = $2 FOR UPDATE", goodsTable), goodsID, projectID)
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
	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description = $%d", argID))
		args = append(args, *input.Description)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, goodsTable, setQuery, argID)
	args = append(args, goodsID)
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

func (r *GoodsPostgres) Delete(goodsID, projectID int) error {
	query := fmt.Sprintf(`UPDATE %s SET remove = true WHERE id = $1 AND project_id = $2`, goodsTable)
	_, err := r.db.Exec(query, goodsID, projectID)
	return err
}
