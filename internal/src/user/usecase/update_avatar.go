package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/user/domain"
)

func (uc *UserUseCase) UpdateAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, filename string) (*domain.User, error) {

	secureURL, err := uc.uploader.Upload(ctx, file, filename)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}

	user, err := uc.userRepo.UpdateAvatarURL(ctx, userID, secureURL, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("save avatar url: %w", err)
	}

	return user, nil
}
