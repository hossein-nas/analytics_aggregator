package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
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
