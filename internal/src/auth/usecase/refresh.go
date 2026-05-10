package usecase

import (
	"context"
	"fmt"
	"strings"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
)

// Refresh validates an incoming refresh token, rotates it (delete old → save new),
// and returns a fresh token pair. The client must store and present the new refresh
// token on the next call — the old one is invalidated immediately.
//
// Rotation guarantees that a stolen refresh token cannot be silently reused:
// the first legitimate use invalidates it, so any subsequent attempt with the
// same token will fail.
func (uc *AuthUseCase) Refresh(ctx context.Context, plaintextToken string) (authdomain.TokenPair, error) {
	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.TokenPair{}, authdomain.ErrInvalidToken
	}

	// ── 1. Look up the stored token by its hash ───────────────────────────────
	tokenHash := auth.HashRefreshToken(plaintextToken)

	stored, err := uc.authRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		// Treat any lookup failure as invalid — don't leak whether the token exists.
		return authdomain.TokenPair{}, authdomain.ErrTokenNotFound
	}

	// ── 2. Guard against expired tokens ──────────────────────────────────────
	if stored.IsExpired() {
		// Clean up the stale token and tell the client to log in again.
		_ = uc.authRepo.DeleteRefreshToken(ctx, tokenHash)
		return authdomain.TokenPair{}, authdomain.ErrExpiredToken
	}

	// ── 3. Rotate: invalidate the used token before issuing a new one ─────────
	if err = uc.authRepo.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("rotate token: %w", err)
	}

	// ── 4. Fetch the user so the new access token contains current claims ──────
	user, err := uc.userRepo.GetByID(ctx, stored.UserID())
	if err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("get user for refresh: %w", err)
	}

	// ── 5. Issue and return a fresh token pair ────────────────────────────────
	return uc.issueTokenPair(ctx, user)
}
