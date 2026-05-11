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

	"saythis-backend/internal/src/user/domain"
)

const pgUniqueViolation = "23505"

type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, full_name, role, status, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query,
		user.ID(), user.Email(), user.FullName(), user.Role(),
		user.Status(), user.EmailVerifiedAt(), user.CreatedAt(), user.UpdatedAt(),
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrDuplicateEmail
		}
		slog.Error("failed to insert user", "error", err)
		return fmt.Errorf("insert user: %w", err)
	}
	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)
	return nil
}

// GetByID fetches a single user by primary key, reconstituting the domain object
// from the row data.
func (r *PostgresUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, full_name, COALESCE(avatar_url, ''), role, status,
		       email_verified_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var (
		dbID            uuid.UUID
		email           string
		fullName        string
		avatarURL       string
		role            domain.UserRole
		status          domain.UserStatus
		emailVerifiedAt *time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dbID, &email, &fullName, &avatarURL,
		&role, &status, &emailVerifiedAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return domain.ReconstitueUser(dbID, email, fullName, avatarURL, role, status, emailVerifiedAt, createdAt, updatedAt), nil
}

// SoftDelete marks the user as deleted by setting status = 'deleted'.
// The row is preserved for audit purposes.
// Returns ErrUserNotFound if no user with the given ID exists.
func (r *PostgresUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET    status     = 'deleted',
		       updated_at = NOW()
		WHERE  id = $1
	`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// UpdateFullName changes the user's full_name and returns the fully reconstituted
// user with refreshed timestamps. Only active users can have their name updated;
// deleted or suspended accounts are treated as not found.
func (r *PostgresUserRepo) UpdateFullName(ctx context.Context, id uuid.UUID, fullName string, updatedAt time.Time) (*domain.User, error) {
	query := `
		UPDATE users
		SET    full_name  = $2,
		       updated_at = $3
		WHERE  id = $1
		  AND  status    = 'active'
		RETURNING id, email, full_name, COALESCE(avatar_url, ''),
		          role, status, email_verified_at, created_at, updated_at
	`
	var (
		dbID            uuid.UUID
		email           string
		dbFullName      string
		avatarURL       string
		role            domain.UserRole
		status          domain.UserStatus
		emailVerifiedAt *time.Time
		createdAt       time.Time
		dbUpdatedAt     time.Time
	)
	err := r.db.QueryRow(ctx, query, id, fullName, updatedAt).Scan(
		&dbID, &email, &dbFullName, &avatarURL,
		&role, &status, &emailVerifiedAt,
		&createdAt, &dbUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update full name: %w", err)
	}
	return domain.ReconstitueUser(dbID, email, dbFullName, avatarURL, role, status, emailVerifiedAt, createdAt, dbUpdatedAt), nil
}

// UpdateAvatarURL sets avatar_url for the given active user and returns the
// fully reconstituted user with refreshed timestamps.
func (r *PostgresUserRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string, updatedAt time.Time) (*domain.User, error) {
	query := `
		UPDATE users
		SET    avatar_url = $2,
		       updated_at = $3
		WHERE  id     = $1
		  AND  status = 'active'
		RETURNING id, email, full_name, COALESCE(avatar_url, ''),
		          role, status, email_verified_at, created_at, updated_at
	`
	var (
		dbID            uuid.UUID
		email           string
		fullName        string
		dbAvatarURL     string
		role            domain.UserRole
		status          domain.UserStatus
		emailVerifiedAt *time.Time
		createdAt       time.Time
		dbUpdatedAt     time.Time
	)
	err := r.db.QueryRow(ctx, query, id, avatarURL, updatedAt).Scan(
		&dbID, &email, &fullName, &dbAvatarURL,
		&role, &status, &emailVerifiedAt,
		&createdAt, &dbUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update avatar url: %w", err)
	}
	return domain.ReconstitueUser(dbID, email, fullName, dbAvatarURL, role, status, emailVerifiedAt, createdAt, dbUpdatedAt), nil
}

// GetByEmail fetches a single user by email address.
// The email comparison is case-insensitive because the column is typed as CITEXT.
// Returns ErrUserNotFound if no row matches.
func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, full_name, COALESCE(avatar_url, ''), role, status,
		       email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var (
		dbID            uuid.UUID
		dbEmail         string
		fullName        string
		avatarURL       string
		role            domain.UserRole
		status          domain.UserStatus
		emailVerifiedAt *time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)
	err := r.db.QueryRow(ctx, query, email).Scan(
		&dbID, &dbEmail, &fullName, &avatarURL,
		&role, &status, &emailVerifiedAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return domain.ReconstitueUser(dbID, dbEmail, fullName, avatarURL, role, status, emailVerifiedAt, createdAt, updatedAt), nil
}
