package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) (*domain.User, error) {

	fullName = strings.TrimSpace(fullName)
	if err := domain.ValidateFullName(fullName); err != nil {
		return nil, err
	}

	updatedAt := time.Now().UTC()
	user, err := uc.userRepo.UpdateFullName(ctx, userID, fullName, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return user, nil
}
