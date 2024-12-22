package scheduler

import (
	"fmt"

	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/appmetric"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/clarity"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/embrace"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/sentry"
	model "github.com/hossein-nas/analytics_aggregator/internal/project/models"
)

func createSentryCollector(project model.Project) (*sentry.Collector, error) {
	if project.SentryConfig == nil {
		return nil, fmt.Errorf("sentry configuration is missing")
	}

	collector := sentry.NewCollector(sentry.Config{
		OrganizationSlug: project.SentryConfig.ProjectID,
		AuthToken:        project.SentryConfig.APIKey,
	})
	return collector, nil
}

func createClarityCollector(project model.Project) (*clarity.Collector, error) {
	return clarity.NewCollector(clarity.Config{
		ProjectID: project.ClarityConfig.ProjectID,
		APIKey:    project.ClarityConfig.APIKey,
	})
}

func createEmbraceCollector(project model.Project) (*embrace.Collector, error) {
	if project.EmbraceConfig == nil {
		return nil, fmt.Errorf("embrace configuration is missing")
	}

	collector := embrace.NewCollector(embrace.Config{
		AppID:  project.EmbraceConfig.AppID,
		APIKey: project.EmbraceConfig.APIKey,
	})
	return collector, nil
}

func createAppMetricCollector(project model.Project) (*appmetric.Collector, error) {
	if project.AppMetricConfig == nil {
		return nil, fmt.Errorf("app metric configuration is missing")
	}

	collector := appmetric.NewCollector(appmetric.Config{
		ApplicationID: project.AppMetricConfig.ApplicationID,
		APIKey:        project.AppMetricConfig.APIKey,
	})
	return collector, nil
}
