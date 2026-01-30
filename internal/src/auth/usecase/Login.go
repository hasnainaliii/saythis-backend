package usecase

import (
	"context"
	"saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/repository"
	"saythis-backend/internal/src/auth/service"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type LoginAuthUseCase struct {
	repo         repository.AuthRepository
	tokenService *service.TokenService
}

func NewLoginAuthUseCase(repo repository.AuthRepository, tokenService *service.TokenService) *LoginAuthUseCase {
	return &LoginAuthUseCase{
		repo:         repo,
		tokenService: tokenService,
	}
}

type LoginResult struct {
	TokenPair *domain.TokenPair
	User      *domain.CredentialsWithUser
}

func (uc *LoginAuthUseCase) Execute(ctx context.Context, email, password string) (*LoginResult, error) {
	creds, err := uc.repo.GetCredentialsWithUser(ctx, email)
	if err != nil {
		zap.S().Debugw("Login failed: user lookup", "email", email, "error", err)
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		zap.S().Debugw("Login failed: password mismatch", "email", email)
		return nil, domain.ErrInvalidCredentials
	}

	claims := domain.Claims{
		UserID: creds.UserID,
		Email:  creds.Email,
		Role:   creds.Role,
	}

	tokenPair, err := uc.tokenService.GenerateTokenPair(claims)
	if err != nil {
		zap.S().Errorw("Failed to generate token pair", "email", email, "error", err)
		return nil, err
	}

	zap.S().Infow("User logged in successfully", "email", email, "user_id", creds.UserID)

	return &LoginResult{
		TokenPair: tokenPair,
		User:      creds,
	}, nil
}
