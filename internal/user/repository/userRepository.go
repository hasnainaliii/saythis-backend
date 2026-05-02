package repository

import (
	"context"
	"saythis-backend/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
}
