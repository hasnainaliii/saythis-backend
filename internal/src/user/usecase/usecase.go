// Package usecase implements the user use-case layer.
// Each file in this package covers one distinct user operation:
//
//   - delete_account.go  — soft-deletes the authenticated user's account
//   - get_profile.go     — fetches the authenticated user's profile
//   - update_profile.go  — updates mutable profile fields (full name)
//   - update_avatar.go   — uploads a new avatar and persists the Cloudinary URL
//   - cloudinary.go      — ImageUploader interface + CloudinaryUploader implementation
//
// All use cases share the UserUseCase struct defined here.
package usecase

import (
	authrepo "saythis-backend/internal/src/auth/repository"
	userrepo "saythis-backend/internal/src/user/repository"
)

// UserUseCase orchestrates all user-profile operations.
// It holds the user repository, the auth repository (for session management),
// and an image uploader (for avatar uploads).
type UserUseCase struct {
	userRepo userrepo.UserRepository
	authRepo authrepo.AuthRepository
	uploader ImageUploader
}

// NewUserUseCase constructs a UserUseCase. Call this once at startup and share
// the result across handlers — it is safe for concurrent use.
func NewUserUseCase(
	userRepo userrepo.UserRepository,
	authRepo authrepo.AuthRepository,
	uploader ImageUploader,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		authRepo: authRepo,
		uploader: uploader,
	}
}
