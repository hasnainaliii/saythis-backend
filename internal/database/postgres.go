package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectWithRetry(databaseURL string) (*pgxpool.Pool, error) {
	const maxAttempts = 10
	const baseDelay = 2 * time.Second

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pool, connectErr := connect(databaseURL)
		if connectErr == nil {
			return pool, nil
		}
		err = connectErr

		if attempt == maxAttempts {
			break
		}

		delay := time.Duration(attempt) * baseDelay
		if delay > 15*time.Second {
			delay = 15 * time.Second
		}

		slog.Warn("Database connection failed, retrying...",
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"retry_in", delay,
			"error", err,
		)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)
}

func connect(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	slog.Info("✅ Successfully connected to the database",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
	)
	return pool, nil
}
