package usecase

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"saythis-backend/internal/src/user/domain"
	"saythis-backend/internal/src/user/repository"

	"time"
)

type UserUseCase struct {
	repo   repository.UserRepository
	logger *log.Logger
}

func NewUserUseCase(repo repository.UserRepository, logger *log.Logger) *UserUseCase {
	logger.Printf("[DEBUG] Created UserUseCase with repo address: %p", repo)
	return &UserUseCase{repo: repo, logger: logger}
}

func (uc *UserUseCase) WithRepository(repo repository.UserRepository) *UserUseCase {
	uc.logger.Printf("[DEBUG] UserUseCase: Using transactional repository at: %p", repo)
	return &UserUseCase{
		repo:   repo,
		logger: uc.logger,
	}
}

func (uc *UserUseCase) RegisterUser(ctx context.Context, email, fullName string) (*domain.User, error) {
	uc.logger.Printf("[DEBUG] UserUseCase.RegisterUser started. Email: %s, FullName: %s", email, fullName)

	if email == "" || fullName == "" {
		uc.logger.Printf("[WARN] Validation failed: Empty email or full name")
		return nil, domain.ErrEmptyEmail
	}

	if !isValidEmail(email) {
		uc.logger.Printf("[WARN] Validation failed: Invalid email format: %s", email)
		return nil, domain.ErrInvalidEmail
	}
	uc.logger.Println("[DEBUG] Basic validation passed")

	role := domain.RoleUser
	timeNow := time.Now()

	uc.logger.Println("[DEBUG] Creating Domain.User object...")
	user, err := domain.NewUser(email, fullName, role, timeNow)
	if err != nil {
		uc.logger.Printf("[ERROR] Domain.User creation failed: %v", err)
		return nil, fmt.Errorf("user creation failed: %w", err)
	}
	uc.logger.Printf("[DEBUG] Domain.User object created: ID=%s, Email=%s, Role=%s", user.ID(), user.Email(), user.Role())

	uc.logger.Println("[DEBUG] Calling UserRepository.Create...")
	err = uc.repo.Create(ctx, user)
	if err != nil {
		uc.logger.Printf("[ERROR] UserRepository.Create failed: %v", err)
		return nil, err
	}

	uc.logger.Println("[DEBUG] UserUseCase.RegisterUser finished successfully")
	return user, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
