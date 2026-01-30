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

type DeleteMeHandler struct {
	userUC *usecase.UserUseCase
}

func NewDeleteMeHandler(userUC *usecase.UserUseCase) *DeleteMeHandler {
	return &DeleteMeHandler{
		userUC: userUC,
	}
}

func (h *DeleteMeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		h.respondWithError(w, "UNAUTHORIZED", "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Fix: UserID is stored as uuid.UUID in context, not string
	uid, ok := userID.(uuid.UUID)
	if !ok {
		zap.S().Error("User ID from context is not of type uuid.UUID")
		h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr := uid.String()

	ctx := r.Context()
	if err := h.userUC.DeleteUser(ctx, userIDStr); err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

func (h *DeleteMeHandler) handleError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if e, ok := err.(*apperror.AppError); ok {
		appErr = e
	} else {
		appErr = apperror.New("INTERNAL_ERROR", "Internal server error", 500)
	}

	h.respondWithError(w, appErr.Code, appErr.Message, appErr.HTTPStatus)
}

func (h *DeleteMeHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
