package handler

import (
	"encoding/json"
	"net/http"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/middleware"
	"saythis-backend/internal/src/user/usecase"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UpdateProfileHandler struct {
	userUC *usecase.UserUseCase
}

func NewUpdateProfileHandler(userUC *usecase.UserUseCase) *UpdateProfileHandler {
	return &UpdateProfileHandler{
		userUC: userUC,
	}
}

type UpdateProfileRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		h.respondWithError(w, "UNAUTHORIZED", "User not authenticated", http.StatusUnauthorized)
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		zap.S().Error("User ID from context is not of type uuid.UUID")
		h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr := uid.String()

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, "INVALID_JSON", "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if at least one field is provided
	if req.FullName == nil && req.AvatarURL == nil {
		h.respondWithError(w, "VALIDATION_ERROR", "At least one field must be provided", http.StatusBadRequest)
		return
	}

	input := usecase.UpdateProfileInput{
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}

	ctx := r.Context()
	if err := h.userUC.UpdateProfile(ctx, userIDStr, input); err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(UpdateProfileResponse{
		Message: "Profile updated successfully",
	})
}

func (h *UpdateProfileHandler) handleError(w http.ResponseWriter, err error) {
	// Check for ProfileValidationError
	if validErr, ok := err.(*usecase.ProfileValidationError); ok {
		h.respondWithError(w, "VALIDATION_ERROR", validErr.Error(), http.StatusBadRequest)
		return
	}

	// Check for AppError
	var appErr *apperror.AppError
	if e, ok := err.(*apperror.AppError); ok {
		appErr = e
	} else {
		appErr = apperror.New("INTERNAL_ERROR", "Internal server error", 500)
	}

	h.respondWithError(w, appErr.Code, appErr.Message, appErr.HTTPStatus)
}

func (h *UpdateProfileHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
