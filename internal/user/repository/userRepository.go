package repository

import (
	"context"
	"saythis-backend/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	// FindById(id uuid.UUID) (*domain.User, error)
	// FindByEmail(email string) (*domain.User, error)
	// DeleteUser(user *domain.User) error
	// update(user *domain.User) error
}
