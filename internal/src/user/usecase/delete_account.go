package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// DeleteAccount soft-deletes the authenticated user's account and immediately
// revokes all of their refresh tokens.
//
// The user row is not physically removed — setting status = 'deleted' is
// sufficient to prevent future logins while retaining the data for audit trails.
//
// Token revocation is best-effort: if the DELETE on refresh_tokens fails, a
// warning is logged but the account is still considered deleted. Access tokens
// are short-lived JWTs and will naturally expire; refresh tokens are the
// higher-risk surface, so we make every attempt to remove them.
//
// Error catalogue:
//
//	ErrUserNotFound — no user row exists for the given ID
func (uc *UserUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {

	// ── 1. Soft-delete the user row ───────────────────────────────────────────
	if err := uc.userRepo.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}

	// ── 2. Revoke all refresh tokens (best-effort) ────────────────────────────
	// A failure here must not roll back the account deletion — the account is
	// already marked deleted, so logins are blocked regardless. We log a warning
	// so the ops team can investigate any lingering tokens via the dashboard.
	if err := uc.authRepo.DeleteAllRefreshTokensByUserID(ctx, userID); err != nil {
		slog.Warn("delete_account: failed to revoke refresh tokens",
			"user_id", userID,
			"error", err,
		)
	}

	slog.Info("user account deleted", "user_id", userID)
	return nil
}
