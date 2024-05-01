package service

import (
	"context"

	"go-service/internal/models"
	"go-service/internal/repository"
)

type ProjectService struct {
	repo repository.Projects
}

func NewProjectService(repo repository.Projects) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, project models.Project) (int, error) {
	return s.repo.Create(ctx, project)
}
func (s *ProjectService) Update(ctx context.Context, projectID int, input models.UpdateProjects) error {
	return s.repo.Update(ctx, projectID, input)
}
func (s *ProjectService) Delete(ctx context.Context, projectID int) error {
	return s.repo.Delete(ctx, projectID)
}
func (s *ProjectService) GetAll(ctx context.Context, limit, offset int) (models.GetAllProjects, error) {
	return s.repo.GetAll(ctx, limit, offset)
}
func (s *ProjectService) GetByID(ctx context.Context, projectID int) (models.Project, error) {
	return s.repo.GetByID(ctx, projectID)
}
