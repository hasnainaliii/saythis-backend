package domain

import "errors"

var (
	ErrEmptyEmail            = errors.New("email cannot be empty")
	ErrEmptyFullName         = errors.New("full name cannot be empty")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrInvalidRole           = errors.New("invalid user role")
	ErrInvalidStatus         = errors.New("invalid user status")
	ErrInvalidFullNameLength = errors.New("full name must be between 3 and 100 characters")

	ErrDuplicateEmail = errors.New("email already in use")
	ErrUserNotFound   = errors.New("user not found")
)
