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
	value := map[string]interface{}{}
	value[statsKey] = *payload
	result, err := r.db.UpdateOne(ctx, bson.M{"_id": projectId}, bson.M{"$set": value})
	if err != nil {
		return nil
	}

	if result.MatchedCount == 0 {
		return StatNotFound
	}

	return nil
}

func (r *SchedulerRepository) LastStats(
	ctx context.Context,
	projectId primitive.ObjectID,
	_type string,
) (*Stat, error) {
	var stat map[string]interface{}
	err := r.db.FindOne(ctx, bson.M{"_id": projectId}).Decode(&stat)
	fmt.Println("stat", stat, _type, projectId)
	if err != nil {
		return nil, StatNotFound
	}

	value := stat[_type+"_stats"].(Stat)

	return &value, nil
}
