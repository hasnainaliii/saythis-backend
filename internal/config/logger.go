package config

import (
	"errors"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

var Logger *slog.Logger

type LoggerConfig struct {
	Env     string
	Service string
	Level   slog.Level
	Output  io.Writer
}

func InitLogger(cfg LoggerConfig) (*slog.Logger, error) {

	if cfg.Output == nil {
		return nil, errors.New("logger output cannot be nil")
	}

	var handler slog.Handler

	// Production → JSON logs
	if cfg.Env == "production" {
		handler = slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
			Level:     cfg.Level,
			AddSource: true,
		})
	} else {

		// Development → colored logs
		handler = tint.NewHandler(cfg.Output, &tint.Options{
			Level:      cfg.Level,
			TimeFormat: "15:04:05",
			AddSource:  true,
		})
	}

	logger := slog.New(handler).With(
		"service", cfg.Service,
	)

	Logger = logger
	slog.SetDefault(logger)

	return logger, nil
}
