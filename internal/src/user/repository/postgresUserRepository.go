package repository

import (
	"context"
	"fmt"
	"log"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/user/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db     database.Querier
	logger *log.Logger
}

func NewPostgresUserRepository(db *pgxpool.Pool, logger *log.Logger) *PostgresUserRepository {
	if db == nil {
		panic("db pool is nil")
	}
	logger.Printf("[DEBUG] Created PostgresUserRepository with pool address: %p", db)
	logger.Printf("[DEBUG] Created PostgresUserRepository with pool address: %p", &db)
	return &PostgresUserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresUserRepository) WithQuerier(q database.Querier) UserRepository {
	r.logger.Printf("[DEBUG] Swapping PostgresUserRepository querier to: %p", q)
	return &PostgresUserRepository{
		db:     q,
		logger: r.logger,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.logger.Printf("[DEBUG] Executing UserRepository.Create for user: %s (ID: %s)", user.Email(), user.ID())

	query := `
        INSERT INTO users (
            id, email, full_name, role, status, email_verified_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING created_at, updated_at
    `

	args := []any{
		user.ID(),
		user.Email(),
		user.FullName(),
		user.Role(),
		user.Status(),
		user.EmailVerifiedAt(),
		user.CreatedAt(),
		user.UpdatedAt(),
	}
	r.logger.Printf("[DEBUG] SQL Query: %s | Args: %+v", query, args)

	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query, args...).Scan(&createdAt, &updatedAt)
	if err != nil {
		r.logger.Printf("[ERROR] Failed to save user to DB: %v", err)
		return fmt.Errorf("failed to save user: %w", err)
	}

	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)
	r.logger.Printf("[DEBUG] User saved successfully. CreatedAt: %v, UpdatedAt: %v", createdAt, updatedAt)
	return nil
}
