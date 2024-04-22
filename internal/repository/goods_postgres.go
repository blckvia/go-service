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

// TODO: rewrite with r.db.Preparex()
// TODO: QueryBuilder (squirell package)
func (r *GoodsPostgres) GetAll(ctx context.Context, limit, offset int) (models.GetAllGoods, error) {
	var goods []models.Goods
	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp LIMIT $1 OFFSET $2`, goodsTable)
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	err = r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(gp.id) FROM %s gp`, goodsTable)).Scan(&total)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	var removed int
	err = r.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(gp.id) FROM %s gp WHERE gp.removed = true`, goodsTable)).Scan(&removed)
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

func (r *GoodsPostgres) GetOne(ctx context.Context, goodsID, projectID int) (models.Goods, error) {
	var goods models.Goods
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	cachedGoods, err := r.cache.Get(ctx, key)
	if err == nil {
		err := json.Unmarshal([]byte(cachedGoods), &goods)
		if err != nil {
			r.logger.Error("Failed to unmarshal cached goods: %v", zap.Error(err))
		} else {
			return goods, nil
		}
	}

	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp WHERE gp.id = $1 AND gp.project_id = $2`, goodsTable)
	err = r.db.QueryRow(ctx, query, goodsID, projectID).Scan(&goods.ID, &goods.ProjectID, &goods.Name, &goods.Description, &goods.Priority, &goods.Removed, &goods.CreatedAt)
	if err != nil {
		return goods, err
	}

	goodsJson, err := json.Marshal(goods)
	if err != nil {
		r.logger.Error("Failed to marshal goods: %v", zap.Error(err))
		return goods, err
	}
	err = r.cache.Set(ctx, key, string(goodsJson), 1*time.Minute)
	if err != nil {
		r.logger.Error("Failed to cache goods: %v", zap.Error(err))
	}

	return goods, nil
}
func (r *GoodsPostgres) Create(ctx context.Context, projectID int, goods models.Goods) (int, error) {
	var id int
	query := fmt.Sprintf(`INSERT INTO %s (project_id, name, description, priority, removed) VALUES ($1, $2, $3, $4, $5) RETURNING id`, goodsTable)
	err := r.db.QueryRow(ctx, query, projectID, goods.Name, goods.Description, goods.Priority, goods.Removed).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *GoodsPostgres) Update(ctx context.Context, goodsID, projectID int, input models.UpdateGoods) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable), goodsID, projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	_, err = tx.Exec(ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND project_id = $2 FOR UPDATE", goodsTable), goodsID, projectID)
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
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

func (r *GoodsPostgres) Delete(ctx context.Context, goodsID, projectID int) error {
	query := fmt.Sprintf(`UPDATE %s SET removed = true WHERE id = $1 AND project_id = $2`, goodsTable)
	commandTag, err := r.db.Exec(ctx, query, goodsID, projectID)
	if err != nil {
		return err
	}

	rowsAffected := commandTag.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

func (r *GoodsPostgres) Reprioritize(ctx context.Context, goodsID, projectID int, priority int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Убедитесь, что вызывается Rollback в случае ошибки

	var exists bool
	err = tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable), goodsID, projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound // Предполагается, что ErrNotFound определен где-то в вашем коде
	}

	query := fmt.Sprintf(`UPDATE %s SET priority = $1 WHERE id = $2 AND project_id = $3`, goodsTable)
	_, err = tx.Exec(ctx, query, priority, goodsID, projectID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`UPDATE %s SET priority = priority + 1 WHERE project_id = $1 AND priority >= $2`, goodsTable)
	_, err = tx.Exec(ctx, query, projectID, priority)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(ctx, key)
	if err != nil {
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}
