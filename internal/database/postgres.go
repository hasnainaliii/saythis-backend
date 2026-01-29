package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Querier defines the interface for database operations,
// satisfied by both *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func Connect(databaseURL string) (*pgxpool.Pool, error) {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		zap.S().Errorw("Unable to parse database URL", "error", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		zap.S().Errorw("Unable to create connection pool", "error", err)
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		zap.S().Errorw("Unable to ping database", "error", err)
		pool.Close()
		return nil, err
	}

	zap.S().Info("✅ Database connected successfully")
	return pool, nil
}
