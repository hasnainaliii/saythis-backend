package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"saythis-backend/internal/user/domain"
	"time"

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
	if user == nil {
		return errors.New("user is nil")
	}

	query := `
		INSERT INTO users (id, email, full_name, role, status, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query,
		user.ID(),
		user.Email(),
		user.FullName(),
		user.Role(),
		user.Status(),
		user.EmailVerifiedAt(),
		user.CreatedAt(),
		user.UpdatedAt(),
	).Scan(&createdAt, &updatedAt)

	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrDuplicateEmail
		}

		slog.Error("failed to insert user", "error", err)
		return fmt.Errorf("failed to save user: %w", err)
	}

	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)

	slog.Info("user created", "user_id", user.ID(), "email", user.Email())
	return nil
}
