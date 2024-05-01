package service

import (
	"context"

	"go-service/internal/models"
	"go-service/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Projects interface {
	Create(ctx context.Context, input models.Project) (int, error)
	Update(ctx context.Context, projectID int, project models.UpdateProjects) error
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

type Service struct {
	Projects
	Goods
}

func New(repo *repository.Repository) *Service {
	return &Service{
		Projects: NewProjectService(repo.Projects),
		Goods:    NewGoodsService(repo.Goods),
	}
}
