package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"saythis-backend/internal/config"
)

// JWTConfig holds the secret key and token lifetimes.
// Built once at startup from environment variables and passed down to AuthUseCase.
type JWTConfig struct {
	Secret          []byte
	AccessTokenTTL  time.Duration // recommended: 15 minutes
	RefreshTokenTTL time.Duration // recommended: 7 days
}

// NewJWTConfig builds a JWTConfig from the application config.
// Call this directly at the use-site — no intermediate variable needed.
func NewJWTConfig(cfg *config.Config) JWTConfig {
	return JWTConfig{
		Secret:          []byte(cfg.JWTSecret),
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

// Claims are the fields embedded inside every signed access token.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken signs a new JWT containing the user's identity.
func GenerateAccessToken(cfg JWTConfig, userID uuid.UUID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(cfg.Secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken parses and validates a JWT string, returning its claims on success.
func ValidateAccessToken(cfg JWTConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return cfg.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// GenerateRefreshToken creates a cryptographically random 32-byte token.
// Returns both the plaintext (sent to client, never stored) and the SHA-256
// hash (stored in the DB so a breach cannot be used to replay tokens).
func GenerateRefreshToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	plaintext = base64.URLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// HashRefreshToken returns the SHA-256 hash of a plaintext token.
// Used when the client sends a token back — we hash it before looking it up in the DB.
func HashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// GenerateSecureToken creates a cryptographically random one-time token.
// Returns the plaintext (sent to the user, never stored) and its SHA-256 hash
// (stored in the DB). Used for email verification and password reset tokens.
func GenerateSecureToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	plaintext = base64.URLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// HashToken returns the SHA-256 hex digest of a plaintext token.
// Used when the client presents a one-time token — hash it before querying the DB.
func HashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
