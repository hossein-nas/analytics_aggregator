package scheduler

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SchedulerService interface {
	StoreStats(ctx context.Context, projectId primitive.ObjectID, _type string, payload *Stat) error
	LastStats(ctx context.Context, projectId primitive.ObjectID, _type string) (*Stat, error)
}

type service struct {
	repo Repository
}

func NewSchedulerService(repo Repository) SchedulerService {
	return &service{
		repo: repo,
	}
}

func (s *service) StoreStats(ctx context.Context, projectId primitive.ObjectID, _type string, payload *Stat) error {
	println("Storing stats.")
	return s.repo.StoreStats(ctx, projectId, _type, payload)
}

func (s *service) LastStats(ctx context.Context, projectId primitive.ObjectID, _type string) (*Stat, error) {
	return s.repo.LastStats(ctx, projectId, _type)
}
