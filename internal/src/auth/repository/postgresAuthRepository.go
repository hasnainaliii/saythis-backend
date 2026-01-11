package repository

import (
	"context"
	"fmt"
	"saythis-backend/internal/src/auth/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuthRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAuthRepository(db *pgxpool.Pool) *PostgresAuthRepository {

	return &PostgresAuthRepository{
		db: db,
	}
}

func (r *PostgresAuthRepository) Register(ctx context.Context, cred *domain.Credentials) error {
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

	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		cred.ID(),
		cred.UserID(),
		cred.PasswordHash(),
		cred.LastLogin(),
		cred.FailedAttempts(),
		cred.LockedUntil(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}
