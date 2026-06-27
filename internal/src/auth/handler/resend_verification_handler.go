package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/auth/usecase"
)

type ResendVerificationHandler struct {
	usecase *usecase.AuthUseCase
	jwtCfg  auth.JWTConfig
}

func NewResendVerificationHandler(uc *usecase.AuthUseCase, jwtCfg auth.JWTConfig) *ResendVerificationHandler {
	return &ResendVerificationHandler{usecase: uc, jwtCfg: jwtCfg}
}

func (h *ResendVerificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "missing authentication")
		return
	}

	if err := h.usecase.ResendVerificationEmail(r.Context(), claims.UserID); err != nil {
		status, msg := mapAuthError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, map[string]string{
		"message": "verification email sent — please check your inbox",
	})
}
