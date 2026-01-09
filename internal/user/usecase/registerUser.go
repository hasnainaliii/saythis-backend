package usecase

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"saythis-backend/internal/user/domain"
	"saythis-backend/internal/user/repository"
	"time"
)

type UserUseCase struct {
	repo   repository.UserRepository
	logger *log.Logger
}

func NewUserUseCase(repo repository.UserRepository, logger *log.Logger) *UserUseCase {
	return &UserUseCase{repo: repo, logger: logger}
}

func (uc *UserUseCase) RegisterUser(ctx context.Context, email, fullName string) (*domain.User, error) {
	uc.logger.Println("user registration attempt", "email", email)
	if email == "" || fullName == "" {
		return nil, domain.ErrEmptyEmail
	}

	if !isValidEmail(email) {
		return nil, domain.ErrInvalidEmail
	}

	// existingUser, err := uc.repo.FindByEmail(ctx, email)
	// if err == nil && existingUser != nil {
	// 	return nil, ErrEmailAlreadyExists
	// }

	role := domain.RoleUser
	timeNow := time.Now()

	user, err := domain.NewUser(email, fullName, role, timeNow)
	if err != nil {
		uc.logger.Fatalf("user registration failed", "email", email, "error", err)
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	err = uc.repo.Create(ctx, user)

	if err != nil {
		uc.logger.Fatalf("user registration failed", "email", email, "error", err)
		return nil, err
	}

	return user, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
