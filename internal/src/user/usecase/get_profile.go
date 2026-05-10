package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

// GetProfile returns the full profile of the authenticated user.
//
// Error catalogue:
//
//	ErrUserNotFound — no active user exists for the given ID
func (uc *UserUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}
