package domain

import "errors"

var (
	ErrEmptyPasswordHash = errors.New("password hash cannot be empty")
	ErrAccountLocked     = errors.New("account is locked")
)
