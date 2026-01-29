package usecase

import (
	"context"
	"saythis-backend/internal/src/user/domain"
	"saythis-backend/internal/src/user/repository"
	"time"

	"go.uber.org/zap"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) WithRepository(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, email, fullName string) (*domain.User, error) {

	role := domain.RoleUser
	timeNow := time.Now()

	user, err := domain.NewUser(email, fullName, role, timeNow)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		zap.S().Errorw("Failed to create user", "email", email, "error", err)
		return nil, err
	}

	zap.S().Debugw("User created successfully", "email", email, "user_id", user.ID())
	return user, nil
}
