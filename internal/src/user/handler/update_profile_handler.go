package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	userdomain "saythis-backend/internal/src/user/domain"
	"saythis-backend/internal/src/user/usecase"
)

// UpdateProfileHandler handles PATCH /api/v1/users/me.
// The route must be protected by BearerAuth middleware — the handler reads the
// authenticated user's ID directly from the JWT claims in the request context.
type UpdateProfileHandler struct {
	usecase *usecase.UserUseCase
}

func NewUpdateProfileHandler(uc *usecase.UserUseCase) *UpdateProfileHandler {
	return &UpdateProfileHandler{usecase: uc}
}

type updateProfileRequest struct {
	FullName string `json:"full_name"`
}

type updateProfileResponse struct {
	User userPayload `json:"user"`
}

// userPayload is the user object returned in every user-handler response.
// It mirrors the auth handler's payload but includes UpdatedAt so callers
// can confirm the change without a separate GET.
type userPayload struct {
	ID        uuid.UUID             `json:"id"`
	Email     string                `json:"email"`
	FullName  string                `json:"full_name"`
	AvatarURL string                `json:"avatar_url"`
	Role      userdomain.UserRole   `json:"role"`
	Status    userdomain.UserStatus `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// ServeHTTP validates the request body, updates the user's profile, and returns
// the full updated user object so the client can refresh its local state in one
// round-trip.
func (h *UpdateProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.usecase.UpdateProfile(r.Context(), claims.UserID, req.FullName)
	if err != nil {
		status, msg := mapUserError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, updateProfileResponse{
		User: userPayload{
			ID:        user.ID(),
			Email:     user.Email(),
			FullName:  user.FullName(),
			AvatarURL: user.AvatarURL(),
			Role:      user.Role(),
			Status:    user.Status(),
			CreatedAt: user.CreatedAt(),
			UpdatedAt: user.UpdatedAt(),
		},
	})
}
