package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmailVerificationToken represents a one-time token sent to a user's email
// after registration. It maps 1-to-1 with the email_verification_tokens table.
//
// Security: only the SHA-256 hash is stored in the DB. The plaintext token
// is sent to the user's inbox and never persisted.
type EmailVerificationToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

// NewEmailVerificationToken constructs a fresh token ready to be persisted.
func NewEmailVerificationToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *EmailVerificationToken {
	return &EmailVerificationToken{
		id:        uuid.New(),
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

// ReconstitueEmailVerificationToken rebuilds the domain object from a DB row.
func ReconstitueEmailVerificationToken(id, userID uuid.UUID, tokenHash string, expiresAt, createdAt time.Time) *EmailVerificationToken {
	return &EmailVerificationToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

// ── Getters ──────────────────────────────────────────────────────────────────

func (t *EmailVerificationToken) ID() uuid.UUID        { return t.id }
func (t *EmailVerificationToken) UserID() uuid.UUID    { return t.userID }
func (t *EmailVerificationToken) TokenHash() string    { return t.tokenHash }
func (t *EmailVerificationToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *EmailVerificationToken) CreatedAt() time.Time { return t.createdAt }

// IsExpired reports whether the token's expiry time has passed.
func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}
