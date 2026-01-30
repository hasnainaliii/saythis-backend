package domain

import (
	"time"

	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

type CredentialsWithUser struct {
	UserID       uuid.UUID
	Email        string
	FullName     string
	Role         string
	Status       string
	PasswordHash string
	CreatedAt    time.Time
}
