package repository

import (
	"context"
	"fmt"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/user/domain"
	"strings"
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

func (r *PostgresUserRepository) SoftDelete(ctx context.Context, userID string) error {
	query := `
        UPDATE users 
        SET deleted_at = NOW(), 
            status = 'deleted', 
            updated_at = NOW()
        WHERE id = $1 
          AND deleted_at IS NULL
    `

	tag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to delete user", 500)
	}

	if tag.RowsAffected() == 0 {
		return apperror.New("USER_NOT_FOUND", "User not found or already deleted", 404)
	}

	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
        SELECT id, email, full_name, avatar_url, role, status, email_verified_at, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `

	var id, email, fullName, role, status string
	var avatarURL *string
	var emailVerifiedAt, deletedAt *time.Time
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&id, &email, &fullName, &avatarURL, &role, &status,
		&emailVerifiedAt, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, apperror.New("USER_NOT_FOUND", "User not found", 404)
		}
		return nil, apperror.Wrap(err, "DATABASE_ERROR", "failed to find user", 500)
	}

	// Reconstruct the user domain object
	user, err := domain.ReconstructUser(id, email, fullName, avatarURL, role, status, emailVerifiedAt, createdAt, updatedAt, deletedAt)
	if err != nil {
		return nil, apperror.Wrap(err, "DOMAIN_ERROR", "failed to reconstruct user", 500)
	}

	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return apperror.New("VALIDATION_ERROR", "No fields to update", 400)
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add userID as the last argument
	args = append(args, userID)

	query := fmt.Sprintf(`
        UPDATE users
        SET %s
        WHERE id = $%d AND deleted_at IS NULL
    `, strings.Join(setClauses, ", "), argIndex)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return apperror.Wrap(err, "DATABASE_ERROR", "failed to update user", 500)
	}

	if tag.RowsAffected() == 0 {
		return apperror.New("USER_NOT_FOUND", "User not found", 404)
	}

	return nil
}
