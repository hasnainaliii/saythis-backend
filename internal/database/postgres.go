package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(database string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(database)
	if err != nil {
		slog.Error("Unable to parse the database url :", "error", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Unable to create a connection pool :", "error", err)
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err = pool.Ping(pingCtx); err != nil {
		slog.Error("Unable to ping the database :", "error", err)
		pool.Close()
		return nil, err
	}

	slog.Info("✅ Successfully connected to the database ")
	return pool, nil
}
