package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth/usecase"
)

type VerifyEmailHandler struct {
	usecase *usecase.AuthUseCase
}

func NewVerifyEmailHandler(uc *usecase.AuthUseCase) *VerifyEmailHandler {
	return &VerifyEmailHandler{usecase: uc}
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *VerifyEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req verifyEmailRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.usecase.VerifyEmail(r.Context(), req.Token); err != nil {
		status, msg := mapAuthError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, map[string]string{
		"message": "email verified successfully",
	})
}
