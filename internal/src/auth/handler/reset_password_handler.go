package handler

import (
	"net/http"

	"saythis-backend/internal/src/auth/usecase"
	"saythis-backend/internal/helper"
)

// ResetPasswordHandler handles POST /api/v1/auth/reset-password.
// Validates the one-time token, enforces password rules, and updates the credential.
type ResetPasswordHandler struct {
	usecase *usecase.AuthUseCase
}

func NewResetPasswordHandler(uc *usecase.AuthUseCase) *ResetPasswordHandler {
	return &ResetPasswordHandler{usecase: uc}
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req resetPasswordRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.usecase.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		status, msg := mapAuthError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, map[string]string{
		"message": "password has been reset successfully",
	})
}
