package handler

import (
	"errors"
	"net/http"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

func mapAuthError(err error) (int, string) {
	switch {

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

	case errors.Is(err, authdomain.ErrInvalidCredentials),
		errors.Is(err, authdomain.ErrTokenNotFound),
		errors.Is(err, authdomain.ErrExpiredToken):
		return http.StatusUnauthorized, err.Error()

	case errors.Is(err, authdomain.ErrAccountSuspended),
		errors.Is(err, authdomain.ErrAccountLocked):
		return http.StatusForbidden, err.Error()

	case errors.Is(err, userdomain.ErrDuplicateEmail):
		return http.StatusConflict, userdomain.ErrDuplicateEmail.Error()

	case errors.Is(err, authdomain.ErrEmailAlreadyVerified):
		return http.StatusConflict, authdomain.ErrEmailAlreadyVerified.Error()

	case errors.Is(err, authdomain.ErrResendTooSoon):
		return http.StatusTooManyRequests, authdomain.ErrResendTooSoon.Error()

	case errors.Is(err, userdomain.ErrUserNotFound):
		return http.StatusNotFound, userdomain.ErrUserNotFound.Error()

	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
