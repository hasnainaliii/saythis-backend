package handler

import (
	"errors"
	"net/http"

	userdomain "saythis-backend/internal/src/user/domain"
)

// mapUserError translates user domain sentinel errors into HTTP status codes and
// human-readable messages. Add new cases here whenever a new user error is
// introduced — never leak raw error strings to the client.
func mapUserError(err error) (int, string) {
	switch {

	// ── 400 Bad Request ───────────────────────────────────────────────────────
	// Validation errors caused by malformed or incomplete input.
	case errors.Is(err, userdomain.ErrEmptyFullName),
		errors.Is(err, userdomain.ErrInvalidFullNameLength):
		return http.StatusBadRequest, err.Error()

	// ── 404 Not Found ─────────────────────────────────────────────────────────
	case errors.Is(err, userdomain.ErrUserNotFound):
		return http.StatusNotFound, "user not found"

	// ── 500 Internal Server Error ─────────────────────────────────────────────
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
