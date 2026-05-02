package usecase

import (
	"context"
	"fmt"
	"saythis-backend/internal/user/domain"
	"saythis-backend/internal/user/repository"
	"time"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, email, fullName string) (*domain.User, error) {
	user, err := domain.NewUser(email, fullName, domain.RoleUser, time.Now().UTC())

	if err != nil {
		return nil, err
	}

	if err = uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
