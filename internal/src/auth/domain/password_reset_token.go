package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a short-lived token issued when a user
// requests a password reset. It maps 1-to-1 with the password_reset_tokens table.
//
// Security: only the SHA-256 hash is stored in the DB. The plaintext token
// is sent to the user's inbox and never persisted.
type PasswordResetToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

// NewPasswordResetToken constructs a fresh token ready to be persisted.
func NewPasswordResetToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		id:        uuid.New(),
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

// ReconstituePasswordResetToken rebuilds the domain object from a DB row.
func ReconstituePasswordResetToken(id, userID uuid.UUID, tokenHash string, expiresAt, createdAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

// ── Getters ──────────────────────────────────────────────────────────────────

func (t *PasswordResetToken) ID() uuid.UUID        { return t.id }
func (t *PasswordResetToken) UserID() uuid.UUID    { return t.userID }
func (t *PasswordResetToken) TokenHash() string    { return t.tokenHash }
func (t *PasswordResetToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *PasswordResetToken) CreatedAt() time.Time { return t.createdAt }

// IsExpired reports whether the token's expiry time has passed.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}
