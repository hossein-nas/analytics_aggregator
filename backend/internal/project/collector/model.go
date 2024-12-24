package collector

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
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Type    CollectorType      `json:"type" bson:"type"`
	Status  CollectorStatus    `json:"status" bson:"status"`
	LastRun time.Time          `json:"last_run" bson:"last_run"`
	Error   string             `json:"error,omitempty" bson:"error,omitempty"`
}

type Collector interface {
	Collect(ctx context.Context) error
	Validate() error
	GetMetrics() (map[string]interface{}, error)
}
