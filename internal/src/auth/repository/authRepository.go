package repository

import (
	"context"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"
)

type AuthRepository interface {
	Register(ctx context.Context, cred *domain.Credentials) error
	GetCredentialsWithUser(ctx context.Context, email string) (*domain.CredentialsWithUser, error)
	WithQuerier(q database.Querier) AuthRepository
}
