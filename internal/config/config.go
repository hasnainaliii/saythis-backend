package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DataBaseUrl string
	Port        string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}
	cfg := &Config{
		DataBaseUrl: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.DataBaseUrl == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.Port == "" {
		cfg.Port = ":8080"
	}
	if !strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = ":" + cfg.Port
	}
	return cfg, nil
}
