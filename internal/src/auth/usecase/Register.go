package usecase

import (
	"context"
	"regexp"
	"saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type RegisterAuthUseCase struct {
	repo repository.AuthRepository
}

func NewRegisterAuthUseCase(repo repository.AuthRepository) *RegisterAuthUseCase {
	return &RegisterAuthUseCase{
		repo: repo,
	}
}

func (uc *RegisterAuthUseCase) Execute(ctx context.Context, userID uuid.UUID, password string) error {

	if err := validatePasswordComplexity(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		zap.S().Errorw("Failed to hash password", "user_id", userID, "error", err)
		return err
	}

	now := time.Now()
	creds, err := domain.NewCredentials(
		uuid.New(),
		userID,
		string(hashedPassword),
		now,
	)
	if err != nil {
		return err
	}

	if err := uc.repo.Register(ctx, creds); err != nil {
		zap.S().Errorw("Failed to save credentials", "user_id", userID, "error", err)
		return err
	}

	zap.S().Debugw("Credentials created successfully", "user_id", userID)
	return nil
}

func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return domain.ErrPasswordTooShort
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return domain.ErrPasswordMissingNumber
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	if !hasSpecial {
		return domain.ErrPasswordMissingSpecialChar
	}

	return nil
}

func (uc *RegisterAuthUseCase) WithRepository(repo repository.AuthRepository) *RegisterAuthUseCase {
	return &RegisterAuthUseCase{
		repo: repo,
	}
}
