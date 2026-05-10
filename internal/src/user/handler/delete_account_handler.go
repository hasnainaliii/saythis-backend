package handler

import (
	"net/http"

	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/user/usecase"
	"saythis-backend/internal/helper"
)

// DeleteAccountHandler handles DELETE /api/v1/users/me.
// The route must be protected by BearerAuth middleware — the handler reads the
// authenticated user's ID directly from the JWT claims in the request context.
type DeleteAccountHandler struct {
	usecase *usecase.UserUseCase
}

func NewDeleteAccountHandler(uc *usecase.UserUseCase) *DeleteAccountHandler {
	return &DeleteAccountHandler{usecase: uc}
}

// ServeHTTP soft-deletes the account of the currently authenticated user and
// revokes all of their active refresh tokens. Responds with 204 No Content on
// success — there is no body to return for a deletion.
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
