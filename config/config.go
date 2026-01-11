package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
}

// Load loads the configuration from environment variables.
// It attempts to load from .env file first, then falls back to system environment variables.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		Port:        os.Getenv("PORT"),
	}

	// Set default port if not specified
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Set default Redis URL if not specified
	if cfg.RedisURL == "" {
		cfg.RedisURL = "redis://localhost:6379/0"
	}

	return cfg
}
