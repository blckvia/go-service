package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

type Projects interface {
	Create(input models.Project) (int, error)
	Update(projectID int, input models.UpdateProjects) error
	Delete(projectID int) error
	GetAll(limit, offset int) (models.GetAllProjects, error)
	GetByID(projectID int) (models.Project, error)
}

type Goods interface {
	Create(projectID int, goods models.Goods) (int, error)
	Update(goodsID, projectID int, input models.UpdateGoods) error
	Delete(goodsID, projectID int) error
	GetAll(limit, offset int) (models.GetAllGoods, error)
	GetOne(goodsID, projectID int) (models.Goods, error)
	Reprioritize(goodsID, projectID int, priority int) error
}

type Repository struct {
	Projects
	Goods
}

func New(ctx context.Context, db *pgx.Conn, cache r.Cache, logger *zap.Logger) *Repository {
	return &Repository{
		Goods:    NewGoodsPostgres(ctx, db, cache, logger),
		Projects: NewProjectPostgres(ctx, db, cache, logger),
	}
}
