package domain

import "errors"

var (
	ErrEmptyPasswordHash  = errors.New("password hash cannot be empty")
	ErrAccountLocked      = errors.New("account is locked")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)
