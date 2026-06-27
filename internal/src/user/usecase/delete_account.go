package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

func (uc *UserUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {

	if err := uc.userRepo.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	if err := uc.authRepo.DeleteAllRefreshTokensByUserID(ctx, userID); err != nil {
		slog.Warn("delete_account: failed to revoke refresh tokens",
			"user_id", userID,
			"error", err,
		)
	}

	slog.Info("user account deleted", "user_id", userID)
	return nil
}
