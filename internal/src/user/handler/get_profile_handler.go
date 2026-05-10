package handler

import (
	"net/http"

	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/user/usecase"
	"saythis-backend/internal/helper"
)

// GetProfileHandler handles GET /api/v1/users/me.
// The route must be protected by BearerAuth middleware — the handler reads the
// authenticated user's ID directly from the JWT claims in the request context.
type GetProfileHandler struct {
	usecase *usecase.UserUseCase
}

func NewGetProfileHandler(uc *usecase.UserUseCase) *GetProfileHandler {
	return &GetProfileHandler{usecase: uc}
}

type getProfileResponse struct {
	User userPayload `json:"user"`
}

// ServeHTTP fetches and returns the authenticated user's profile data.
func (h *GetProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.usecase.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		status, msg := mapUserError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, getProfileResponse{
		User: userPayload{
			ID:        user.ID(),
			Email:     user.Email(),
			FullName:  user.FullName(),
			Role:      user.Role(),
			Status:    user.Status(),
			CreatedAt: user.CreatedAt(),
			UpdatedAt: user.UpdatedAt(),
		},
	})
}
