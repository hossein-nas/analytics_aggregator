package project

import (
	"context"
	"time"

	"github.com/hossein-nas/analytics_aggregator/internal/project/collector"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/appmetric"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/clarity"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/embrace"
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
	project := &model.Project{
		ID:         primitive.NewObjectID(),
		Name:       input.Name,
		Key:        input.Key,
		Collectors: input.Collectors,
		CreatedBy:  userID,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Configure Sentry if provided
	if input.SentryConfig != nil {
		sentryConfig := sentry.Config{
			BaseCollector: collector.BaseCollector{
				ID:   primitive.NewObjectID(),
				Type: "sentry",
			},
			OrganizationSlug: input.SentryConfig.OrganizationSlug,
			ProjectSlug:      input.SentryConfig.ProjectSlug,
			AuthToken:        input.SentryConfig.AuthToken,
			Host:             input.SentryConfig.Host,
		}
		project.SentryConfig = &sentryConfig
	}

	// Configure Clarity if provided
	if input.ClarityConfig != nil {
		clarityConfig := clarity.Config{
			BaseCollector: collector.BaseCollector{
				ID:   primitive.NewObjectID(),
				Type: "clarity",
			},
			ProjectID: input.ClarityConfig.ProjectID,
			APIKey:    input.ClarityConfig.APIKey,
		}
		project.ClarityConfig = &clarityConfig
	}

	// Configure Embrace if provided
	if input.EmbraceConfig != nil {
		embraceConfig := embrace.Config{
			BaseCollector: collector.BaseCollector{
				ID:   primitive.NewObjectID(),
				Type: "embrace",
			},
			AppID:     input.EmbraceConfig.AppID,
			APIKey:    input.EmbraceConfig.APIKey,
			AppSecret: input.EmbraceConfig.AppSecret,
		}
		project.EmbraceConfig = &embraceConfig
	}

	// Configure AppMetric if provided
	if input.AppMetricConfig != nil {
		appMetricConfig := appmetric.Config{
			BaseCollector: collector.BaseCollector{
				ID:   primitive.NewObjectID(),
				Type: "appmetric",
			},
			APIKey: input.AppMetricConfig.APIKey,
		}
		project.AppMetricConfig = &appMetricConfig
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
