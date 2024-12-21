package main

import (
	"context"
	_ "context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hossein-nas/analytics_aggregator/internal/auth"
	"github.com/hossein-nas/analytics_aggregator/internal/config"
	"github.com/hossein-nas/analytics_aggregator/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	// Load configuration
	dbConfig := database.NewConfig(cfg.MongoDB.URI, cfg.MongoDB.Database)

	// Connect to MongoDB
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.CreateIndexes(ctx); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	publicRouter := mux.NewRouter().PathPrefix("/api/").Subrouter()

	// Auth setup
	authHandler := auth.NewHandler(db.Database(), cfg.JWT)
	protectedRouter := auth.RegisterRoutes(publicRouter, authHandler)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := db.Close(shutdownCtx); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}

		os.Exit(0)
	}()

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", publicRouter))
}
