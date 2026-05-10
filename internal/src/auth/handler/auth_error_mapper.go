package handler

import (
	"errors"
	"net/http"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

// mapAuthError translates domain sentinel errors into HTTP status codes and
// human-readable messages. Add new cases here whenever a new auth error is
// introduced — never leak raw error strings to the client.
func mapAuthError(err error) (int, string) {
	switch {

	// ── 400 Bad Request ───────────────────────────────────────────────────────
	// Validation errors caused by malformed or incomplete input.
	case errors.Is(err, userdomain.ErrEmptyEmail),
		errors.Is(err, userdomain.ErrInvalidEmail),
		errors.Is(err, userdomain.ErrEmptyFullName),
		errors.Is(err, userdomain.ErrInvalidFullNameLength),
		errors.Is(err, userdomain.ErrInvalidRole),
		errors.Is(err, authdomain.ErrEmptyPassword),
		errors.Is(err, authdomain.ErrPasswordTooShort),
		errors.Is(err, authdomain.ErrPasswordTooLong),
		errors.Is(err, authdomain.ErrInvalidToken):
		return http.StatusBadRequest, err.Error()

	// ── 401 Unauthorized ──────────────────────────────────────────────────────
	// Authentication failed — do NOT reveal whether the email exists.
	case errors.Is(err, authdomain.ErrInvalidCredentials),
		errors.Is(err, authdomain.ErrTokenNotFound),
		errors.Is(err, authdomain.ErrExpiredToken):
		return http.StatusUnauthorized, err.Error()

	// ── 403 Forbidden ─────────────────────────────────────────────────────────
	// The identity is known but access has been explicitly denied.
	case errors.Is(err, authdomain.ErrAccountSuspended),
		errors.Is(err, authdomain.ErrAccountLocked):
		return http.StatusForbidden, err.Error()

	// ── 409 Conflict ──────────────────────────────────────────────────────────
	case errors.Is(err, userdomain.ErrDuplicateEmail):
		return http.StatusConflict, userdomain.ErrDuplicateEmail.Error()

	case errors.Is(err, authdomain.ErrEmailAlreadyVerified):
		return http.StatusConflict, authdomain.ErrEmailAlreadyVerified.Error()

	// ── 500 Internal Server Error ─────────────────────────────────────────────
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
