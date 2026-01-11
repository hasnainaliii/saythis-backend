package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"saythis-backend/internal/src/user/usecase"
)

type RegisterUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type RegisterUserHandler struct {
	useCase *usecase.UserUseCase
	logger  *log.Logger
}

func NewRegisterUserHandler(uc *usecase.UserUseCase, logger *log.Logger) *RegisterUserHandler {
	return &RegisterUserHandler{
		useCase: uc,
		logger:  logger,
	}
}

func (h *RegisterUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("In the handler")

	// var req RegisterUserRequest
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, "invalid request body", http.StatusBadRequest)
	// 	return
	// }
	const maxRequestBodySize = 1048576
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.useCase.RegisterUser(r.Context(), req.Email, req.FullName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	resp := map[string]interface{}{
		"id":         user.ID(),
		"email":      user.Email(),
		"full_name":  user.FullName(),
		"role":       user.Role(),
		"status":     user.Status(),
		"created_at": user.CreatedAt(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
