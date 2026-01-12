package usecase

import (
	"context"
	"errors"
	"log"
	"regexp"
	"saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterAuthUseCase struct {
	repo   repository.AuthRepository
	logger *log.Logger
}

func NewRegisterAuthUseCase(repo repository.AuthRepository, logger *log.Logger) *RegisterAuthUseCase {
	logger.Printf("[DEBUG] Created RegisterAuthUseCase with repo address: %p", repo)
	return &RegisterAuthUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *RegisterAuthUseCase) Execute(ctx context.Context, userID uuid.UUID, password string) error {
	uc.logger.Printf("[DEBUG] RegisterAuthUseCase.Execute started for UserID: %s", userID)

	if err := validatePasswordComplexity(password); err != nil {
		uc.logger.Printf("[WARN] Password complexity check failed: %v", err)
		return err
	}
	uc.logger.Println("[DEBUG] Password complexity check passed")

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		uc.logger.Printf("[ERROR] Failed to hash password: %v", err)
		return err
	}
	uc.logger.Println("[DEBUG] Password hashed successfully using bcrypt")

	now := time.Now()
	creds, err := domain.NewCredentials(
		uuid.New(),
		userID,
		string(hashedPassword),
		now,
	)
	if err != nil {
		uc.logger.Printf("[ERROR] Failed to create Credentials domain object: %v", err)
		return err
	}
	uc.logger.Printf("[DEBUG] Domain.Credentials object created: ID=%s, UserID=%s", creds.ID(), creds.UserID())

	uc.logger.Println("[DEBUG] Calling AuthRepository.Register...")
	return uc.repo.Register(ctx, creds)
}

func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (uc *RegisterAuthUseCase) WithRepository(repo repository.AuthRepository) *RegisterAuthUseCase {
	uc.logger.Printf("[DEBUG] RegisterAuthUseCase: Using transactional repository at: %p", repo)
	return &RegisterAuthUseCase{
		repo:   repo,
		logger: uc.logger,
	}
}
