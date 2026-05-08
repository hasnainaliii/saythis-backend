package handler

import (
	"net/http"

	"saythis-backend/internal/auth/usecase"
	"saythis-backend/internal/util"
)

type RefreshHandler struct {
	usecase *usecase.AuthUseCase
}

func NewRefreshHandler(uc *usecase.AuthUseCase) *RefreshHandler {
	return &RefreshHandler{usecase: uc}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req refreshRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.usecase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		status, msg := mapAuthError(err)
		util.Error(w, status, msg)
		return
	}

	util.JSON(w, http.StatusOK, refreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
