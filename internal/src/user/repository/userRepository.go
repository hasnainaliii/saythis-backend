package repository

import (
	"context"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	SoftDelete(ctx context.Context, userID string) error
	WithQuerier(q database.Querier) UserRepository
}
