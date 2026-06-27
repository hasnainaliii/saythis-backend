package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/user/usecase"
)

type DeleteAccountHandler struct {
	usecase *usecase.UserUseCase
}

func NewDeleteAccountHandler(uc *usecase.UserUseCase) *DeleteAccountHandler {
	return &DeleteAccountHandler{usecase: uc}
}

func (h *DeleteAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.usecase.DeleteAccount(r.Context(), claims.UserID); err != nil {
		status, msg := mapUserError(err)
		helper.Error(w, status, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
