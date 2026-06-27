package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

const resendCooldown = 24 * time.Hour

func (uc *AuthUseCase) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return userdomain.ErrUserNotFound
		}
		return fmt.Errorf("resend_verification: look up user: %w", err)
	}

	if user.EmailVerifiedAt() != nil {
		return authdomain.ErrEmailAlreadyVerified
	}

	latest, err := uc.authRepo.FindLatestEmailVerificationTokenByUserID(ctx, userID)
	if err != nil && !errors.Is(err, authdomain.ErrTokenNotFound) {
		return fmt.Errorf("resend_verification: check rate limit: %w", err)
	}

	if latest != nil && time.Since(latest.CreatedAt()) < resendCooldown {
		return authdomain.ErrResendTooSoon
	}

	if err = uc.authRepo.DeleteEmailVerificationTokensByUserID(ctx, userID); err != nil {
		return fmt.Errorf("resend_verification: clear old tokens: %w", err)
	}

	plaintext, tokenHash, err := auth.GenerateSecureToken()
	if err != nil {
		return fmt.Errorf("resend_verification: generate token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	token := authdomain.NewEmailVerificationToken(userID, tokenHash, expiresAt)

	if err = uc.authRepo.SaveEmailVerificationToken(ctx, token); err != nil {
		return fmt.Errorf("resend_verification: save token: %w", err)
	}

	verificationURL := uc.frontendURL + "/verify-email?token=" + plaintext

	if err = uc.emailSender.SendVerification(ctx, user.Email(), verificationURL); err != nil {
		slog.Error("resend_verification: failed to send email",
			"user_id", userID,
			"error", err,
		)
		return nil
	}

	slog.Info("resend_verification: verification email dispatched", "user_id", userID)
	return nil
}
