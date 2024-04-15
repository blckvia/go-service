package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

var ErrNotFound = errors.New("record not found")

type GoodsPostgres struct {
	db    *sqlx.DB
	cache r.Cache
}

func NewGoodsPostgres(db *sqlx.DB, cache r.Cache) *GoodsPostgres {
	return &GoodsPostgres{
		db:    db,
		cache: cache,
	}
}

func (r *GoodsPostgres) GetAll(limit, offset int) (models.GetAllGoods, error) {
	ctx := context.Background()
	key := fmt.Sprintf(`goods:all:%d:%d`, limit, offset)
	data, err := r.cache.Get(ctx, key)
	if err == nil {
		var response models.GetAllGoods
		err = json.Unmarshal([]byte(data), &response)
		if err != nil {
			return models.GetAllGoods{}, err
		}
		return response, nil
	}

	var goods []models.Goods
	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp LIMIT $1 OFFSET $2`, goodsTable)
	if err := r.db.Select(&goods, query, limit, offset); err != nil {
		return models.GetAllGoods{}, err
	}

	var total int
	err = r.db.Get(&total, fmt.Sprintf(`SELECT COUNT(*) FROM %s gp`, goodsTable))
	if err != nil {
		return models.GetAllGoods{}, err
	}

	var removed int
	err = r.db.Get(&removed, fmt.Sprintf(`SELECT COUNT(id) FROM %s gp WHERE gp.removed = true`, goodsTable))
	if err != nil {
		return models.GetAllGoods{}, err
	}

	meta := models.Meta{
		Total:   total,
		Removed: removed,
		Limit:   limit,
		Offset:  offset,
	}

	response := models.GetAllGoods{
		Meta:  meta,
		Goods: goods,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	err = r.cache.Set(ctx, key, string(jsonData), 1*time.Minute)
	if err != nil {
		return models.GetAllGoods{}, err
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

	ctx := context.Background()
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to invalidate cache for key %s: %v", key, err)
	}

	return nil
}

func (r *GoodsPostgres) Delete(goodsID, projectID int) error {
	query := fmt.Sprintf(`UPDATE %s SET removed = true WHERE id = $1 AND project_id = $2`, goodsTable)
	res, err := r.db.Exec(query, goodsID, projectID)
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

	ctx := context.Background()
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to invalidate cache for key %s: %v", key, err)
	}

	return nil
}

func (r *GoodsPostgres) Reprioritize(goodsID, projectID int, priority int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

	query := fmt.Sprintf(`UPDATE %s SET priority = $1 WHERE id = $2 AND project_id = $3`, goodsTable)
	_, err = tx.Exec(query, priority, goodsID, projectID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`UPDATE %s SET priority = priority + 1 WHERE project_id = $1 AND priority >= $2`, goodsTable)
	_, err = tx.Exec(query, projectID, priority)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to invalidate cache for key %s: %v", key, err)
	}

	return nil
}
