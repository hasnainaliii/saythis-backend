package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuthCredentials struct {
	id             uuid.UUID
	userID         uuid.UUID
	passwordHash   string
	lastLogin      *time.Time
	failedAttempts int
	lockedUntil    *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func NewAuthCredentials(userID uuid.UUID, passwordHash string, timeNow time.Time) *AuthCredentials {
	return &AuthCredentials{
		id:           uuid.New(),
		userID:       userID,
		passwordHash: passwordHash,
		createdAt:    timeNow,
		updatedAt:    timeNow,
	}
}

func ReconstitueAuthCredentials(
	id, userID uuid.UUID,
	passwordHash string,
	lastLogin *time.Time,
	failedAttempts int,
	lockedUntil *time.Time,
	createdAt, updatedAt time.Time,
) *AuthCredentials {
	return &AuthCredentials{
		id:             id,
		userID:         userID,
		passwordHash:   passwordHash,
		lastLogin:      lastLogin,
		failedAttempts: failedAttempts,
		lockedUntil:    lockedUntil,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (c *AuthCredentials) ID() uuid.UUID           { return c.id }
func (c *AuthCredentials) UserID() uuid.UUID       { return c.userID }
func (c *AuthCredentials) PasswordHash() string    { return c.passwordHash }
func (c *AuthCredentials) LastLogin() *time.Time   { return c.lastLogin }
func (c *AuthCredentials) FailedAttempts() int     { return c.failedAttempts }
func (c *AuthCredentials) LockedUntil() *time.Time { return c.lockedUntil }
func (c *AuthCredentials) CreatedAt() time.Time    { return c.createdAt }
func (c *AuthCredentials) UpdatedAt() time.Time    { return c.updatedAt }

func (c *AuthCredentials) IsLocked() bool {
	return c.lockedUntil != nil && time.Now().UTC().Before(*c.lockedUntil)
}
