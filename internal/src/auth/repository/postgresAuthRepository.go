package repository

import (
	"context"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuthRepository struct {
	db database.Querier
}

func NewPostgresAuthRepository(db *pgxpool.Pool) *PostgresAuthRepository {
	if db == nil {
		panic("db pool is nil")
	}
	return &PostgresAuthRepository{
		db: db,
	}
}

func (r *PostgresAuthRepository) WithQuerier(q database.Querier) AuthRepository {
	return &PostgresAuthRepository{
		db: q,
	}
}

func (r *PostgresAuthRepository) Register(ctx context.Context, cred *domain.Credentials) error {

	query := ` 
	INSERT INTO auth_credentials (
		id, user_id, password_hash, last_login, failed_attempts, locked_until, created_at, updated_at
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8
	)
	
	`

	_, err := r.db.Exec(ctx, query,
		cred.ID(),
		cred.UserID(),
		cred.PasswordHash(),
		cred.LastLogin(),
		cred.FailedAttempts(),
		cred.LockedUntil(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)

	if err != nil {
		if appErr := apperror.MapPostgresError(err, nil); appErr != nil {
			return appErr
		}
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to save credentials", 500)
	}

	return nil
}
