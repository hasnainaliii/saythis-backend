package handler

import (
	"errors"
	"net/http"
	"time"

	"saythis-backend/internal/response"
	"saythis-backend/internal/user/domain"
	"saythis-backend/internal/user/usecase"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type CreateUserResponse struct {
	ID        uuid.UUID         `json:"id"`
	Email     string            `json:"email"`
	FullName  string            `json:"full_name"`
	Role      domain.UserRole   `json:"role"`
	Status    domain.UserStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

type CreateUserHandler struct {
	usecase *usecase.UserUseCase
}

func NewCreateUserHandler(uc *usecase.UserUseCase) *CreateUserHandler {
	return &CreateUserHandler{usecase: uc}
}

func (h *CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.usecase.CreateUser(r.Context(), req.Email, req.FullName)
	if err != nil {
		status, msg := errorToHTTP(err)
		response.Error(w, status, msg)
		return
	}

	response.JSON(w, http.StatusCreated, CreateUserResponse{
		ID:        user.ID(),
		Email:     user.Email(),
		FullName:  user.FullName(),
		Role:      user.Role(),
		Status:    user.Status(),
		CreatedAt: user.CreatedAt(),
	})
}

func errorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrEmptyEmail),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrEmptyFullName),
		errors.Is(err, domain.ErrInvalidFullNameLength),
		errors.Is(err, domain.ErrInvalidRole):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, domain.ErrDuplicateEmail):
		return http.StatusConflict, domain.ErrDuplicateEmail.Error()

	default:

		return http.StatusInternalServerError, "internal server error"
	}
}
