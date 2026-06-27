package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

func NewPasswordResetToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		id:        uuid.New(),
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

func ReconstituePasswordResetToken(id, userID uuid.UUID, tokenHash string, expiresAt, createdAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

func (t *PasswordResetToken) ID() uuid.UUID        { return t.id }
func (t *PasswordResetToken) UserID() uuid.UUID    { return t.userID }
func (t *PasswordResetToken) TokenHash() string    { return t.tokenHash }
func (t *PasswordResetToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *PasswordResetToken) CreatedAt() time.Time { return t.createdAt }

func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}
