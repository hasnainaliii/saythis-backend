package handler

import (
	"net/http"
	"time"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/therapy/usecase"
)

// GetProgressHandler handles GET /api/v1/therapy/progress.
// Returns all completed exercises for the authenticated user along with a
// running total, which the client uses to derive exercise unlock state.
type GetProgressHandler struct {
	usecase *usecase.TherapyUseCase
}

func NewGetProgressHandler(uc *usecase.TherapyUseCase) *GetProgressHandler {
	return &GetProgressHandler{usecase: uc}
}

type exerciseProgressItem struct {
	ChapterID   string    `json:"chapter_id"`
	ExerciseID  string    `json:"exercise_id"`
	Rating      int       `json:"rating"`
	Remarks     string    `json:"remarks"`
	CompletedAt time.Time `json:"completed_at"`
}

type getProgressResponse struct {
	CompletedExercises []exerciseProgressItem `json:"completed_exercises"`
	TotalCompleted     int                    `json:"total_completed"`
}

func (h *GetProgressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	progressList, err := h.usecase.GetProgress(r.Context(), claims.UserID)
	if err != nil {
		status, msg := mapTherapyError(err)
		helper.Error(w, status, msg)
		return
	}

	items := make([]exerciseProgressItem, 0, len(progressList))
	for _, p := range progressList {
		items = append(items, exerciseProgressItem{
			ChapterID:   p.ChapterID(),
			ExerciseID:  p.ExerciseID(),
			Rating:      p.Rating(),
			Remarks:     p.Remarks(),
			CompletedAt: p.CompletedAt(),
		})
	}

	helper.JSON(w, http.StatusOK, getProgressResponse{
		CompletedExercises: items,
		TotalCompleted:     len(items),
	})
}
