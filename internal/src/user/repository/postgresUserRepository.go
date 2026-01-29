package repository

import (
	"context"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/user/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db database.Querier
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	if db == nil {
		panic("db pool is nil")
	}
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) WithQuerier(q database.Querier) UserRepository {
	return &PostgresUserRepository{
		db: q,
	}
}

var userConstraintMapping = map[string]*apperror.AppError{
	"users_email_key": apperror.ErrDuplicateEmail,
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {

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

	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query, args...).Scan(&createdAt, &updatedAt)
	if err != nil {
		if appErr := apperror.MapPostgresError(err, userConstraintMapping); appErr != nil {
			return appErr
		}
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to save user", 500)
	}

	user.SetCreatedAt(createdAt)
	user.SetUpdatedAt(updatedAt)
	return nil
}
