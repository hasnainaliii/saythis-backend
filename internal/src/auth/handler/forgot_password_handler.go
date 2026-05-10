package handler

import (
	"net/http"

	"saythis-backend/internal/src/auth/usecase"
	"saythis-backend/internal/helper"
)

// ForgotPasswordHandler handles POST /api/v1/auth/forgot-password.
// Always responds with 200 OK regardless of whether the email exists —
// this prevents user enumeration.
type ForgotPasswordHandler struct {
	usecase *usecase.AuthUseCase
}

func NewForgotPasswordHandler(uc *usecase.AuthUseCase) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{usecase: uc}
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req forgotPasswordRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// ForgotPassword always returns nil — any internal error is logged, not surfaced.
	_ = h.usecase.ForgotPassword(r.Context(), req.Email)

	helper.JSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a password reset link has been sent",
	})
}
