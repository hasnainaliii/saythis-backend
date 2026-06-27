package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

type AuthRepository interface {
	Register(ctx context.Context, user *userdomain.User, creds *authdomain.AuthCredentials) error

	FindCredentialsByUserID(ctx context.Context, userID uuid.UUID) (*authdomain.AuthCredentials, error)

	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLogin time.Time) error

	RecordFailedAttempt(ctx context.Context, userID uuid.UUID) error

	SaveRefreshToken(ctx context.Context, token *authdomain.RefreshToken) error

	FindRefreshToken(ctx context.Context, tokenHash string) (*authdomain.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, tokenHash string) error

	DeleteAllRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error

	SaveEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) error

	FindEmailVerificationToken(ctx context.Context, tokenHash string) (*authdomain.EmailVerificationToken, error)

	DeleteEmailVerificationToken(ctx context.Context, tokenHash string) error

	FindLatestEmailVerificationTokenByUserID(ctx context.Context, userID uuid.UUID) (*authdomain.EmailVerificationToken, error)

	DeleteEmailVerificationTokensByUserID(ctx context.Context, userID uuid.UUID) error

	MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error

	SavePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) error

	FindPasswordResetToken(ctx context.Context, tokenHash string) (*authdomain.PasswordResetToken, error)

	DeletePasswordResetToken(ctx context.Context, tokenHash string) error

	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}
