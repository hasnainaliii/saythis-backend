package usecase

import (
	authrepo "saythis-backend/internal/src/auth/repository"
	userrepo "saythis-backend/internal/src/user/repository"
)

type UserUseCase struct {
	userRepo userrepo.UserRepository
	authRepo authrepo.AuthRepository
	uploader ImageUploader
}

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
