package repository

import (
	"context"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, userID string) (*domain.User, error)
	Update(ctx context.Context, userID string, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, userID string) error
	WithQuerier(q database.Querier) UserRepository
}
