// pkg/database/mongodb.go
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	URI            string
	Database       string
	Username       string
	Password       string
	ConnectTimeout time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
}

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewConfig returns a default configuration
func NewConfig(url, db string) *Config {
	return &Config{
		URI:            url,
		Database:       db,
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    10,
	}
}

// Connect establishes a connection to MongoDB with the given configuration
func Connect(cfg *Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	// Add credentials if provided
	if cfg.Username != "" && cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.Database),
	}, nil
}

// Close disconnects from MongoDB
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Database returns the mongo.Database instance
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Collection returns a mongo.Collection instance
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// WithTransaction executes the provided function within a transaction
func (m *MongoDB) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}

// Health checks if the database connection is healthy
func (m *MongoDB) Health(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// CreateIndexes creates indexes for the users collection
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Users collection indexes
	usersIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "username", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_username"),
		},
	}

	// Refresh tokens collection indexes
	refreshTokenIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "token", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_user_token"),
		},
		{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(0).
				SetName("ttl_expires_at"),
		},
	}

	// Create indexes for users collection
	_, err := m.Collection("users").Indexes().CreateMany(ctx, usersIndexes)
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %v", err)
	}

	// Create indexes for refresh tokens collection
	_, err = m.Collection("refresh_tokens").Indexes().CreateMany(ctx, refreshTokenIndexes)
	if err != nil {
		return fmt.Errorf("failed to create refresh_tokens indexes: %v", err)
	}

	return nil
}
