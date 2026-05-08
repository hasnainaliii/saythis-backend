package domain

import "errors"

var (
	// ── Password validation errors ──────────────────────────────────────────
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password cannot exceed 72 characters")

	// ── Token errors ────────────────────────────────────────────────────────
	ErrInvalidToken  = errors.New("invalid or malformed token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrTokenNotFound = errors.New("token not found")

	// ── Login / credential errors ────────────────────────────────────────────
	// ErrInvalidCredentials is intentionally generic — returned for both
	// "email not found" and "wrong password" so callers cannot enumerate accounts.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrAccountSuspended is surfaced when a user exists but has been suspended
	// by an administrator. Unlike deleted accounts, the user can appeal.
	ErrAccountSuspended = errors.New("account has been suspended")

	// ErrAccountLocked is returned when too many consecutive failed login attempts
	// have triggered a temporary lockout. The client should display the lock
	// expiry time so the user knows when to retry.
	ErrAccountLocked = errors.New("account is temporarily locked — too many failed attempts")

	// ErrCredentialsNotFound is a repository-level sentinel returned when no
	// auth_credentials row exists for a given user_id. Should not normally be
	// returned to the client; callers map it to ErrInvalidCredentials.
	ErrCredentialsNotFound = errors.New("credentials not found")
)
