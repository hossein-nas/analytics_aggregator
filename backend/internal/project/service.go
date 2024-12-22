package project

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	CreateProject(ctx context.Context, userID string, input CreateProjectInput) (*Project, error)
	UpdateProject(ctx context.Context, key string, input UpdateProjectInput) (*Project, error)
	GetProject(ctx context.Context, key string) (*Project, error)
	ListProjects(ctx context.Context, userID string) ([]Project, error)
}

type CreateProjectInput struct {
	Name     string          `json:"name" validate:"required"`
	Key      string          `json:"key" validate:"required,alphanum"`
	Services []ServiceConfig `json:"services" validate:"required,dive"`
}

type UpdateProjectInput struct {
	Name     string          `json:"name,omitempty"`
	Services []ServiceConfig `json:"services,omitempty"`
	Active   *bool           `json:"active,omitempty"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateProject(ctx context.Context, userID string, input CreateProjectInput) (*Project, error) {
	project := &Project{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Key:       input.Key,
		Services:  input.Services,
		CreatedBy: userID,
		Active:    true,
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *service) UpdateProject(ctx context.Context, key string, input UpdateProjectInput) (*Project, error) {
	project, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Services != nil {
		project.Services = input.Services
	}
	if input.Active != nil {
		project.Active = *input.Active
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *service) GetProject(ctx context.Context, key string) (*Project, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *service) ListProjects(ctx context.Context, userID string) ([]Project, error) {
	return s.repo.List(ctx, userID)
}
