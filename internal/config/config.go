package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPPort    string
}

func Load() (*Config, error) {
	godotenv.Load()
	// 1. Read DATABASE_URL (must not be empty)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return &Config{}, errors.New("DB_URL must be set")
	}
	// 2. Read HTTP_PORT (default "8080" if empty)
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	// 3. Return &Config{...}
	return &Config{DatabaseURL: dbURL, HTTPPort: httpPort}, nil
}
