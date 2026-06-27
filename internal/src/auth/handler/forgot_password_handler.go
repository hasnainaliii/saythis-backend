package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth/usecase"
)

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

	_ = h.usecase.ForgotPassword(r.Context(), req.Email)

	helper.JSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a password reset link has been sent",
	})
}
