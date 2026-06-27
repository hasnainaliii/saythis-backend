package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error

	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	SoftDelete(ctx context.Context, id uuid.UUID) error

	UpdateFullName(ctx context.Context, id uuid.UUID, fullName string, updatedAt time.Time) (*domain.User, error)

	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string, updatedAt time.Time) (*domain.User, error)
}
