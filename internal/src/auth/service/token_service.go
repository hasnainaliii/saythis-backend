package service

import (
	"fmt"
	"saythis-backend/internal/config"
	"saythis-backend/internal/src/auth/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		secret:             []byte(cfg.JWTSecret),
		accessTokenExpiry:  cfg.AccessTokenExpiry,
		refreshTokenExpiry: cfg.RefreshTokenExpiry,
	}
}

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

func (s *TokenService) GenerateTokenPair(claims domain.Claims) (*domain.TokenPair, error) {
	accessToken, accessExp, err := s.generateToken(claims, "access", s.accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.generateToken(claims, "refresh", s.refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
	}, nil
}

func (s *TokenService) generateToken(claims domain.Claims, tokenType string, expiry time.Duration) (string, int64, error) {
	now := time.Now()
	exp := now.Add(expiry)

	jwtClaims := JWTClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "saythis-backend",
			Subject:   claims.UserID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}

	return signedToken, exp.Unix(), nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*domain.Claims, error) {
	return s.validateToken(tokenString, "access")
}

func (s *TokenService) ValidateRefreshToken(tokenString string) (*domain.Claims, error) {
	return s.validateToken(tokenString, "refresh")
}

func (s *TokenService) validateToken(tokenString string, expectedType string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.Type != expectedType {
		return nil, domain.ErrInvalidToken
	}

	return &domain.Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
