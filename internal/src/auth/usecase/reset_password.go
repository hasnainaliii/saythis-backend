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

// ResetPassword validates a password-reset token, replaces the user's bcrypt
// password hash, and deletes the token so it cannot be reused.
//
// Validation order:
//  1. Token present (not blank)
//  2. New password strength (empty / too short / too long)
//  3. Token found in DB (maps to ErrInvalidToken on failure)
//  4. Token not expired (deletes stale token, returns ErrExpiredToken)
//  5. Bcrypt hash + credential update
//  6. Token deletion (one-time use)
//
// Error catalogue:
//
//	ErrInvalidToken  — token is blank or not found in the DB
//	ErrExpiredToken  — token exists but its expiry has passed
//	ErrEmptyPassword — new password field was blank
//	ErrPasswordTooShort / ErrPasswordTooLong — password fails length rules
func (uc *AuthUseCase) ResetPassword(ctx context.Context, plaintextToken, newPassword string) error {

	// ── 1. Token presence check ───────────────────────────────────────────────
	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.ErrInvalidToken
	}

	// ── 2. New password validation ────────────────────────────────────────────
	if strings.TrimSpace(newPassword) == "" {
		return authdomain.ErrEmptyPassword
	}
	if len(newPassword) < minPasswordLength {
		return authdomain.ErrPasswordTooShort
	}
	if len(newPassword) > maxPasswordLength {
		return authdomain.ErrPasswordTooLong
	}

	// ── 3. Look up stored token by hash ───────────────────────────────────────
	tokenHash := auth.HashToken(plaintextToken)

	token, err := uc.authRepo.FindPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return authdomain.ErrInvalidToken
	}

	// ── 4. Check expiry ───────────────────────────────────────────────────────
	if token.IsExpired() {
		_ = uc.authRepo.DeletePasswordResetToken(ctx, tokenHash)
		return authdomain.ErrExpiredToken
	}

	// ── 5. Hash new password ──────────────────────────────────────────────────
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// ── 6. Persist the new credential ─────────────────────────────────────────
	if err = uc.authRepo.UpdatePassword(ctx, token.UserID(), string(hashed)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// ── 7. Delete token (one-time use) ────────────────────────────────────────
	if err = uc.authRepo.DeletePasswordResetToken(ctx, tokenHash); err != nil {
		slog.Warn("reset_password: failed to delete used token",
			"user_id", token.UserID(),
			"error", err,
		)
	}

	slog.Info("reset_password: password updated successfully", "user_id", token.UserID())
	return nil
}
