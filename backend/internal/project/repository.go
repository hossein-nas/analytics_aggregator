package project

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrDuplicateKey    = errors.New("project key already exists")
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	GetByKey(ctx context.Context, key string) (*Project, error)
	Update(ctx context.Context, project *Project) error
	List(ctx context.Context, userID string) ([]Project, error)
}

type mongoRepository struct {
	db *mongo.Collection
}

func NewMongoRepository(db *mongo.Collection) Repository {
	return &mongoRepository{db: db}
}

func (r *mongoRepository) Create(ctx context.Context, project *Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	_, err := r.db.InsertOne(ctx, project)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	return err
}

func (r *mongoRepository) GetByKey(ctx context.Context, key string) (*Project, error) {
	var project Project
	err := r.db.FindOne(ctx, bson.M{"key": key}).Decode(&project)
	if err == mongo.ErrNoDocuments {
		return nil, ErrProjectNotFound
	}
	return &project, err
}

func (r *mongoRepository) Update(ctx context.Context, project *Project) error {
	project.UpdatedAt = time.Now()

	result, err := r.db.UpdateOne(
		ctx,
		bson.M{"_id": project.ID},
		bson.M{"$set": project},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *mongoRepository) List(ctx context.Context, userID string) ([]Project, error) {
	cursor, err := r.db.Find(ctx, bson.M{"created_by": userID})
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err = cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}
