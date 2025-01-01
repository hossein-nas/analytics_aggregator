package scheduler

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	StatNotFound  = errors.New("Stat not found.")
	DuplicateStat = errors.New("There is an entry for this same stat.")
)

type Repository interface {
	StoreStats(ctx context.Context, projectId primitive.ObjectID, _type string, payload *Stat) error
	LastStats(ctx context.Context, projectId primitive.ObjectID, _type string) (*Stat, error)
}

type SchedulerRepository struct {
	db *mongo.Collection
}

func NewSchedulerRepository(db *mongo.Collection) Repository {
	return &SchedulerRepository{db: db}
}

func (r *SchedulerRepository) StoreStats(ctx context.Context, projectId primitive.ObjectID, _type string, payload *Stat) error {
	statsKey := _type + "_stats"
	update := bson.M{
		"$set": bson.M{
			statsKey: payload,
		},
	}

	result, err := r.db.UpdateOne(ctx, bson.M{"_id": projectId}, update)
	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	if result.MatchedCount == 0 {
		return StatNotFound
	}

	return nil
}

func (r *SchedulerRepository) LastStats(ctx context.Context, projectId primitive.ObjectID, _type string) (*Stat, error) {
	var project struct {
		SentryStats    *Stat `bson:"sentry_stats,omitempty"`
		ClarityStats   *Stat `bson:"clarity_stats,omitempty"`
		EmbraceStats   *Stat `bson:"embrace_stats,omitempty"`
		AppMetricStats *Stat `bson:"app_metric_stats,omitempty"`
	}

	err := r.db.FindOne(ctx, bson.M{"_id": projectId}).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, StatNotFound
		}
		return nil, err
	}

	switch _type {
	case "sentry":
		return project.SentryStats, nil
	case "clarity":
		return project.ClarityStats, nil
	case "embrace":
		return project.EmbraceStats, nil
	case "appmetric":
		return project.AppMetricStats, nil
	default:
		return nil, fmt.Errorf("unknown collector type: %s", _type)
	}
}
