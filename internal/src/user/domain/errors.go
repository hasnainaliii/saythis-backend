package domain

import "errors"

var (
	ErrEmptyEmail            = errors.New("email cannot be empty")
	ErrEmptyFullName         = errors.New("full name cannot be empty")
	ErrInvalidRole           = errors.New("invalid user role")
	ErrInvalidEmail          = errors.New("invalid Email")
	ErrInvalidStatus         = errors.New("invalid user status")
	ErrInvalidTimezone       = errors.New("invalid timezone")
	ErrInvalidFullNameLength = errors.New("invalid name length")
)
