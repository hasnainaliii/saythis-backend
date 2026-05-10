package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"saythis-backend/internal/src/auth/usecase"
	userdomain "saythis-backend/internal/src/user/domain"
	"saythis-backend/internal/helper"
)

type RegisterHandler struct {
	usecase *usecase.AuthUseCase
}

func NewRegisterHandler(uc *usecase.AuthUseCase) *RegisterHandler {
	return &RegisterHandler{usecase: uc}
}

type registerRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

type registerResponse struct {
	User         userPayload `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

type userPayload struct {
	ID        uuid.UUID             `json:"id"`
	Email     string                `json:"email"`
	FullName  string                `json:"full_name"`
	Role      userdomain.UserRole   `json:"role"`
	Status    userdomain.UserStatus `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req registerRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.usecase.Register(r.Context(), req.Email, req.FullName, req.Password)
	if err != nil {
		status, msg := mapAuthError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusCreated, registerResponse{
		User: userPayload{
			ID:        user.ID(),
			Email:     user.Email(),
			FullName:  user.FullName(),
			Role:      user.Role(),
			Status:    user.Status(),
			CreatedAt: user.CreatedAt(),
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
