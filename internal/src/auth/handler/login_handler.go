package handler

import (
	"net/http"

	"saythis-backend/internal/src/auth/usecase"
	"saythis-backend/internal/helper"
)

type LoginHandler struct {
	usecase *usecase.AuthUseCase
}

func NewLoginHandler(uc *usecase.AuthUseCase) *LoginHandler {
	return &LoginHandler{usecase: uc}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	User         userPayload `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req loginRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		status, msg := mapAuthError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, loginResponse{
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
