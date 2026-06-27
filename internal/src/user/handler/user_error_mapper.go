package handler

import (
	"errors"
	"net/http"

	userdomain "saythis-backend/internal/src/user/domain"
)

func mapUserError(err error) (int, string) {
	switch {

	case errors.Is(err, userdomain.ErrEmptyFullName),
		errors.Is(err, userdomain.ErrInvalidFullNameLength):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, userdomain.ErrUserNotFound):
		return http.StatusNotFound, "user not found"

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
