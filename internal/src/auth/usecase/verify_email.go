package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
)

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, plaintextToken string) error {
	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.ErrInvalidToken
	}

	tokenHash := auth.HashToken(plaintextToken)

	token, err := uc.authRepo.FindEmailVerificationToken(ctx, tokenHash)
	if err != nil {
		return authdomain.ErrInvalidToken
	}

	if token.IsExpired() {
		_ = uc.authRepo.DeleteEmailVerificationToken(ctx, tokenHash)
		return authdomain.ErrExpiredToken
	}

	if err = uc.authRepo.MarkEmailVerified(ctx, token.UserID(), time.Now().UTC()); err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}

	if err = uc.authRepo.DeleteEmailVerificationToken(ctx, tokenHash); err != nil {
		_ = err
	}

	return nil
}
