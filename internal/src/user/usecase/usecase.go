// Package usecase implements the user use-case layer.
// Each file in this package covers one distinct user operation:
//
//   - delete_account.go — soft-deletes the authenticated user's account
//   - update_profile.go — updates mutable profile fields (e.g. full name)
//
// All use cases share the UserUseCase struct defined here.
package usecase

import (
	authrepo "saythis-backend/internal/src/auth/repository"
	userrepo "saythis-backend/internal/src/user/repository"
)

// UserUseCase orchestrates all user-profile operations.
// It holds both the user repository (for profile data) and the auth repository
// (for session management — e.g. revoking tokens on account deletion).
type UserUseCase struct {
	userRepo userrepo.UserRepository
	authRepo authrepo.AuthRepository
}

// NewUserUseCase constructs a UserUseCase. Call this once at startup and share
// the result across handlers — it is safe for concurrent use.
func NewUserUseCase(
	userRepo userrepo.UserRepository,
	authRepo authrepo.AuthRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}
