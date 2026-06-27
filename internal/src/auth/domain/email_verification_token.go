package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

func NewEmailVerificationToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *EmailVerificationToken {
	return &EmailVerificationToken{
		id:        uuid.New(),
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

func ReconstitueEmailVerificationToken(id, userID uuid.UUID, tokenHash string, expiresAt, createdAt time.Time) *EmailVerificationToken {
	return &EmailVerificationToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

func (t *EmailVerificationToken) ID() uuid.UUID        { return t.id }
func (t *EmailVerificationToken) UserID() uuid.UUID    { return t.userID }
func (t *EmailVerificationToken) TokenHash() string    { return t.tokenHash }
func (t *EmailVerificationToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *EmailVerificationToken) CreatedAt() time.Time { return t.createdAt }

func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}
