package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"saythis-backend/internal/src/auth/repository"
	"saythis-backend/internal/src/auth/service"
	"time"

	"go.uber.org/zap"
)

const (
	TokenExpiryMinutes = 20
	TokenLength        = 32
)

type ForgotPasswordUseCase struct {
	authRepo     repository.AuthRepository
	emailService *service.EmailService
}

func NewForgotPasswordUseCase(authRepo repository.AuthRepository, emailService *service.EmailService) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		authRepo:     authRepo,
		emailService: emailService,
	}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, email string) error {

	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		zap.S().Errorw("Failed to lookup user for password reset", "email", email, "error", err)
		return err
	}

	// If user not found, return success anyway (security: don't reveal if email exists)
	if user == nil {
		zap.S().Infow("Password reset requested for non-existent email", "email", email)
		return nil
	}

	token, err := generateSecureToken(TokenLength)
	if err != nil {
		zap.S().Errorw("Failed to generate password reset token", "error", err)
		return err
	}

	expiresAt := time.Now().Add(TokenExpiryMinutes * time.Minute)

	if err := uc.authRepo.CreatePasswordResetToken(ctx, user.ID.String(), token, expiresAt); err != nil {
		zap.S().Errorw("Failed to store password reset token", "user_id", user.ID, "error", err)
		return err
	}

	if err := uc.emailService.SendPasswordResetEmail(email, token); err != nil {
		zap.S().Errorw("Failed to send password reset email", "email", email, "error", err)
		return err
	}

	zap.S().Infow("Password reset email sent", "email", email, "user_id", user.ID)
	return nil
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
