package repository

import (
	"context"

	model "github.com/hossein-nas/analytics_aggregator/internal/project/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectRepository struct {
	collection *mongo.Collection
}

func NewProjectRepository(collection *mongo.Collection) *ProjectRepository {
	return &ProjectRepository{collection: collection}
}

func (r *ProjectRepository) GetAllProjects() ([]model.Project, error) {
	var projects []model.Project
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}
	return projects, nil
}
