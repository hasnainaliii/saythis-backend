package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/usecase"

	"go.uber.org/zap"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type LoginHandler struct {
	loginUC *usecase.LoginAuthUseCase
}

func NewLoginHandler(loginUC *usecase.LoginAuthUseCase) *LoginHandler {
	return &LoginHandler{
		loginUC: loginUC,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		h.respondWithError(w, "MISSING_FIELDS", "Email and password are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	result, err := h.loginUC.Execute(ctx, req.Email, req.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := LoginResponse{
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
		ExpiresAt:    result.TokenPair.ExpiresAt,
		User: UserResponse{
			ID:       result.User.UserID.String(),
			Email:    result.User.Email,
			FullName: result.User.FullName,
			Role:     result.User.Role,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *LoginHandler) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrInvalidCredentials) {
		h.respondWithError(w, "INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
		return
	}

	zap.S().Errorw("Login failed", "error", err)
	h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
}

func (h *LoginHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
