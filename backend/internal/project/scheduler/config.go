package scheduler

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	CollectionInterval time.Duration
	MaxWorkers         int
}

func LoadConfig() (*Config, error) {
	interval, err := strconv.Atoi(getEnvOrDefault("METRICS_COLLECTION_INTERVAL_SECONDS", "300")) // default 5 minutes
	if err != nil {
		return nil, fmt.Errorf("invalid metrics collection interval: %w", err)
	}

	workers, err := strconv.Atoi(getEnvOrDefault("METRICS_MAX_WORKERS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid max workers value: %w", err)
	}

	return &Config{
		CollectionInterval: time.Duration(interval) * time.Second,
		MaxWorkers:         workers,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
