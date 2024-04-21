package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"go-service/internal/models"
	r "go-service/pkg/redis"
)

type Projects interface {
	Create(ctx context.Context, input models.Project) (int, error)
	Update(ctx context.Context, projectID int, input models.UpdateProjects) error
	Delete(ctx context.Context, projectID int) error
	GetAll(ctx context.Context, limit, offset int) (models.GetAllProjects, error)
	GetByID(ctx context.Context, projectID int) (models.Project, error)
}

type Goods interface {
	Create(ctx context.Context, projectID int, goods models.Goods) (int, error)
	Update(ctx context.Context, goodsID, projectID int, input models.UpdateGoods) error
	Delete(ctx context.Context, goodsID, projectID int) error
	GetAll(ctx context.Context, limit, offset int) (models.GetAllGoods, error)
	GetOne(ctx context.Context, goodsID, projectID int) (models.Goods, error)
	Reprioritize(ctx context.Context, goodsID, projectID int, priority int) error
}

type Repository struct {
	Projects
	Goods
}

func New(ctx context.Context, db *pgx.Conn, cache r.Cache) *Repository {
	return &Repository{
		Goods: NewGoodsPostgres(ctx, db, cache),
		//Projects: NewProjectPostgres(ctx, db),
	}
}
