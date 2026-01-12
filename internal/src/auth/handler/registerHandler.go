package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"saythis-backend/internal/src/auth/usecase"
	userDomain "saythis-backend/internal/src/user/domain"
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterHandler struct {
	orchestrator *usecase.RegisterOrchestrator
	logger       *log.Logger
}

func NewRegisterHandler(orchestrator *usecase.RegisterOrchestrator, logger *log.Logger) *RegisterHandler {
	logger.Printf("[DEBUG] Created RegisterHandler with orchestrator: %p", orchestrator)
	return &RegisterHandler{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("[API] Incoming registration request: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		h.logger.Printf("[WARN] Method not allowed: %s", r.Method)
		h.respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("[ERROR] Failed to decode request body: %v", err)
		h.respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	h.logger.Printf("[DEBUG] Request decoded: Email=%s, FullName=%s, Password=[REDACTED]", req.Email, req.FullName)

	if req.Email == "" || req.FullName == "" || req.Password == "" {
		h.logger.Printf("[WARN] Missing required fields in request")
		h.respondWithError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	h.logger.Println("[DEBUG] Handing over to RegisterOrchestrator...")
	user, err := h.orchestrator.Register(ctx, req.Email, req.FullName, req.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := RegisterResponse{
		ID:        user.ID(),
		Email:     user.Email(),
		FullName:  user.FullName(),
		CreatedAt: user.CreatedAt(),
	}

	h.logger.Printf("[API] Successfully registered user: %s. Returning 201 Created.", user.ID())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RegisterHandler) handleError(w http.ResponseWriter, err error) {
	h.logger.Printf("[ERROR] Registration failed: %v", err)

	switch {
	case errors.Is(err, userDomain.ErrEmptyEmail),
		errors.Is(err, userDomain.ErrInvalidEmail),
		errors.Is(err, userDomain.ErrEmptyFullName),
		errors.Is(err, userDomain.ErrInvalidFullNameLength):
		h.logger.Printf("[INFO] Error mapped to 400 Bad Request")
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

	case err.Error() == "password must be at least 8 characters",
		err.Error() == "password must contain at least one number",
		err.Error() == "password must contain at least one special character":
		h.logger.Printf("[INFO] Error mapped to 400 Bad Request (Password policy)")
		h.respondWithError(w, err.Error(), http.StatusBadRequest)

	default:
		h.logger.Printf("[CRITICAL] Unhandled error: %v. Returning 500.", err)
		h.respondWithError(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *RegisterHandler) respondWithError(w http.ResponseWriter, message string, code int) {
	h.logger.Printf("[API] Response: %d | Message: %s", code, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
