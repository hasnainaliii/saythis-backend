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

// ForgotPassword initiates the password-reset flow for the given email address.
//
// Security contract: this method ALWAYS returns nil to the caller, regardless
// of whether the email exists in the system. This prevents user enumeration —
// an attacker cannot determine which email addresses are registered.
//
// If a valid account is found, a short-lived reset token is generated and
// an email is dispatched. Any internal failure (token generation, DB write,
// email send) is logged but never propagated.
func (uc *AuthUseCase) ForgotPassword(ctx context.Context, rawEmail string) error {
	email := strings.ToLower(strings.TrimSpace(rawEmail))

	// ── 1. Look up the user — silently no-op if not found ─────────────────────
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			// Return nil — do not reveal that this email is not registered.
			return nil
		}
		// Unexpected DB error — log it but still return nil to the client.
		slog.Error("forgot_password: failed to look up user",
			"email", email,
			"error", err,
		)
		return nil
	}

	// ── 2. Generate a cryptographically secure reset token ────────────────────
	plaintext, tokenHash, err := auth.GenerateSecureToken()
	if err != nil {
		slog.Error("forgot_password: failed to generate reset token",
			"user_id", user.ID(),
			"error", err,
		)
		return nil
	}

	// ── 3. Persist the hashed token (expires in 15 minutes) ──────────────────
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	resetToken := authdomain.NewPasswordResetToken(user.ID(), tokenHash, expiresAt)

	if err = uc.authRepo.SavePasswordResetToken(ctx, resetToken); err != nil {
		slog.Error("forgot_password: failed to save reset token",
			"user_id", user.ID(),
			"error", err,
		)
		return nil
	}

	// ── 4. Send the reset email ───────────────────────────────────────────────
	resetURL := uc.frontendURL + "/reset-password?token=" + plaintext

	if err = uc.emailSender.SendPasswordReset(ctx, user.Email(), resetURL); err != nil {
		slog.Error("forgot_password: failed to send reset email",
			"user_id", user.ID(),
			"error", err,
		)
		// Return nil — do not signal to the client that anything went wrong.
	}

	slog.Info("forgot_password: reset email dispatched", "user_id", user.ID())
	return nil
}
