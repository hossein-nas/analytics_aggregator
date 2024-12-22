package repository

import (
	"context"

	model "github.com/hossein-nas/analytics_aggregator/internal/project/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectRepository struct {
	collection *mongo.Collection
}

func NewProjectRepository(collection *mongo.Collection) *ProjectRepository {
	return &ProjectRepository{collection: collection}
}

func (r *ProjectRepository) GetAllProjects(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	cursor, err := r.collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return projects, cursor.All(ctx, &projects)
}
