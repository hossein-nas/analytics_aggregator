package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDB MongoDBConfig
	JWT     JWTConfig
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// DB connection string
	// for localhost mongoDB
	// const connectionString = "mongodb://localhost:27017"
	var connectionString = os.Getenv("DB_URL")

	// Database Name
	var dbName = os.Getenv("DB_NAME")
	// Load configuration from environment or file
	return &Config{
		MongoDB: MongoDBConfig{
			URI:      connectionString,
			Database: dbName,
		},
		JWT: JWTConfig{
			AccessSecret:  os.Getenv("ACCESS_SECRET"),
			RefreshSecret: os.Getenv("REFRESH_SECRET"),
		},
	}, nil
}
