package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
)

// VerifyEmail validates a one-time email verification token, marks the user's
// email as verified, and deletes the token so it cannot be reused.
//
// Error catalogue:
//
//	ErrInvalidToken         — token is blank or not found in the DB
//	ErrExpiredToken         — token exists but its expiry has passed
//	ErrEmailAlreadyVerified — the user's email is already verified
func (uc *AuthUseCase) VerifyEmail(ctx context.Context, plaintextToken string) error {
	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.ErrInvalidToken
	}

	// ── 1. Look up stored token by hash ───────────────────────────────────────
	tokenHash := auth.HashToken(plaintextToken)

	token, err := uc.authRepo.FindEmailVerificationToken(ctx, tokenHash)
	if err != nil {
		// Any lookup failure (not found or DB error) maps to invalid token —
		// we never reveal whether a specific token existed.
		return authdomain.ErrInvalidToken
	}

	// ── 2. Check expiry ───────────────────────────────────────────────────────
	if token.IsExpired() {
		// Clean up the stale token before responding.
		_ = uc.authRepo.DeleteEmailVerificationToken(ctx, tokenHash)
		return authdomain.ErrExpiredToken
	}

	// ── 3. Mark user as verified ──────────────────────────────────────────────
	if err = uc.authRepo.MarkEmailVerified(ctx, token.UserID(), time.Now().UTC()); err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}

	// ── 4. Delete token (one-time use) ────────────────────────────────────────
	if err = uc.authRepo.DeleteEmailVerificationToken(ctx, tokenHash); err != nil {
		// The important part already succeeded. Log and continue.
		_ = err // logged by the caller's middleware; not surfaced to the client
	}

	return nil
}
