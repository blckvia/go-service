package service

import (
	"go-service/internal/models"
	"go-service/internal/repository"
)

type ProjectService struct {
	repo repository.Projects
}

func NewProjectService(repo repository.Projects) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(input models.Project) (int, error) {
	return s.repo.Create(input)
}
func (s *ProjectService) Update(projectID int, input models.Project) error {
	return s.repo.Update(projectID, input)
}
func (s *ProjectService) Delete(projectID int) error {
	return s.repo.Delete(projectID)
}
func (s *ProjectService) GetAll() ([]models.Project, error) {
	return s.repo.GetAll()
}
func (s *ProjectService) GetByID(projectID int) (models.Project, error) {
	return s.repo.GetByID(projectID)
}
