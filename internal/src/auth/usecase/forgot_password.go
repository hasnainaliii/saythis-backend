package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

func (uc *AuthUseCase) ForgotPassword(ctx context.Context, rawEmail string) error {
	email := strings.ToLower(strings.TrimSpace(rawEmail))

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil
		}
		slog.Error("forgot_password: failed to look up user",
			"email", email,
			"error", err,
		)
		return nil
	}

	plaintext, tokenHash, err := auth.GenerateSecureToken()
	if err != nil {
		slog.Error("forgot_password: failed to generate reset token",
			"user_id", user.ID(),
			"error", err,
		)
		return nil
	}

	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	resetToken := authdomain.NewPasswordResetToken(user.ID(), tokenHash, expiresAt)

	if err = uc.authRepo.SavePasswordResetToken(ctx, resetToken); err != nil {
		slog.Error("forgot_password: failed to save reset token",
			"user_id", user.ID(),
			"error", err,
		)
		return nil
	}

	resetURL := uc.frontendURL + "/reset-password?token=" + plaintext

	if err = uc.emailSender.SendPasswordReset(ctx, user.Email(), resetURL); err != nil {
		slog.Error("forgot_password: failed to send reset email",
			"user_id", user.ID(),
			"error", err,
		)
	}

	slog.Info("forgot_password: reset email dispatched", "user_id", user.ID())
	return nil
}
