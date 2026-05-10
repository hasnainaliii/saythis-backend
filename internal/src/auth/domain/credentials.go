package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuthCredentials holds the security material for a single user account.
// It maps 1-to-1 with the auth_credentials table.
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

// NewAuthCredentials creates fresh credentials for a newly registered user.
// Used exclusively by the registration flow where there is no prior state.
func NewAuthCredentials(userID uuid.UUID, passwordHash string, timeNow time.Time) *AuthCredentials {
	return &AuthCredentials{
		id:           uuid.New(),
		userID:       userID,
		passwordHash: passwordHash,
		createdAt:    timeNow,
		updatedAt:    timeNow,
	}
}

// ReconstitueAuthCredentials rebuilds an AuthCredentials from a database row.
// Unlike NewAuthCredentials, it accepts all persisted fields (including nullable
// ones) and skips validation because data retrieved from the DB is already trusted.
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

// ── Getters ──────────────────────────────────────────────────────────────────

func (c *AuthCredentials) ID() uuid.UUID           { return c.id }
func (c *AuthCredentials) UserID() uuid.UUID       { return c.userID }
func (c *AuthCredentials) PasswordHash() string    { return c.passwordHash }
func (c *AuthCredentials) LastLogin() *time.Time   { return c.lastLogin }
func (c *AuthCredentials) FailedAttempts() int     { return c.failedAttempts }
func (c *AuthCredentials) LockedUntil() *time.Time { return c.lockedUntil }
func (c *AuthCredentials) CreatedAt() time.Time    { return c.createdAt }
func (c *AuthCredentials) UpdatedAt() time.Time    { return c.updatedAt }

// ── Domain logic ──────────────────────────────────────────────────────────────

// IsLocked reports whether the account is currently in a lockout period.
// A non-nil lockedUntil that is still in the future means the account is locked.
func (c *AuthCredentials) IsLocked() bool {
	return c.lockedUntil != nil && time.Now().UTC().Before(*c.lockedUntil)
}
