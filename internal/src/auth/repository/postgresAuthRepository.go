package repository

import (
	"context"
	"errors"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"

	"github.com/jackc/pgx/v5"
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

func (r *PostgresAuthRepository) GetCredentialsWithUser(ctx context.Context, email string) (*domain.CredentialsWithUser, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.role, u.status, c.password_hash, u.created_at
		FROM users u
		JOIN auth_credentials c ON u.id = c.user_id
		WHERE u.email = $1
	`

	var cred domain.CredentialsWithUser
	err := r.db.QueryRow(ctx, query, email).Scan(
		&cred.UserID,
		&cred.Email,
		&cred.FullName,
		&cred.Role,
		&cred.Status,
		&cred.PasswordHash,
		&cred.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, apperror.Wrap(err, "DATABASE_ERROR", "failed to fetch credentials", 500)
	}

	return &cred, nil
}
