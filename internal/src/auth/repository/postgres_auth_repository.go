package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

var _ AuthRepository = (*PostgresAuthRepo)(nil)

const pgUniqueViolation = "23505"

type PostgresAuthRepo struct {
	db *pgxpool.Pool
}

func NewPostgresAuthRepo(db *pgxpool.Pool) *PostgresAuthRepo {
	return &PostgresAuthRepo{db: db}
}

func (r *PostgresAuthRepo) Register(ctx context.Context, user *userdomain.User, creds *authdomain.AuthCredentials) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userQuery := `
		INSERT INTO users (id, email, full_name, role, status, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, userQuery,
		user.ID(), user.Email(), user.FullName(), user.Role(),
		user.Status(), user.EmailVerifiedAt(), user.CreatedAt(), user.UpdatedAt(),
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		slog.Error("failed to insert user", "error", err, "email", user.Email())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return userdomain.ErrDuplicateEmail
		}
		return fmt.Errorf("insert user: %w", err)
	}
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	// ── 2. Insert credentials ────────────────────────────────────────────────
	credsQuery := `
		INSERT INTO auth_credentials (id, user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, credsQuery,
		creds.ID(), creds.UserID(), creds.PasswordHash(), creds.CreatedAt(), creds.UpdatedAt(),
	)
	if err != nil {
		slog.Error("failed to insert credentials", "error", err, "user_id", creds.UserID())
		return fmt.Errorf("insert credentials: %w", err)
	}

	// ── 3. Commit ────────────────────────────────────────────────────────────
	if err = tx.Commit(ctx); err != nil {
		slog.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("user registered", "user_id", user.ID(), "email", user.Email())
	return nil
}

// FindCredentialsByUserID returns the auth_credentials row for the given user ID.
// Every column — including nullable ones — is scanned out and used to reconstitute
// the AuthCredentials domain object. Returns ErrCredentialsNotFound if no row exists.
func (r *PostgresAuthRepo) FindCredentialsByUserID(ctx context.Context, userID uuid.UUID) (*authdomain.AuthCredentials, error) {
	query := `
		SELECT id, user_id, password_hash,
		       last_login, failed_attempts, locked_until,
		       created_at, updated_at
		FROM auth_credentials
		WHERE user_id = $1
	`
	var (
		id             uuid.UUID
		dbUserID       uuid.UUID
		passwordHash   string
		lastLogin      *time.Time
		failedAttempts int
		lockedUntil    *time.Time
		createdAt      time.Time
		updatedAt      time.Time
	)
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&id, &dbUserID, &passwordHash,
		&lastLogin, &failedAttempts, &lockedUntil,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authdomain.ErrCredentialsNotFound
		}
		return nil, fmt.Errorf("find credentials by user_id: %w", err)
	}
	return authdomain.ReconstitueAuthCredentials(
		id, dbUserID, passwordHash,
		lastLogin, failedAttempts, lockedUntil,
		createdAt, updatedAt,
	), nil
}

// UpdateLastLogin sets last_login to the provided timestamp on the credentials
// row that belongs to the given user. Called after every successful login so
// administrators have an audit trail of account activity.
func (r *PostgresAuthRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error {
	query := `
		UPDATE auth_credentials
		SET last_login      = $1,
		    failed_attempts = 0,
		    locked_until    = NULL,
		    updated_at      = NOW()
		WHERE user_id = $2
	`
	_, err := r.db.Exec(ctx, query, lastLogin, userID)
	if err != nil {
		return fmt.Errorf("update last_login: %w", err)
	}
	return nil
}

// SaveRefreshToken inserts a new refresh token record.
func (r *PostgresAuthRepo) SaveRefreshToken(ctx context.Context, token *authdomain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		token.ID(), token.UserID(), token.TokenHash(), token.ExpiresAt(), token.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// FindRefreshToken looks up a refresh token by its SHA-256 hash.
func (r *PostgresAuthRepo) FindRefreshToken(ctx context.Context, tokenHash string) (*authdomain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var (
		id        uuid.UUID
		userID    uuid.UUID
		hash      string
		expiresAt time.Time
		createdAt time.Time
	)
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&id, &userID, &hash, &expiresAt, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authdomain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return authdomain.ReconstitueRefreshToken(id, userID, hash, expiresAt, createdAt), nil
}

// DeleteRefreshToken removes a refresh token by its SHA-256 hash.
func (r *PostgresAuthRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteAllRefreshTokensByUserID removes every refresh token that belongs to the
// given user. Used during account deletion to immediately invalidate all active
// sessions so that outstanding tokens cannot be replayed.
func (r *PostgresAuthRepo) DeleteAllRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete all refresh tokens for user: %w", err)
	}
	return nil
}

// ── Email Verification ────────────────────────────────────────────────────────

// SaveEmailVerificationToken inserts a new email verification token record.
func (r *PostgresAuthRepo) SaveEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		token.ID(), token.UserID(), token.TokenHash(), token.ExpiresAt(), token.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save email verification token: %w", err)
	}
	return nil
}

// FindEmailVerificationToken looks up an email verification token by its SHA-256 hash.
func (r *PostgresAuthRepo) FindEmailVerificationToken(ctx context.Context, tokenHash string) (*authdomain.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM email_verification_tokens
		WHERE token_hash = $1
	`
	var (
		id        uuid.UUID
		userID    uuid.UUID
		hash      string
		expiresAt time.Time
		createdAt time.Time
	)
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&id, &userID, &hash, &expiresAt, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authdomain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find email verification token: %w", err)
	}
	return authdomain.ReconstitueEmailVerificationToken(id, userID, hash, expiresAt, createdAt), nil
}

// DeleteEmailVerificationToken removes an email verification token by its hash.
func (r *PostgresAuthRepo) DeleteEmailVerificationToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM email_verification_tokens WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("delete email verification token: %w", err)
	}
	return nil
}

// MarkEmailVerified stamps email_verified_at on the users table for the given user.
func (r *PostgresAuthRepo) MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error {
	query := `
		UPDATE users
		SET email_verified_at = $2,
		    updated_at        = NOW()
		WHERE id = $1
	`
	tag, err := r.db.Exec(ctx, query, userID, verifiedAt)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return userdomain.ErrUserNotFound
	}
	return nil
}

// ── Password Reset ────────────────────────────────────────────────────────────

// SavePasswordResetToken inserts a new password reset token record.
func (r *PostgresAuthRepo) SavePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		token.ID(), token.UserID(), token.TokenHash(), token.ExpiresAt(), token.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save password reset token: %w", err)
	}
	return nil
}

// FindPasswordResetToken looks up a password reset token by its SHA-256 hash.
func (r *PostgresAuthRepo) FindPasswordResetToken(ctx context.Context, tokenHash string) (*authdomain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	var (
		id        uuid.UUID
		userID    uuid.UUID
		hash      string
		expiresAt time.Time
		createdAt time.Time
	)
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&id, &userID, &hash, &expiresAt, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authdomain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find password reset token: %w", err)
	}
	return authdomain.ReconstituePasswordResetToken(id, userID, hash, expiresAt, createdAt), nil
}

// DeletePasswordResetToken removes a password reset token by its hash.
func (r *PostgresAuthRepo) DeletePasswordResetToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM password_reset_tokens WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("delete password reset token: %w", err)
	}
	return nil
}

// RecordFailedAttempt increments failed_attempts by 1.
// When the running total reaches 3 the account is locked for 24 hours
// by setting locked_until atomically in the same UPDATE statement.
func (r *PostgresAuthRepo) RecordFailedAttempt(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE auth_credentials
		SET failed_attempts = failed_attempts + 1,
		    locked_until    = CASE
		                          WHEN failed_attempts + 1 >= 3
		                          THEN NOW() + INTERVAL '24 hours'
		                          ELSE locked_until
		                      END,
		    updated_at      = NOW()
		WHERE user_id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("record failed attempt: %w", err)
	}
	return nil
}

// UpdatePassword replaces the bcrypt password hash in auth_credentials for the given user.
func (r *PostgresAuthRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE auth_credentials
		SET password_hash = $2,
		    updated_at    = NOW()
		WHERE user_id = $1
	`
	tag, err := r.db.Exec(ctx, query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return authdomain.ErrCredentialsNotFound
	}
	return nil
}
