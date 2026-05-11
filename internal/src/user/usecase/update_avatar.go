package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

// UpdateAvatar uploads a new avatar image to Cloudinary and persists the
// returned secure URL on the user's profile.
//
// Error catalogue:
//
//	ErrUserNotFound — no active user exists for the given ID
func (uc *UserUseCase) UpdateAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, filename string) (*domain.User, error) {

	// ── 1. Upload to Cloudinary ───────────────────────────────────────────────
	secureURL, err := uc.uploader.Upload(ctx, file, filename)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	// ── 2. Persist the new URL ────────────────────────────────────────────────
	user, err := uc.userRepo.UpdateAvatarURL(ctx, userID, secureURL, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("save avatar url: %w", err)
	}

	return user, nil
}
