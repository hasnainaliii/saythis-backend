package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DataBaseUrl string
	Port        string
}

func Load() (*Config, error) {
	var err error = godotenv.Load()

	if err != nil {
		log.Println("Warning : .envv file not found")
	}

	return &Config{
		DataBaseUrl: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}, nil
}
