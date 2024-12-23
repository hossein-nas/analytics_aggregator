package project

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/hossein-nas/analytics_aggregator/internal/project/scheduler"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProjectSetup(r *mux.Router, collection *mongo.Collection) (*scheduler.Scheduler, error) {
	schedulerConfig, err := scheduler.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load scheduler config: %v", err)
	}
	// Initialize repositories
	projectRepo := NewMongoRepository(collection)

	// Initialize services
	projectService := NewService(projectRepo)

	// Initialize handlers
	projectHandler := NewHandler(projectService)

	// Register routes
	RegisterRoutes(r, projectHandler)

	// Initialize and return scheduler

	projectScheduler := scheduler.NewScheduler(schedulerConfig, projectRepo)
	return projectScheduler, nil
}
