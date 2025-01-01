package models

import (
	"errors"
	"time"

	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/appmetric"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/clarity"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/embrace"
	"github.com/hossein-nas/analytics_aggregator/internal/project/collector/sentry"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Stat represents statistics for a service
type Stat struct {
	LastCollectedAt time.Time `json:"last_collected_at,omitempty" bson:"last_collected_at,omitempty"`
	LastStatus      string    `json:"last_status,omitempty" bson:"last_status,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty" bson:"error_message,omitempty"`
	CollectionCount int64     `json:"collection_count,omitempty" bson:"collection_count,omitempty"`
}

type Project struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	Key        string             `json:"key" bson:"key"`
	CreatedBy  string             `json:"created_by" bson:"created_by"`
	Active     bool               `json:"active" bson:"active"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	Collectors []string           `json:"collectors" bson:"collectors"` // e.g., ["sentry", "clarity"]

	// Service configs - only populated if service is enabled
	SentryConfig    *sentry.Config    `json:"sentry_config,omitempty" bson:"sentry_config,omitempty"`
	ClarityConfig   *clarity.Config   `json:"clarity_config,omitempty" bson:"clarity_config,omitempty"`
	EmbraceConfig   *embrace.Config   `json:"embrace_config,omitempty" bson:"embrace_config,omitempty"`
	AppMetricConfig *appmetric.Config `json:"app_metric_config,omitempty" bson:"app_metric_config,omitempty"`

	// Stats for each service
	SentryStats    *Stat `json:"sentry_stats,omitempty" bson:"sentry_stats,omitempty"`
	ClarityStats   *Stat `json:"clarity_stats,omitempty" bson:"clarity_stats,omitempty"`
	EmbraceStats   *Stat `json:"embrace_stats,omitempty" bson:"embrace_stats,omitempty"`
	AppMetricStats *Stat `json:"app_metric_stats,omitempty" bson:"app_metric_stats,omitempty"`
}

var (
	ErrProjectNotFound = errors.New("project not found")
)

type CreateProjectInput struct {
	Name       string   `json:"name" validate:"required" bson:"name"`
	Key        string   `json:"key" validate:"required,alphanum" bson:"key"`
	Collectors []string `json:"collectors" validate:"required" bson:"collectors"`

	SentryConfig    *sentry.Config    `json:"sentry_config,omitempty" bson:"sentry_config,omitempty"`
	ClarityConfig   *clarity.Config   `json:"clarity_config,omitempty" bson:"clarity,omitempty"`
	EmbraceConfig   *embrace.Config   `json:"embrace_config,omitempty" bson:"embrace_config,omitempty"`
	AppMetricConfig *appmetric.Config `json:"app_metric_config,omitempty" bson:"app_metric_config,omitempty"`
}

type UpdateProjectInput struct {
	Name       string   `json:"name,omitempty"`
	Collectors []string `json:"collectors,omitempty"`
	Active     *bool    `json:"active,omitempty"`

	SentryConfig    *sentry.Config    `json:"sentry_config,omitempty" bson:"sentry_config,omitempty"`
	ClarityConfig   *clarity.Config   `json:"clarity_config,omitempty" bson:"clarity,omitempty"`
	EmbraceConfig   *embrace.Config   `json:"embrace_config,omitempty" bson:"embrace_config,omitempty"`
	AppMetricConfig *appmetric.Config `json:"app_metric_config,omitempty" bson:"app_metric_config,omitempty"`
}
