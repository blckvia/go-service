package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go-service/internal/models"
	p "go-service/pkg/prometheus"

	//p "go-service/pkg/prometheus"
	r "go-service/pkg/redis"
)

var ErrNotFound = errors.New("record not found")

type GoodsPostgres struct {
	ctx    context.Context
	db     *pgxpool.Pool
	cache  r.Cache
	logger *zap.Logger
	nats   *nats.Conn
	tracer trace.Tracer
}

func NewGoodsPostgres(ctx context.Context, db *pgxpool.Pool, cache r.Cache, logger *zap.Logger, nats *nats.Conn, tracer trace.Tracer) *GoodsPostgres {
	return &GoodsPostgres{
		ctx:    ctx,
		db:     db,
		cache:  cache,
		logger: logger,
		nats:   nats,
		tracer: tracer,
	}
}

// GetAll get all Goods
func (r *GoodsPostgres) GetAll(ctx context.Context, limit, offset int) (models.GetAllGoods, error) {
	var goods []models.Goods

	_, span := r.tracer.Start(ctx, "GetAllGoods")
	defer span.End()

	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp LIMIT $1 OFFSET $2`, goodsTable)

	span.AddEvent("getAll", trace.WithAttributes(attribute.String("query", query)))
	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return models.GetAllGoods{}, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "getAllGoods", query)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	rows, err := pgxConn.Query(r.ctx, "getAllGoods", limit, offset)
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
	_, err = pgxConn.Prepare(r.ctx, "countAllGoods", countQuery)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	err = pgxConn.QueryRow(r.ctx, "countAllGoods").Scan(&total)
	if err != nil {
		return models.GetAllGoods{}, err
	}
	span.AddEvent("countAll", trace.WithAttributes(attribute.Int("total", total)))

	var removed int
	removedQuery := fmt.Sprintf(`SELECT COUNT(gp.id) FROM %s gp WHERE gp.removed = true`, goodsTable)
	_, err = pgxConn.Prepare(r.ctx, "countRemovedGoods", removedQuery)
	if err != nil {
		return models.GetAllGoods{}, err
	}

	err = pgxConn.QueryRow(r.ctx, "countRemovedGoods").Scan(&removed)
	if err != nil {
		return models.GetAllGoods{}, err
	}
	span.AddEvent("countRemoved", trace.WithAttributes(attribute.Int("removed", removed)))

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
func (r *GoodsPostgres) GetOne(ctx context.Context, goodsID, projectID int) (models.Goods, error) {
	var goods models.Goods

	_, span := r.tracer.Start(ctx, "GetOneItem")
	defer span.End()

	span.AddEvent("redis get", trace.WithAttributes(attribute.String("key", fmt.Sprintf("goods:%d:%d", goodsID, projectID))))
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	cachedGoods, err := r.cache.Get(r.ctx, key)
	if err == nil {
		p.CacheHitsTotal.WithLabelValues(fmt.Sprintf("goods:%d", goodsID), fmt.Sprintf("project:%d", projectID)).Inc()
		err := json.Unmarshal([]byte(cachedGoods), &goods)
		if err != nil {
			r.logger.Error("Failed to unmarshal cached goods: %v", zap.Error(err))
		} else {
			return goods, nil
		}
	} else {
		p.CacheMissesTotal.WithLabelValues(fmt.Sprintf("goods:%d", goodsID), fmt.Sprintf("project:%d", projectID)).Inc()
	}

	query := fmt.Sprintf(`SELECT gp.id, gp.project_id, gp.name, gp.description, gp.priority, gp.removed, gp.created_at FROM %s gp WHERE gp.id = $1 AND gp.project_id = $2`, goodsTable)

	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return goods, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "getOneItem", query)
	if err != nil {
		return goods, err
	}

	err = pgxConn.QueryRow(r.ctx, "getOneItem", goodsID, projectID).Scan(&goods.ID, &goods.ProjectID, &goods.Name, &goods.Description, &goods.Priority, &goods.Removed, &goods.CreatedAt)
	if err != nil {
		return goods, err
	}
	span.AddEvent("write goods in JSON")
	goodsJson, err := json.Marshal(goods)

	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to marshal goods: %v", zap.Error(err))
		return goods, err
	}

	span.AddEvent("redis set", trace.WithAttributes(attribute.String("key", fmt.Sprintf("goods:%d:%d", goodsID, projectID))))
	err = r.cache.Set(r.ctx, key, string(goodsJson), 1*time.Minute)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to cache goods: %v", zap.Error(err))
	}

	return goods, nil
}

// Create method creates a new item of Goods
func (r *GoodsPostgres) Create(ctx context.Context, projectID int, goods models.Goods) (int, error) {
	var id int

	_, span := r.tracer.Start(ctx, "CreateItem")
	defer span.End()

	query := fmt.Sprintf(`INSERT INTO %s (project_id, name, description, priority, removed) VALUES ($1, $2, $3, $4, $5) RETURNING id`, goodsTable)

	conn, err := r.db.Acquire(r.ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	pgxConn := conn.Conn()

	_, err = pgxConn.Prepare(r.ctx, "createItem", query)
	if err != nil {
		return 0, err
	}

	span.AddEvent("create item", trace.WithAttributes(attribute.String("query", query)))
	err = pgxConn.QueryRow(r.ctx, "createItem", projectID, goods.Name, goods.Description, goods.Priority, goods.Removed).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Update method updates item of Goods
func (r *GoodsPostgres) Update(ctx context.Context, goodsID, projectID int, input models.UpdateGoods) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	_, span := r.tracer.Start(ctx, "UpdateItem")
	defer span.End()

	var exists bool
	err = tx.QueryRow(r.ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable), goodsID, projectID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		return ErrNotFound
	}
	span.AddEvent("update item", trace.WithAttributes(attribute.Int("goodsID", goodsID), attribute.Int("projectID", projectID), attribute.Bool("exists", exists)))

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

	span.AddEvent("set query", trace.WithAttributes(attribute.String("setQuery", setQuery)))
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

	span.AddEvent("invalidate goods in cache", trace.WithAttributes(attribute.String("key", fmt.Sprintf("goods:%d:%d", goodsID, projectID))))
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("Invalidate error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

// Delete marks item of Goods as deleted
func (r *GoodsPostgres) Delete(ctx context.Context, goodsID, projectID int) error {
	_, span := r.tracer.Start(ctx, "DeleteItem")
	defer span.End()

	query := fmt.Sprintf(`UPDATE %s SET removed = true WHERE id = $1 AND project_id = $2`, goodsTable)
	span.AddEvent("delete item", trace.WithAttributes(attribute.String("query", query)))
	commandTag, err := r.db.Exec(r.ctx, query, goodsID, projectID)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	rowsAffected := commandTag.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	span.AddEvent("invalidate goods in cache", trace.WithAttributes(attribute.String("key", fmt.Sprintf("goods:%d:%d", goodsID, projectID))))
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}

// Reprioritize method changes priority of item of Goods
func (r *GoodsPostgres) Reprioritize(ctx context.Context, goodsID, projectID int, priority int) error {
	tx, err := r.db.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx) // Убедитесь, что вызывается Rollback в случае ошибки

	_, span := r.tracer.Start(ctx, "ReprioritizeItem")
	defer span.End()

	var exists bool
	err = tx.QueryRow(r.ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND project_id = $2)", goodsTable), goodsID, projectID).Scan(&exists)
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

	span.AddEvent("invalidate goods in cache", trace.WithAttributes(attribute.String("key", fmt.Sprintf("goods:%d:%d", goodsID, projectID))))
	key := fmt.Sprintf("goods:%d:%d", goodsID, projectID)
	err = r.cache.Delete(r.ctx, key)
	if err != nil {
		span.RecordError(err, trace.WithAttributes(attribute.String("error", err.Error())))
		span.SetStatus(codes.Error, err.Error())
		r.logger.Error("Failed to invalidate cache for key %s: %v", zap.String("key", key), zap.Error(err))
	}

	return nil
}
