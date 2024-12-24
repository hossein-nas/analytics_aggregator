package project

import (
	"context"
	"time"

	"github.com/hossein-nas/analytics_aggregator/internal/project/collector"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/sentry"
	model "github.com/hossein-nas/analytics_aggregator/internal/project/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service interface {
	CreateProject(ctx context.Context, userID string, input model.CreateProjectInput) (*model.Project, error)
	UpdateProject(ctx context.Context, key string, input model.UpdateProjectInput) (*model.Project, error)
	GetProject(ctx context.Context, key string) (*model.Project, error)
	ListProjects(ctx context.Context, userID string) ([]model.Project, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateProject(ctx context.Context, userID string, input model.CreateProjectInput) (*model.Project, error) {
	var sentryConfig sentry.Config = sentry.Config{
		BaseCollector: collector.BaseCollector{
			ID:   primitive.NewObjectID(),
			Type: "sentry",
		},
		OrganizationSlug: input.SentryConfig.OrganizationSlug,
		ProjectSlug:      input.SentryConfig.ProjectSlug,
		AuthToken:        input.SentryConfig.AuthToken,
		Host:             input.SentryConfig.Host,
	}

	project := &model.Project{
		ID:              primitive.NewObjectID(),
		Name:            input.Name,
		Key:             input.Key,
		Collectors:      input.Collectors,
		CreatedBy:       userID,
		Active:          true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		SentryConfig:    &sentryConfig,
		ClarityConfig:   input.ClarityConfig,
		EmbraceConfig:   input.EmbraceConfig,
		AppMetricConfig: input.AppMetricConfig,
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *service) UpdateProject(ctx context.Context, key string, input model.UpdateProjectInput) (*model.Project, error) {
	project, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Collectors != nil {
		project.Collectors = input.Collectors
	}
	if input.Active != nil {
		project.Active = *input.Active
	}

	// Update collector configs if provided
	if input.SentryConfig != nil {
		project.SentryConfig = input.SentryConfig
	}
	if input.ClarityConfig != nil {
		project.ClarityConfig = input.ClarityConfig
	}
	if input.EmbraceConfig != nil {
		project.EmbraceConfig = input.EmbraceConfig
	}
	if input.AppMetricConfig != nil {
		project.AppMetricConfig = input.AppMetricConfig
	}

	project.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *service) GetProject(ctx context.Context, key string) (*model.Project, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *service) ListProjects(ctx context.Context, userID string) ([]model.Project, error) {
	return s.repo.List(ctx, userID)
}
