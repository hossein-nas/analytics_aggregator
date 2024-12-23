package model

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Key        string             `bson:"key" json:"key"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
	Active     bool               `bson:"active" json:"active"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	Collectors []string           `bson:"collectors"` // e.g., ["sentry", "clarity"]

	// Service configs - only populated if service is enabled
	SentryConfig    *SentryConfig    `bson:"sentry_config,omitempty"`
	ClarityConfig   *ClarityConfig   `bson:"clarity_config,omitempty"`
	EmbraceConfig   *EmbraceConfig   `bson:"embrace_config,omitempty"`
	AppMetricConfig *AppMetricConfig `bson:"app_metric_config,omitempty"`
}

type SentryConfig struct {
	ProjectID string `bson:"project_id"`
	APIKey    string `bson:"api_key"`
}

type ClarityConfig struct {
	ProjectID string `bson:"project_id"`
	APIKey    string `bson:"api_key"`
}

type EmbraceConfig struct {
	AppID  string `bson:"app_id"`
	APIKey string `bson:"api_key"`
}

type AppMetricConfig struct {
	ApplicationID string `bson:"application_id"`
	APIKey        string `bson:"api_key"`
}

var (
	ErrProjectNotFound = errors.New("project not found")
)

type CreateProjectInput struct {
	Name       string   `json:"name" validate:"required"`
	Key        string   `json:"key" validate:"required,alphanum"`
	Collectors []string `json:"collectors" validate:"required"`

	SentryConfig    *SentryConfig    `json:"sentry_config,omitempty"`
	ClarityConfig   *ClarityConfig   `json:"clarity_config,omitempty"`
	EmbraceConfig   *EmbraceConfig   `json:"embrace_config,omitempty"`
	AppMetricConfig *AppMetricConfig `json:"app_metric_config,omitempty"`
}

type UpdateProjectInput struct {
	Name       string   `json:"name,omitempty"`
	Collectors []string `json:"collectors,omitempty"`
	Active     *bool    `json:"active,omitempty"`

	SentryConfig    *SentryConfig    `json:"sentry_config,omitempty"`
	ClarityConfig   *ClarityConfig   `json:"clarity_config,omitempty"`
	EmbraceConfig   *EmbraceConfig   `json:"embrace_config,omitempty"`
	AppMetricConfig *AppMetricConfig `json:"app_metric_config,omitempty"`
}
