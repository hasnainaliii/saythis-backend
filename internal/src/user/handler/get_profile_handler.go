package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/user/usecase"
)

type GetProfileHandler struct {
	usecase *usecase.UserUseCase
}

func NewGetProfileHandler(uc *usecase.UserUseCase) *GetProfileHandler {
	return &GetProfileHandler{usecase: uc}
}

type getProfileResponse struct {
	User userPayload `json:"user"`
}

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
			ID:              user.ID(),
			Email:           user.Email(),
			FullName:        user.FullName(),
			AvatarURL:       user.AvatarURL(),
			Role:            user.Role(),
			Status:          user.Status(),
			EmailVerifiedAt: user.EmailVerifiedAt(),
			CreatedAt:       user.CreatedAt(),
			UpdatedAt:       user.UpdatedAt(),
		},
	})
}
