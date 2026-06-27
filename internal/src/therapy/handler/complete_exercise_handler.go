package handler

import (
	"net/http"
	"time"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/therapy/usecase"
)

type CompleteExerciseHandler struct {
	usecase *usecase.TherapyUseCase
}

func NewCompleteExerciseHandler(uc *usecase.TherapyUseCase) *CompleteExerciseHandler {
	return &CompleteExerciseHandler{usecase: uc}
}

type completeExerciseRequest struct {
	ChapterID  string `json:"chapter_id"`
	ExerciseID string `json:"exercise_id"`
	Rating     int    `json:"rating"`
	Remarks    string `json:"remarks"`
}

type completeExerciseResponse struct {
	ChapterID   string    `json:"chapter_id"`
	ExerciseID  string    `json:"exercise_id"`
	Rating      int       `json:"rating"`
	Remarks     string    `json:"remarks"`
	CompletedAt time.Time `json:"completed_at"`
}

func (h *CompleteExerciseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req completeExerciseRequest
	if err := helper.DecodeJSON(r, &req); err != nil {
		helper.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	progress, err := h.usecase.CompleteExercise(
		r.Context(),
		claims.UserID,
		req.ChapterID,
		req.ExerciseID,
		req.Rating,
		req.Remarks,
	)
	if err != nil {
		status, msg := mapTherapyError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, completeExerciseResponse{
		ChapterID:   progress.ChapterID(),
		ExerciseID:  progress.ExerciseID(),
		Rating:      progress.Rating(),
		Remarks:     progress.Remarks(),
		CompletedAt: progress.CompletedAt(),
	})
}
