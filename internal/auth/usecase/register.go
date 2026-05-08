// Package usecase implements the authentication use-case layer.
// Each file in this package covers one distinct authentication flow:
//
//   - register.go  — new account creation (this file)
//   - login.go     — credential verification and session creation
//   - refresh.go   — access-token rotation via a valid refresh token
//
// All use cases share the AuthUseCase struct and its private helpers defined
// in this file. Methods on that struct live in their respective files.
package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"saythis-backend/internal/auth"
	authdomain "saythis-backend/internal/auth/domain"
	authrepo "saythis-backend/internal/auth/repository"
	userdomain "saythis-backend/internal/user/domain"
	userrepo "saythis-backend/internal/user/repository"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // bcrypt silently truncates beyond this — enforce it explicitly
)

// AuthUseCase is the central orchestrator for all authentication flows.
// It coordinates the user repository, auth repository, and JWT configuration
// but owns no database connections itself (those live in the repositories).
type AuthUseCase struct {
	authRepo authrepo.AuthRepository
	userRepo userrepo.UserRepository
	jwtCfg   auth.JWTConfig
}

// NewAuthUseCase constructs an AuthUseCase. Call this once at startup and
// share the result across handlers — it is safe for concurrent use.
func NewAuthUseCase(
	authRepo authrepo.AuthRepository,
	userRepo userrepo.UserRepository,
	jwtCfg auth.JWTConfig,
) *AuthUseCase {
	return &AuthUseCase{
		authRepo: authRepo,
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

// Register creates a new user account, hashes the password, persists both
// the user and credentials atomically in a single transaction, then issues
// a fresh token pair so the caller is immediately authenticated.
//
// Validation order:
//  1. Password strength (empty / too short / too long)
//  2. Email format + full-name format (delegated to userdomain.NewUser)
//  3. Uniqueness (enforced by the DB; returns userdomain.ErrDuplicateEmail on conflict)
func (uc *AuthUseCase) Register(ctx context.Context, email, fullName, password string) (*userdomain.User, authdomain.TokenPair, error) {

	// ── 1. Password validation ────────────────────────────────────────────────
	if strings.TrimSpace(password) == "" {
		return nil, authdomain.TokenPair{}, authdomain.ErrEmptyPassword
	}
	if len(password) < minPasswordLength {
		return nil, authdomain.TokenPair{}, authdomain.ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return nil, authdomain.TokenPair{}, authdomain.ErrPasswordTooLong
	}

	// ── 2. Build domain objects ───────────────────────────────────────────────
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

	// ── 3. Persist user + credentials atomically ──────────────────────────────
	if err = uc.authRepo.Register(ctx, user, creds); err != nil {
		return nil, authdomain.TokenPair{}, fmt.Errorf("register: %w", err)
	}

	// ── 4. Issue token pair ───────────────────────────────────────────────────
	tokens, err := uc.issueTokenPair(ctx, user)
	if err != nil {
		return nil, authdomain.TokenPair{}, err
	}

	return user, tokens, nil
}

// ── Shared private helpers ────────────────────────────────────────────────────

// issueTokenPair signs a new JWT access token and generates + persists a new
// refresh token for the given user. It is called by Register, Login, and Refresh.
//
// The plaintext refresh token is returned to the caller (sent to the client);
// only its SHA-256 hash is stored in the database.
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
