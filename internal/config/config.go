package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	AppEnv      string
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		AppEnv:      os.Getenv("APP_ENV"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	if cfg.Port == "" {
		cfg.Port = ":8080"
	}
	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	return cfg, nil
}
