package handler

import (
	"errors"
	"net/http"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

func mapTherapyError(err error) (int, string) {
	switch {

	case errors.Is(err, therapydomain.ErrInvalidChapterID),
		errors.Is(err, therapydomain.ErrInvalidExerciseID),
		errors.Is(err, therapydomain.ErrInvalidRating):
		return http.StatusBadRequest, err.Error()

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
