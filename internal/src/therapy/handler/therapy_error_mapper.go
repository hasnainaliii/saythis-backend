package handler

import (
	"errors"
	"net/http"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

// mapTherapyError translates therapy domain sentinel errors into HTTP status
// codes and human-readable messages. Add new cases here whenever a new therapy
// error is introduced — never leak raw error strings to the client.
func mapTherapyError(err error) (int, string) {
	switch {

	// ── 400 Bad Request ───────────────────────────────────────────────────────
	case errors.Is(err, therapydomain.ErrInvalidChapterID),
		errors.Is(err, therapydomain.ErrInvalidExerciseID),
		errors.Is(err, therapydomain.ErrInvalidRating):
		return http.StatusBadRequest, err.Error()

	// ── 500 Internal Server Error ─────────────────────────────────────────────
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
