package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
}

func NewRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		id:        uuid.New(),
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}
}

func ReconstitueRefreshToken(id, userID uuid.UUID, tokenHash string, expiresAt, createdAt time.Time) *RefreshToken {
	return &RefreshToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

func (t *RefreshToken) ID() uuid.UUID        { return t.id }
func (t *RefreshToken) UserID() uuid.UUID    { return t.userID }
func (t *RefreshToken) TokenHash() string    { return t.tokenHash }
func (t *RefreshToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *RefreshToken) CreatedAt() time.Time { return t.createdAt }

func (t *RefreshToken) IsExpired() bool { return time.Now().UTC().After(t.expiresAt) }

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
