package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	authdomain "saythis-backend/internal/src/auth/domain"
	userdomain "saythis-backend/internal/src/user/domain"
)

var dummyHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy-timing-shield-saythis"), bcrypt.DefaultCost)
	if err != nil {
		panic("auth: could not pre-compute dummy bcrypt hash for timing protection: " + err.Error())
	}
	dummyHash = h
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*userdomain.User, authdomain.TokenPair, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, authdomain.TokenPair{}, userdomain.ErrEmptyEmail
	}
	if strings.TrimSpace(password) == "" {
		return nil, authdomain.TokenPair{}, authdomain.ErrEmptyPassword
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
		}
		return nil, authdomain.TokenPair{}, fmt.Errorf("lookup user: %w", err)
	}

	switch user.Status() {
	case userdomain.StatusSuspended:
		return nil, authdomain.TokenPair{}, authdomain.ErrAccountSuspended

	case userdomain.StatusDeleted:
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
	}

	creds, err := uc.authRepo.FindCredentialsByUserID(ctx, user.ID())
	if err != nil {
		return nil, authdomain.TokenPair{}, fmt.Errorf("fetch credentials: %w", err)
	}

	if creds.IsLocked() {
		return nil, authdomain.TokenPair{}, authdomain.ErrAccountLocked
	}

	if err = bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash()), []byte(password)); err != nil {
		if recErr := uc.authRepo.RecordFailedAttempt(ctx, user.ID()); recErr != nil {
			slog.Warn("login: failed to record failed attempt",
				"user_id", user.ID(),
				"error", recErr,
			)
		}
		return nil, authdomain.TokenPair{}, authdomain.ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err = uc.authRepo.UpdateLastLogin(ctx, user.ID(), now); err != nil {
		slog.Warn("login: failed to record successful login",
			"user_id", user.ID(),
			"error", err,
		)
	}

	tokens, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, authdomain.TokenPair{}, err
	}

	slog.Info("user logged in", "user_id", user.ID(), "email", user.Email())

	return user, tokens, nil
}
