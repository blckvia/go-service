package service

import (
	"go-service/internal/models"
	"go-service/internal/repository"
)

type Projects interface {
	Create(input models.Project) (int, error)
	Update(projectID int, input models.Project) error
	Delete(projectID int) error
	GetAll() ([]models.Project, error)
	GetByID(projectID int) (models.Project, error)
}

type Goods interface {
	Create(projectID int, goods models.Goods) (int, error)
	Update(goodsID, projectID int, input models.UpdateGoods) error
	Delete(goodsID, projectID int) error
	GetAll(limit, offset int) (models.GetAllGoodsResponse, error)
	GetOne(goodsID, projectID int) (models.Goods, error)
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
