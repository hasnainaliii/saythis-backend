package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

// dummyHash is a pre-computed bcrypt hash used as a constant-time shield.
//
// When a login attempt arrives for an email that does not exist in the database,
// we still call bcrypt.CompareHashAndPassword against this dummy hash before
// returning ErrInvalidCredentials. Without this, an attacker could distinguish
// "email not found" from "wrong password" purely by measuring response latency
// (user-enumeration via timing side-channel).
//
// The hash is generated once at package initialisation, not per-request, so
// there is no meaningful startup overhead beyond a single bcrypt computation.
var dummyHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy-timing-shield-saythis"), bcrypt.DefaultCost)
	if err != nil {
		// This only fails on an unrecoverable system error (e.g. broken random source).
		// Panicking here is intentional: starting without timing protection would be
		// a silent security regression.
		panic("auth: could not pre-compute dummy bcrypt hash for timing protection: " + err.Error())
	}
	dummyHash = h
}

// Login authenticates a user with their email and password, then issues a fresh
// token pair on success.
//
// Security properties:
//   - Constant-time response for "email not found" vs "wrong password" via a dummy
//     bcrypt comparison, preventing user-enumeration through timing attacks.
//   - Deleted accounts are treated identically to non-existent ones, so attackers
//     cannot discover previously deleted accounts.
//   - Suspended accounts surface an explicit error so users understand why access
//     is denied (and can contact support).
//   - Account lockout from repeated failures is checked before password verification.
//   - last_login is updated asynchronously (best-effort) — a failure here does not
//     prevent the user from logging in.
//
// Error catalogue:
//
//	ErrEmptyEmail          — email field was blank
//	ErrEmptyPassword       — password field was blank
//	ErrInvalidCredentials  — email not found, wrong password, or deleted account
//	ErrAccountSuspended    — account exists but is suspended
//	ErrAccountLocked       — too many recent failed attempts
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*userdomain.User, authdomain.TokenPair, error) {

	// ── 1. Input normalisation + basic validation ─────────────────────────────
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, authdomain.TokenPair{}, userdomain.ErrEmptyEmail
	}
	if strings.TrimSpace(password) == "" {
		return nil, authdomain.TokenPair{}, authdomain.ErrEmptyPassword
	}

	// ── 2. Resolve the user by email ──────────────────────────────────────────
	// If the email is not found, we still run a dummy bcrypt compare below to
	// keep the response time indistinguishable from a "wrong password" attempt.
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			// Timing shield: consume the same amount of CPU as a real bcrypt verify.
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
		}
		return nil, authdomain.TokenPair{}, fmt.Errorf("lookup user: %w", err)
	}

	// ── 3. Account-status gate ────────────────────────────────────────────────
	switch user.Status() {
	case userdomain.StatusSuspended:
		// The account exists but access has been revoked by an administrator.
		// Return a distinct error so the client can show a meaningful message.
		return nil, authdomain.TokenPair{}, authdomain.ErrAccountSuspended

	case userdomain.StatusDeleted:
		// Treat deleted accounts identically to non-existent ones.
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
	}

	// ── 4. Load credentials ───────────────────────────────────────────────────
	creds, err := uc.authRepo.FindCredentialsByUserID(ctx, user.ID())
	if err != nil {
		// No credentials row is a data-integrity problem, not a user error.
		// Map it to a generic internal error rather than ErrInvalidCredentials
		// so the ops team can distinguish it in logs.
		return nil, authdomain.TokenPair{}, fmt.Errorf("fetch credentials: %w", err)
	}

	// ── 5. Lockout check ──────────────────────────────────────────────────────
	if creds.IsLocked() {
		return nil, authdomain.TokenPair{}, authdomain.ErrAccountLocked
	}

	// ── 6. Password verification ──────────────────────────────────────────────
	if err = bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash()), []byte(password)); err != nil {
		// bcrypt returns bcrypt.ErrMismatchedHashAndPassword on a bad password;
		// any other error is a system fault. Either way, return the generic sentinel.
		return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
	}

	// ── 7. Record the successful login (best-effort) ──────────────────────────
	now := time.Now().UTC()
	if err = uc.authRepo.UpdateLastLogin(ctx, user.ID(), now); err != nil {
		// Non-fatal: a failure here must not block the user from logging in.
		// Log a warning so the issue surfaces in dashboards without impacting UX.
		slog.Warn("login: failed to update last_login",
			"user_id", user.ID(),
			"error", err,
		)
	}

	// ── 8. Issue a fresh token pair ───────────────────────────────────────────
	tokens, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, authdomain.TokenPair{}, err
	}

	slog.Info("user logged in", "user_id", user.ID(), "email", user.Email())

	return user, tokens, nil
}
