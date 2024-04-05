package service

import (
	"go-service/internal/models"
	"go-service/internal/repository"
)

type GoodsService struct {
	repo repository.Goods
}

func NewGoodsService(repo repository.Goods) *GoodsService {
	return &GoodsService{repo: repo}
}

func (s *GoodsService) GetAll() ([]models.Goods, error) {
	return s.repo.GetAll()
}
func (s *GoodsService) GetByID(goodsID int) (models.Goods, error) {
	return s.repo.GetByID(goodsID)
}
func (s *GoodsService) Create(projectID int, goods models.Goods) (int, error) {
	return s.repo.Create(projectID, goods)
}
func (s *GoodsService) Update(goodsID int, input models.Goods) error {
	return s.repo.Update(goodsID, input)
}
func (s *GoodsService) Delete(goodsID int) error {
	return s.repo.Delete(goodsID)
}
