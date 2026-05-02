package domain

import "errors"

var (
	ErrEmptyEmail            = errors.New("email cannot be empty")
	ErrEmptyFullName         = errors.New("full name cannot be empty")
	ErrInvalidRole           = errors.New("invalid user role")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrInvalidStatus         = errors.New("invalid user status")
	ErrInvalidFullNameLength = errors.New("full name must be between 3 and 100 characters")
	ErrDuplicateEmail        = errors.New("email already in use")
)
