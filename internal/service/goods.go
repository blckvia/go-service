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

func (s *GoodsService) GetAll(limit, offset int) (models.GetAllGoods, error) {
	return s.repo.GetAll(limit, offset)
}
func (s *GoodsService) GetOne(goodsID, projectID int) (models.Goods, error) {
	return s.repo.GetOne(goodsID, projectID)
}
func (s *GoodsService) Create(projectID int, goods models.Goods) (int, error) {
	return s.repo.Create(projectID, goods)
}
func (s *GoodsService) Update(goodsID, projectID int, input models.UpdateGoods) error {
	return s.repo.Update(goodsID, projectID, input)
}
func (s *GoodsService) Delete(goodsID, projectID int) error {
	return s.repo.Delete(goodsID, projectID)
}

func (s *GoodsService) Reprioritize(goodsID, projectID int, priority int) error {
	return s.repo.Reprioritize(goodsID, projectID, priority)
}
