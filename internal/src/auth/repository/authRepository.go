package repository

import (
	"context"
	"saythis-backend/internal/database"
	"saythis-backend/internal/src/auth/domain"
	"time"
)

type AuthRepository interface {
	Register(ctx context.Context, cred *domain.Credentials) error
	GetCredentialsWithUser(ctx context.Context, email string) (*domain.CredentialsWithUser, error)
	WithQuerier(q database.Querier) AuthRepository

	// Password reset methods
	CreatePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error)
	MarkTokenAsUsed(ctx context.Context, token string) error
	UpdatePassword(ctx context.Context, userID, newPasswordHash string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.UserBasic, error)
}
