package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"saythis-backend/internal/apperror"
	authDomain "saythis-backend/internal/src/auth/domain"
	"saythis-backend/internal/src/auth/usecase"
	userDomain "saythis-backend/internal/src/user/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse provides a structured error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RegisterHandler struct {
	orchestrator *usecase.RegisterOrchestrator
}

func NewRegisterHandler(orchestrator *usecase.RegisterOrchestrator) *RegisterHandler {
	return &RegisterHandler{
		orchestrator: orchestrator,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.FullName == "" || req.Password == "" {
		h.respondWithError(w, "MISSING_FIELDS", "Missing required fields", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	zap.S().Infow("User registration attempt", "email", req.Email)
	user, err := h.orchestrator.Register(ctx, req.Email, req.FullName, req.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}

	zap.S().Infow("User registered successfully", "email", req.Email, "user_id", user.ID())

	resp := RegisterResponse{
		ID:        user.ID(),
		Email:     user.Email(),
		FullName:  user.FullName(),
		CreatedAt: user.CreatedAt(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RegisterHandler) handleError(w http.ResponseWriter, err error) {

	// Check for AppError first (database constraint errors, etc.)
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		h.respondWithError(w, appErr.Code, appErr.Message, appErr.HTTPStatus)
		return
	}

	// Handle domain validation errors
	switch {
	case errors.Is(err, userDomain.ErrEmptyEmail):
		h.respondWithError(w, "EMPTY_EMAIL", err.Error(), http.StatusBadRequest)
	case errors.Is(err, userDomain.ErrInvalidEmail):
		h.respondWithError(w, "INVALID_EMAIL", err.Error(), http.StatusBadRequest)
	case errors.Is(err, userDomain.ErrEmptyFullName):
		h.respondWithError(w, "EMPTY_FULL_NAME", err.Error(), http.StatusBadRequest)
	case errors.Is(err, userDomain.ErrInvalidFullNameLength):
		h.respondWithError(w, "INVALID_FULL_NAME_LENGTH", err.Error(), http.StatusBadRequest)
	case errors.Is(err, authDomain.ErrPasswordTooShort):
		h.respondWithError(w, "PASSWORD_TOO_SHORT", err.Error(), http.StatusBadRequest)
	case errors.Is(err, authDomain.ErrPasswordMissingNumber):
		h.respondWithError(w, "PASSWORD_MISSING_NUMBER", err.Error(), http.StatusBadRequest)
	case errors.Is(err, authDomain.ErrPasswordMissingSpecialChar):
		h.respondWithError(w, "PASSWORD_MISSING_SPECIAL_CHAR", err.Error(), http.StatusBadRequest)
	default:
		zap.S().Errorw("Registration failed", "error", err)
		h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
	}
}

func (h *RegisterHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
