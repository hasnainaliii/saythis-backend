package handler

import (
	"encoding/json"
	"net/http"
	"saythis-backend/internal/src/auth/usecase"
	"strings"

	"go.uber.org/zap"
)

type ForgotPasswordHandler struct {
	forgotPasswordUC *usecase.ForgotPasswordUseCase
}

func NewForgotPasswordHandler(forgotPasswordUC *usecase.ForgotPasswordUseCase) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		forgotPasswordUC: forgotPasswordUC,
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		h.respondWithError(w, "VALIDATION_ERROR", "Email is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.forgotPasswordUC.Execute(ctx, email); err != nil {
		// Log error but don't expose it to user (security)
		zap.S().Errorw("Forgot password failed", "error", err)
	}

	// Always return success to prevent email enumeration
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ForgotPasswordResponse{
		Message: "If this email exists in our system, a password reset link has been sent",
	})
}

func (h *ForgotPasswordHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
