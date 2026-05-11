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

	// RecordFailedAttempt increments failed_attempts by 1 for the given user.
	// When the new count reaches 3 the account is locked for 24 hours by
	// setting locked_until = NOW() + INTERVAL '24 hours'.
	// This is a single atomic SQL UPDATE — no race between read and write.
	RecordFailedAttempt(ctx context.Context, userID uuid.UUID) error

	// ── Refresh tokens ────────────────────────────────────────────────────────

	SaveRefreshToken(ctx context.Context, token *authdomain.RefreshToken) error

	FindRefreshToken(ctx context.Context, tokenHash string) (*authdomain.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, tokenHash string) error

	// DeleteAllRefreshTokensByUserID removes every refresh token belonging to the
	// given user. Called during account deletion to immediately invalidate all
	// active sessions across every device.
	DeleteAllRefreshTokensByUserID(ctx context.Context, userID uuid.UUID) error

	// ── Email verification ─────────────────────────────────────────────────────

	// SaveEmailVerificationToken persists a hashed one-time verification token.
	SaveEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) error

	// FindEmailVerificationToken looks up a token by its SHA-256 hash.
	// Returns ErrTokenNotFound if no matching row exists.
	FindEmailVerificationToken(ctx context.Context, tokenHash string) (*authdomain.EmailVerificationToken, error)

	// DeleteEmailVerificationToken removes a verification token after use or expiry.
	DeleteEmailVerificationToken(ctx context.Context, tokenHash string) error

	// MarkEmailVerified stamps email_verified_at on the users row.
	MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error

	// ── Password reset ─────────────────────────────────────────────────────────

	// SavePasswordResetToken persists a hashed one-time password-reset token.
	SavePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) error

	// FindPasswordResetToken looks up a token by its SHA-256 hash.
	// Returns ErrTokenNotFound if no matching row exists.
	FindPasswordResetToken(ctx context.Context, tokenHash string) (*authdomain.PasswordResetToken, error)

	// DeletePasswordResetToken removes a reset token after use or expiry.
	DeletePasswordResetToken(ctx context.Context, tokenHash string) error

	// UpdatePassword replaces the bcrypt password hash in auth_credentials.
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}
