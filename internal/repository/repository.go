package repository

import (
	"github.com/jmoiron/sqlx"

	"go-service/internal/models"
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
	Update(goodsID int, input models.Goods) error
	Delete(goodsID int) error
	GetAll() ([]models.Goods, error)
	GetByID(goodsID int) (models.Goods, error)
}

type Repository struct {
	Projects
	Goods
}

func New(db *sqlx.DB) *Repository {
	return &Repository{}
}
