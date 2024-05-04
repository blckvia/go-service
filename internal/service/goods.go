package service

import (
	"context"

	"go-service/internal/models"
	"go-service/internal/repository"
)

type GoodsService struct {
	repo repository.Goods
}

func NewGoodsService(repo repository.Goods) *GoodsService {
	return &GoodsService{repo: repo}
}

func (s *GoodsService) GetAll(ctx context.Context, limit, offset int) (models.GetAllGoods, error) {
	return s.repo.GetAll(ctx, limit, offset)
}
func (s *GoodsService) GetOne(ctx context.Context, goodsID, projectID int) (models.Goods, error) {
	return s.repo.GetOne(ctx, goodsID, projectID)
}
func (s *GoodsService) Create(ctx context.Context, projectID int, goods models.Goods) (int, error) {
	return s.repo.Create(ctx, projectID, goods)
}
func (s *GoodsService) Update(ctx context.Context, goodsID, projectID int, input models.UpdateGoods) error {
	return s.repo.Update(ctx, goodsID, projectID, input)
}
func (s *GoodsService) Delete(ctx context.Context, goodsID, projectID int) error {
	return s.repo.Delete(ctx, goodsID, projectID)
}
func (s *GoodsService) Reprioritize(ctx context.Context, goodsID, projectID int, priority int) error {
	return s.repo.Reprioritize(ctx, goodsID, projectID, priority)
}
