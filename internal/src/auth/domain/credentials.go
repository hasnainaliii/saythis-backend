package domain

import (
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	id             uuid.UUID
	userID         uuid.UUID
	passwordHash   string
	lastLogin      *time.Time
	failedAttempts int
	lockedUntil    *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func NewCredentials(id uuid.UUID, userId uuid.UUID, passwordHash string, now time.Time) (*Credentials, error) {

	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	return &Credentials{
		id:           id,
		userID:       userId,
		passwordHash: passwordHash,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ****************
// --- Getters ---
// ****************

func (c *Credentials) ID() uuid.UUID           { return c.id }
func (c *Credentials) UserID() uuid.UUID       { return c.userID }
func (c *Credentials) PasswordHash() string    { return c.passwordHash }
func (c *Credentials) LastLogin() *time.Time   { return c.lastLogin }
func (c *Credentials) FailedAttempts() int     { return c.failedAttempts }
func (c *Credentials) LockedUntil() *time.Time { return c.lockedUntil }
func (c *Credentials) CreatedAt() time.Time    { return c.createdAt }
func (c *Credentials) UpdatedAt() time.Time    { return c.updatedAt }

func (c *Credentials) RecordLogin(now time.Time) error {
	c.lastLogin = &now
	c.failedAttempts = 0
	c.updatedAt = now
	return nil
}
