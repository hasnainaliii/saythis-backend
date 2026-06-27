package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
)

func (uc *AuthUseCase) ResetPassword(ctx context.Context, plaintextToken, newPassword string) error {

	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.ErrInvalidToken
	}

	if strings.TrimSpace(newPassword) == "" {
		return authdomain.ErrEmptyPassword
	}
	if len(newPassword) < minPasswordLength {
		return authdomain.ErrPasswordTooShort
	}
	if len(newPassword) > maxPasswordLength {
		return authdomain.ErrPasswordTooLong
	}

	tokenHash := auth.HashToken(plaintextToken)

	token, err := uc.authRepo.FindPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return authdomain.ErrInvalidToken
	}

	if token.IsExpired() {
		_ = uc.authRepo.DeletePasswordResetToken(ctx, tokenHash)
		return authdomain.ErrExpiredToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err = uc.authRepo.UpdatePassword(ctx, token.UserID(), string(hashed)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err = uc.authRepo.DeletePasswordResetToken(ctx, tokenHash); err != nil {
		slog.Warn("reset_password: failed to delete used token",
			"user_id", token.UserID(),
			"error", err,
		)
	}

	slog.Info("reset_password: password updated successfully", "user_id", token.UserID())
	return nil
}
