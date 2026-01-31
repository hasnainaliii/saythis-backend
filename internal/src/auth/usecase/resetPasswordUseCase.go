package usecase

import (
	"context"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/src/auth/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordUseCase struct {
	authRepo repository.AuthRepository
}

func NewResetPasswordUseCase(authRepo repository.AuthRepository) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		authRepo: authRepo,
	}
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) error {
	// Validate password length
	if len(input.NewPassword) < 8 {
		return apperror.New("VALIDATION_ERROR", "Password must be at least 8 characters", 400)
	}

	// Get the token from database
	resetToken, err := uc.authRepo.GetPasswordResetToken(ctx, input.Token)
	if err != nil {
		return err
	}

	// Check if token is expired
	if resetToken.IsExpired() {
		return apperror.New("TOKEN_EXPIRED", "Password reset token has expired", 400)
	}

	// Check if token is already used
	if resetToken.IsUsed() {
		return apperror.New("TOKEN_USED", "Password reset token has already been used", 400)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		zap.S().Errorw("Failed to hash new password", "error", err)
		return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to process password", 500)
	}

	// Update the password
	if err := uc.authRepo.UpdatePassword(ctx, resetToken.UserID.String(), string(hashedPassword)); err != nil {
		return err
	}

	// Mark token as used
	if err := uc.authRepo.MarkTokenAsUsed(ctx, input.Token); err != nil {
		zap.S().Errorw("Failed to mark token as used", "error", err)
		// Don't return error, password was already updated
	}

	zap.S().Infow("Password reset successful", "user_id", resetToken.UserID)
	return nil
}
