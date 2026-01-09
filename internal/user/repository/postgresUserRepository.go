package repository

import (
	"context"
	"fmt"
	"saythis-backend/internal/user/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	if db == nil {
		panic("db pool is nil")
	}

	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// query := `
	// INSERT INTO users (
	// 	id, email, full_name, role, status, email_verified_at, created_at, updated_at
	// ) VALUES (
	// 	$1,$2,$3,$4,$5,$6,$7,$8
	// )
	// ON CONFLICT (id) DO UPDATE SET
	// 	email = EXCLUDED.email,
	// 	full_name = EXCLUDED.full_name,
	// 	role = EXCLUDED.role,
	// 	status = EXCLUDED.status,
	// 	email_verified_at = EXCLUDED.email_verified_at,
	// 	updated_at = NOW()
	// RETURNING created_at, updated_at
	// `
	query := `
        INSERT INTO users (
            id, email, full_name, role, status, email_verified_at, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
		return fmt.Errorf("failed to save user: %w", err)
	}

	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)
	return nil
}
