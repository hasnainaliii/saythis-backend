package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DataBaseURL        string
	Port               string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	ResendAPIKey       string
	AppBaseURL         string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		zap.S().Warn("⚠️  .env file not found, using environment variables")
	}

	cfg := &Config{
		DataBaseURL:        os.Getenv("DATABASE_URL"),
		Port:               os.Getenv("PORT"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTokenExpiry:  parseDuration(os.Getenv("ACCESS_TOKEN_EXPIRY"), 15*time.Minute),
		RefreshTokenExpiry: parseDuration(os.Getenv("REFRESH_TOKEN_EXPIRY"), 7*24*time.Hour),
		ResendAPIKey:       os.Getenv("FORGOTPASSWORD_APIKEY"),
		AppBaseURL:         os.Getenv("APP_BASE_URL"),
	}

	if cfg.DataBaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	if cfg.Port == "" {
		cfg.Port = ":8080"
		zap.S().Info("Using default port :8080")
	}

	if !strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = ":" + cfg.Port
	}

	// Set default base URL for testing
	if cfg.AppBaseURL == "" {
		cfg.AppBaseURL = "http://localhost" + cfg.Port
	}

	zap.S().Infow("Configuration loaded", "port", cfg.Port)
	return cfg, nil
}

func parseDuration(value string, defaultVal time.Duration) time.Duration {
	if value == "" {
		return defaultVal
	}
	minutes, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return time.Duration(minutes) * time.Minute
}
