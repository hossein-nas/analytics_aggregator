package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectorType string

const (
	SentryCollector     CollectorType = "sentry"
	ClarityCollector    CollectorType = "clarity"
	EmbraceCollector    CollectorType = "embrace"
	AppMetricaCollector CollectorType = "appmetrica"
)

type CollectorStatus string

const (
	StatusActive   CollectorStatus = "active"
	StatusInactive CollectorStatus = "inactive"
	StatusError    CollectorStatus = "error"
)

type BaseCollector struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ProjectID primitive.ObjectID `bson:"project_id"`
	Type      CollectorType      `bson:"type"`
	Status    CollectorStatus    `bson:"status"`
	LastRun   time.Time          `bson:"last_run"`
	Error     string             `bson:"error,omitempty"`
}

type Collector interface {
	Collect(ctx context.Context) error
	Validate() error
	GetMetrics() (map[string]interface{}, error)
}
