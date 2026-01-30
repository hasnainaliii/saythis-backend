package handler

import (
	"encoding/json"
	"net/http"
	"saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/service"

	"go.uber.org/zap"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type RefreshHandler struct {
	tokenService *service.TokenService
}

func NewRefreshHandler(tokenService *service.TokenService) *RefreshHandler {
	return &RefreshHandler{
		tokenService: tokenService,
	}
}

func (h *RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		h.respondWithError(w, "MISSING_FIELDS", "Refresh token is required", http.StatusBadRequest)
		return
	}

	claims, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		zap.S().Debugw("Refresh token validation failed", "error", err)
		h.respondWithError(w, "INVALID_TOKEN", "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	tokenPair, err := h.tokenService.GenerateTokenPair(*claims)
	if err != nil {
		zap.S().Errorw("Failed to generate new token pair", "error", err)
		h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RefreshHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

var _ = domain.ErrInvalidToken
