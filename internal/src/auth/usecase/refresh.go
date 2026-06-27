package usecase

import (
	"context"
	"fmt"
	"strings"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
)

func (uc *AuthUseCase) Refresh(ctx context.Context, plaintextToken string) (authdomain.TokenPair, error) {
	if strings.TrimSpace(plaintextToken) == "" {
		return authdomain.TokenPair{}, authdomain.ErrInvalidToken
	}

	tokenHash := auth.HashRefreshToken(plaintextToken)

	stored, err := uc.authRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return authdomain.TokenPair{}, authdomain.ErrTokenNotFound
	}

	if stored.IsExpired() {
		_ = uc.authRepo.DeleteRefreshToken(ctx, tokenHash)
		return authdomain.TokenPair{}, authdomain.ErrExpiredToken
	}

	if err = uc.authRepo.DeleteRefreshToken(ctx, tokenHash); err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("rotate token: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, stored.UserID())
	if err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("get user for refresh: %w", err)
	}

	return uc.issueTokenPair(ctx, user)
}
