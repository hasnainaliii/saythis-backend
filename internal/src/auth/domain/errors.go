package domain

import "errors"

var (
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password cannot exceed 72 characters")

	ErrInvalidToken  = errors.New("invalid or malformed token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrTokenNotFound = errors.New("token not found")

	ErrEmailAlreadyVerified = errors.New("email address is already verified")
	ErrResendTooSoon        = errors.New("you can only request a new verification email once every 24 hours")

	ErrInvalidCredentials = errors.New("invalid email or password")

	ErrAccountSuspended = errors.New("account has been suspended")

	ErrAccountLocked = errors.New("account is temporarily locked, too many failed attempts")

	ErrCredentialsNotFound = errors.New("credentials not found")
)
