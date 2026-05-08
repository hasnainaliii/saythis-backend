package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"saythis-backend/internal/user/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
