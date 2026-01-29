package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DataBaseURL string
	Port        string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		zap.S().Warn("⚠️  .env file not found, using environment variables")
	}

	cfg := &Config{
		DataBaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.DataBaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = ":8080"
		zap.S().Info("Using default port :8080")
	}

	if !strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = ":" + cfg.Port
	}

	zap.S().Infow("Configuration loaded", "port", cfg.Port)
	return cfg, nil
}
