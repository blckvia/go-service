package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

var ErrNotFound = errors.New("record not found")

type GoodsPostgres struct {
	ctx    context.Context
	db     *pgx.Conn
	cache  r.Cache
	logger *zap.Logger
	//nats  *nats.Conn
}

func NewGoodsPostgres(ctx context.Context, db *pgx.Conn, cache r.Cache, logger *zap.Logger /*nats *nats.Conn*/) *GoodsPostgres {
	return &GoodsPostgres{
		ctx:    ctx,
		db:     db,
		cache:  cache,
		logger: logger,
		//nats:  nats,
	}
}

// GetAll get all Goods
func (r *GoodsPostgres) GetAll(limit, offset int) (models.GetAllGoods, error) {
	var goods []models.Goods
	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp LIMIT $1 OFFSET $2`, goodsTable)
	_, err := r.db.Prepare(r.ctx, "getAllGoods", query)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	rows, err := r.db.Query(r.ctx, "getAllGoods", limit, offset)
	if err != nil {
		return models.GetAllGoods{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var good models.Goods
		if err := rows.Scan(&good.ID, &good.ProjectID, &good.Name, &good.Description, &good.Priority, &good.Removed, &good.CreatedAt); err != nil {
			return models.GetAllGoods{}, err
		}
		goods = append(goods, good)
	}

	if err := rows.Err(); err != nil {
		return models.GetAllGoods{}, err
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(gp.id) FROM %s gp`, goodsTable)
	_, err = r.db.Prepare(r.ctx, "countAllGoods", countQuery)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	err = r.db.QueryRow(r.ctx, "countAllGoods").Scan(&total)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	var removed int
	removedQuery := fmt.Sprintf(`SELECT COUNT(gp.id) FROM %s gp WHERE gp.removed = true`, goodsTable)
	_, err = r.db.Prepare(r.ctx, "countRemovedGoods", removedQuery)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	err = r.db.QueryRow(r.ctx, "countRemovedGoods").Scan(&removed)
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

	return response, nil
}

// GetOne one item from Goods
func (r *GoodsPostgres) GetOne(goodsID, projectID int) (models.Goods, error) {
	var goods models.Goods
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	cachedGoods, err := r.cache.Get(r.ctx, key)
	if err == nil {
		err := json.Unmarshal([]byte(cachedGoods), &goods)
		if err != nil {
			r.logger.Error("Failed to unmarshal cached goods: %v", zap.Error(err))
		} else {
			return goods, nil
		}
	}

	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp WHERE gp.id = $1 AND gp.project_id = $2`, goodsTable)
	_, err = r.db.Prepare(r.ctx, "getOneItem", query)
	if err != nil {
		return goods, err
	}

	err = r.db.QueryRow(r.ctx, "getOneItem", goodsID, projectID).Scan(&goods.ID, &goods.ProjectID, &goods.Name, &goods.Description, &goods.Priority, &goods.Removed, &goods.CreatedAt)
	if err != nil {
		return goods, err
	}

	goodsJson, err := json.Marshal(goods)
	if err != nil {
		r.logger.Error("Failed to marshal goods: %v", zap.Error(err))
		return goods, err
	}
	err = r.cache.Set(r.ctx, key, string(goodsJson), 1*time.Minute)
	if err != nil {
		r.logger.Error("Failed to cache goods: %v", zap.Error(err))
	}

	return goods, nil
}

// Create method creates a new item of Goods
func (r *GoodsPostgres) Create(projectID int, goods models.Goods) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (project_id, name, description, priority, removed) VALUES ($1, $2, $3, $4, $5) RETURNING id`, goodsTable)
	_, err := r.db.Prepare(r.ctx, "createItem", query)
	if err != nil {
		return 0, err
	}

	err = r.db.QueryRow(r.ctx, "createItem", projectID, goods.Name, goods.Description, goods.Priority, goods.Removed).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Update method updates item of Goods
func (r *GoodsPostgres) Update(goodsID, projectID int, input models.UpdateGoods) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	var exists bool
	_, err = r.db.Prepare(r.ctx, "checkGoodsExists", fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable))
	if err != nil {
		return err
	}
	err = tx.QueryRow(r.ctx, "checkGoodsExists", goodsID, projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.Exec(r.ctx, fmt.Sprintf("SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp WHERE gp.id = $1 AND gp.project_id = $2 FOR UPDATE", goodsTable), goodsID, projectID)
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
	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description = $%d", argID))
		args = append(args, *input.Description)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, goodsTable, setQuery, argID)
	args = append(args, goodsID)
	_, err = tx.Exec(r.ctx, query, args...)
	if err != nil {
		return err
	}

	err = tx.Commit(r.ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

// Delete marks item of Goods as deleted
func (r *GoodsPostgres) Delete(goodsID, projectID int) error {
	query := fmt.Sprintf(`UPDATE %s SET removed = true WHERE id = $1 AND project_id = $2`, goodsTable)
	commandTag, err := r.db.Exec(r.ctx, query, goodsID, projectID)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

// Reprioritize method changes priority of item of Goods
func (r *GoodsPostgres) Reprioritize(goodsID, projectID int, priority int) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx) // Убедитесь, что вызывается Rollback в случае ошибки

	var exists bool
	_, err = r.db.Prepare(r.ctx, "checkGoodsExists", fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable))
	if err != nil {
		return err
	}
	err = tx.QueryRow(r.ctx, "checkGoodsExists", goodsID, projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound // Предполагается, что ErrNotFound определен где-то в вашем коде
	}

	query := fmt.Sprintf(`UPDATE %s SET priority = $1 WHERE id = $2 AND project_id = $3`, goodsTable)
	_, err = tx.Exec(r.ctx, query, priority, goodsID, projectID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`UPDATE %s SET priority = priority + 1 WHERE project_id = $1 AND priority >= $2`, goodsTable)
	_, err = tx.Exec(r.ctx, query, projectID, priority)
	if err != nil {
		return err
	}

	err = tx.Commit(r.ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}
