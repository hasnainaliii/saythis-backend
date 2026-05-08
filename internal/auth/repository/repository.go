package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	authdomain "saythis-backend/internal/auth/domain"
	userdomain "saythis-backend/internal/user/domain"
)

type AuthRepository interface {
	Register(ctx context.Context, user *userdomain.User, creds *authdomain.AuthCredentials) error

	FindCredentialsByUserID(ctx context.Context, userID uuid.UUID) (*authdomain.AuthCredentials, error)

	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error

	SaveRefreshToken(ctx context.Context, token *authdomain.RefreshToken) error

	FindRefreshToken(ctx context.Context, tokenHash string) (*authdomain.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}
