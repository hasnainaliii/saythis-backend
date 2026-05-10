package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

// UpdateProfile applies profile changes for the authenticated user.
// Currently supports updating the full name; additional fields (bio, avatar, etc.)
// can be added to UpdateProfileInput in the future without breaking callers.
//
// Validation is performed against domain rules before any database call, so
// invalid input is rejected cheaply without a round-trip.
//
// Error catalogue:
//
//	ErrEmptyFullName         — full_name field was blank
//	ErrInvalidFullNameLength — full_name does not meet length constraints
//	ErrUserNotFound          — no active user exists for the given ID
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) (*domain.User, error) {

	// ── 1. Normalise + validate input ─────────────────────────────────────────
	fullName = strings.TrimSpace(fullName)
	if err := domain.ValidateFullName(fullName); err != nil {
		return nil, err
	}

	// ── 2. Persist and return the refreshed user ──────────────────────────────
	updatedAt := time.Now().UTC()
	user, err := uc.userRepo.UpdateFullName(ctx, userID, fullName, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	return user, nil
}
