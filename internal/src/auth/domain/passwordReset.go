package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsExpired checks if the token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// UserBasic contains basic user info for password reset
type UserBasic struct {
	ID    uuid.UUID
	Email string
}
