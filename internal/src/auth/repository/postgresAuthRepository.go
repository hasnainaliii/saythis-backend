package repository

import (
	"context"
	"errors"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"
	"time"

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

func (r *PostgresAuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserBasic, error) {
	query := `
		SELECT id, email
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user domain.UserBasic
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil to indicate user not found (for security)
		}
		return nil, apperror.Wrap(err, "DATABASE_ERROR", "failed to fetch user", 500)
	}

	return &user, nil
}

func (r *PostgresAuthRepository) CreatePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to create password reset token", 500)
	}

	return nil
}

func (r *PostgresAuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var t domain.PasswordResetToken
	err := r.db.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New("INVALID_TOKEN", "Invalid or expired reset token", 400)
		}
		return nil, apperror.Wrap(err, "DATABASE_ERROR", "failed to fetch password reset token", 500)
	}

	return &t, nil
}

func (r *PostgresAuthRepository) MarkTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token = $1
	`

	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to mark token as used", 500)
	}

	return nil
}

func (r *PostgresAuthRepository) UpdatePassword(ctx context.Context, userID, newPasswordHash string) error {
	query := `
		UPDATE auth_credentials
		SET password_hash = $1, updated_at = NOW()
		WHERE user_id = $2
	`

	tag, err := r.db.Exec(ctx, query, newPasswordHash, userID)
	if err != nil {
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to update password", 500)
	}

	if tag.RowsAffected() == 0 {
		return apperror.New("USER_NOT_FOUND", "User not found", 404)
	}

	return nil
}
