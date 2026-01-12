package repository

import (
	"context"
	"fmt"
	"log"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuthRepository struct {
	db     database.Querier
	logger *log.Logger
}

func NewPostgresAuthRepository(db *pgxpool.Pool, logger *log.Logger) *PostgresAuthRepository {
	logger.Printf("[DEBUG] Created PostgresAuthRepository with pool address: %p", db)
	return &PostgresAuthRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresAuthRepository) WithQuerier(q database.Querier) AuthRepository {
	r.logger.Printf("[DEBUG] Swapping PostgresAuthRepository querier to: %p", q)
	return &PostgresAuthRepository{
		db:     q,
		logger: r.logger,
	}
}

func (r *PostgresAuthRepository) Register(ctx context.Context, cred *domain.Credentials) error {
	r.logger.Printf("[DEBUG] Executing AuthRepository.Register for UserID: %s", cred.UserID())

	query := ` 
	INSERT INTO auth_credentials (
		id, user_id, password_hash, last_login, failed_attempts, locked_until, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8
	)
	ON CONFLICT (id) DO UPDATE SET
		password_hash = EXCLUDED.password_hash,
		last_login = EXCLUDED.last_login,
		failed_attempts = EXCLUDED.failed_attempts,
		locked_until = EXCLUDED.locked_until,
		updated_at = NOW()
	RETURNING created_at, updated_at
	`

	args := []any{
		cred.ID(),
		cred.UserID(),
		"[REDACTED_PASSWORD_HASH]", // Security: don't log the hash
		cred.LastLogin(),
		cred.FailedAttempts(),
		cred.LockedUntil(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	}
	r.logger.Printf("[DEBUG] SQL Query: %s | Args: %+v", query, args)

	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		cred.ID(),
		cred.UserID(),
		cred.PasswordHash(), // Use actual hash for DB
		cred.LastLogin(),
		cred.FailedAttempts(),
		cred.LockedUntil(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		r.logger.Printf("[ERROR] Failed to save credentials to DB: %v", err)
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	r.logger.Printf("[DEBUG] Credentials saved successfully. CreatedAt: %v, UpdatedAt: %v", createdAt, updatedAt)
	return nil
}
