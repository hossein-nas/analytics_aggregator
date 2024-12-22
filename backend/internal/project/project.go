package project

import (
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProjectSetup(r *mux.Router, collection *mongo.Collection) {
	// Initialize repositories
	projectRepo := NewMongoRepository(collection)

	// Initialize services
	projectService := NewService(projectRepo)

	// Initialize handlers
	projectHandler := NewHandler(projectService)

	// Register routes
	RegisterRoutes(r, projectHandler)
}
