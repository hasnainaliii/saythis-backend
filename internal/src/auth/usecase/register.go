package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"saythis-backend/internal/src/auth"
	authdomain "saythis-backend/internal/src/auth/domain"
	authrepo "saythis-backend/internal/src/auth/repository"
	userdomain "saythis-backend/internal/src/user/domain"
	userrepo "saythis-backend/internal/src/user/repository"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

type AuthUseCase struct {
	authRepo    authrepo.AuthRepository
	userRepo    userrepo.UserRepository
	jwtCfg      auth.JWTConfig
	emailSender auth.EmailSender
	frontendURL string
}

func NewAuthUseCase(
	authRepo authrepo.AuthRepository,
	userRepo userrepo.UserRepository,
	jwtCfg auth.JWTConfig,
	emailSender auth.EmailSender,
	frontendURL string,
) *AuthUseCase {
	return &AuthUseCase{
		authRepo:    authRepo,
		userRepo:    userRepo,
		jwtCfg:      jwtCfg,
		emailSender: emailSender,
		frontendURL: frontendURL,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, email, fullName, password string) (*userdomain.User, authdomain.TokenPair, error) {

	if strings.TrimSpace(password) == "" {
		return nil, authdomain.TokenPair{}, authdomain.ErrEmptyPassword
	}
	if len(password) < minPasswordLength {
		return nil, authdomain.TokenPair{}, authdomain.ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return nil, authdomain.TokenPair{}, authdomain.ErrPasswordTooLong
	}

	timeNow := time.Now().UTC()

	user, err := userdomain.NewUser(email, fullName, userdomain.RoleUser, timeNow)
	if err != nil {
		return nil, authdomain.TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, authdomain.TokenPair{}, fmt.Errorf("hash password: %w", err)
	}

	creds := authdomain.NewAuthCredentials(user.ID(), string(hash), timeNow)

	if err = uc.authRepo.Register(ctx, user, creds); err != nil {
		return nil, authdomain.TokenPair{}, fmt.Errorf("register: %w", err)
	}

	uc.dispatchVerificationEmail(ctx, user.ID(), user.Email())

	tokens, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, authdomain.TokenPair{}, err
	}

	return user, tokens, nil
}

func (uc *AuthUseCase) dispatchVerificationEmail(ctx context.Context, userID uuid.UUID, userEmail string) {
	plaintext, tokenHash, err := auth.GenerateSecureToken()
	if err != nil {
		slog.Error("register: failed to generate verification token",
			"user_id", userID,
			"error", err,
		)
		return
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	verificationToken := authdomain.NewEmailVerificationToken(userID, tokenHash, expiresAt)

	if err = uc.authRepo.SaveEmailVerificationToken(ctx, verificationToken); err != nil {
		slog.Error("register: failed to save verification token",
			"user_id", userID,
			"error", err,
		)
		return
	}

	verificationURL := uc.frontendURL + "/verify-email?token=" + plaintext

	if err = uc.emailSender.SendVerification(ctx, userEmail, verificationURL); err != nil {
		slog.Error("register: failed to send verification email",
			"user_id", userID,
			"error", err,
		)
	}
}

func (uc *AuthUseCase) issueTokenPair(ctx context.Context, user *userdomain.User) (authdomain.TokenPair, error) {
	accessToken, err := auth.GenerateAccessToken(uc.jwtCfg, user.ID(), user.Email(), string(user.Role()))
	if err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	plaintext, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(uc.jwtCfg.RefreshTokenTTL)
	refreshToken := authdomain.NewRefreshToken(user.ID(), hash, expiresAt)

	if err = uc.authRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return authdomain.TokenPair{}, fmt.Errorf("save refresh token: %w", err)
	}

	return authdomain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: plaintext,
	}, nil
}
