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

// ResendVerificationEmail issues a fresh one-time verification link for the
// given user and delivers it by email.
//
// Rules enforced:
//   - The user's email must not already be verified (ErrEmailAlreadyVerified).
//   - At most one resend per 24 hours is allowed (ErrResendTooSoon).
//
// On success the old token(s) for this user are deleted and a new 24-hour
// token is stored and emailed.
func (uc *AuthUseCase) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {

	// ── 1. Fetch the user — confirm they exist and are not yet verified ────────
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			// Should never happen for an authenticated caller, but handle gracefully.
			return userdomain.ErrUserNotFound
		}
		return fmt.Errorf("resend_verification: look up user: %w", err)
	}

	if user.EmailVerifiedAt() != nil {
		return authdomain.ErrEmailAlreadyVerified
	}

	// ── 2. Enforce once-per-24h rate limit ────────────────────────────────────
	latest, err := uc.authRepo.FindLatestEmailVerificationTokenByUserID(ctx, userID)
	if err != nil && !errors.Is(err, authdomain.ErrTokenNotFound) {
		return fmt.Errorf("resend_verification: check rate limit: %w", err)
	}

	if latest != nil && time.Since(latest.CreatedAt()) < resendCooldown {
		return authdomain.ErrResendTooSoon
	}

	// ── 3. Remove any stale tokens before issuing a fresh one ─────────────────
	if err = uc.authRepo.DeleteEmailVerificationTokensByUserID(ctx, userID); err != nil {
		return fmt.Errorf("resend_verification: clear old tokens: %w", err)
	}

	// ── 4. Generate, persist, and send the new token ──────────────────────────
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
		// The token is saved; the user can retry again after 24 h.
		// Do not surface the mail-delivery failure — just log it.
		return nil
	}

	slog.Info("resend_verification: verification email dispatched", "user_id", userID)
	return nil
}
