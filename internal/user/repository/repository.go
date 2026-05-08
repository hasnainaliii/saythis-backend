package repository

import (
	"context"

	"github.com/google/uuid"

	"saythis-backend/internal/user/domain"
)

// UserRepository handles read and write operations for user profile data.
type UserRepository interface {
	// Create inserts a new user row. Used for non-auth creation paths (e.g. admin tools).
	Create(ctx context.Context, user *domain.User) error

	// GetByID fetches a user by their primary key.
	// Returns ErrUserNotFound if no row exists.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail fetches a user by their email address (case-insensitive because
	// the underlying column is CITEXT). Returns ErrUserNotFound if no row exists.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}
