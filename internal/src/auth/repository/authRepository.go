package repository

import (
	"context"
	"saythis-backend/internal/src/auth/domain"
)

type AuthRepository interface {
	Register(ctx context.Context, cred *domain.Credentials) error
}
