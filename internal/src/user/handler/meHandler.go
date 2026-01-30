package handler

import (
	"encoding/json"
	"net/http"
	"saythis-backend/internal/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	email := r.Context().Value(middleware.EmailKey)
	role := r.Context().Value(middleware.RoleKey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}
