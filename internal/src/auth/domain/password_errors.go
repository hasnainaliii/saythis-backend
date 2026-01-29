package domain

import "errors"

var (
	ErrPasswordTooShort           = errors.New("password must be at least 8 characters")
	ErrPasswordMissingNumber      = errors.New("password must contain at least one number")
	ErrPasswordMissingSpecialChar = errors.New("password must contain at least one special character")
)
