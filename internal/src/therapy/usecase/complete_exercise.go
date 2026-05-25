package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

// CompleteExercise records that a user has completed a specific exercise.
// If the user has already completed the same exercise the record is updated
// with the new rating and remarks (upsert semantics), so re-submissions are safe.
//
// Error catalogue:
//
//	ErrInvalidChapterID  — chapter_id was empty
//	ErrInvalidExerciseID — exercise_id was empty
//	ErrInvalidRating     — rating was not in the range 1–5
func (uc *TherapyUseCase) CompleteExercise(
	ctx context.Context,
	userID uuid.UUID,
	chapterID, exerciseID string,
	rating int,
	remarks string,
) (*therapydomain.ExerciseProgress, error) {

	// ── 1. Normalise + validate inputs ────────────────────────────────────────
	chapterID = strings.TrimSpace(chapterID)
	exerciseID = strings.TrimSpace(exerciseID)
	remarks = strings.TrimSpace(remarks)

	if chapterID == "" {
		return nil, therapydomain.ErrInvalidChapterID
	}
	if exerciseID == "" {
		return nil, therapydomain.ErrInvalidExerciseID
	}
	if rating < 1 || rating > 5 {
		return nil, therapydomain.ErrInvalidRating
	}

	// ── 2. Build domain object and persist ────────────────────────────────────
	progress := therapydomain.NewExerciseProgress(
		userID, chapterID, exerciseID, rating, remarks, time.Now().UTC(),
	)

	if err := uc.therapyRepo.UpsertExerciseProgress(ctx, progress); err != nil {
		return nil, fmt.Errorf("complete exercise: %w", err)
	}

	return progress, nil
}
