package handler

import (
	"net/http"

	"saythis-backend/internal/auth/usecase"
	"saythis-backend/internal/util"
)

// LoginHandler handles POST /api/v1/auth/login.
// It delegates credential verification to the use case, which returns a fresh
// token pair on success. The response shape is identical to the registration
// response so clients can handle both flows with the same code path.
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

	// ── Decode ────────────────────────────────────────────────────────────────
	var req loginRequest
	if err := util.DecodeJSON(r, &req); err != nil {
		util.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// ── Authenticate ──────────────────────────────────────────────────────────
	user, tokens, err := h.usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		status, msg := mapAuthError(err)
		util.Error(w, status, msg)
		return
	}

	// ── Respond ───────────────────────────────────────────────────────────────
	util.JSON(w, http.StatusOK, loginResponse{
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
