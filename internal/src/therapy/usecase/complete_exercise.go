package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

func (uc *TherapyUseCase) CompleteExercise(
	ctx context.Context,
	userID uuid.UUID,
	chapterID, exerciseID string,
	rating int,
	remarks string,
) (*therapydomain.ExerciseProgress, error) {

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

	progress := therapydomain.NewExerciseProgress(
		userID, chapterID, exerciseID, rating, remarks, time.Now().UTC(),
	)

	if err := uc.therapyRepo.UpsertExerciseProgress(ctx, progress); err != nil {
		return nil, fmt.Errorf("complete exercise: %w", err)
	}

	return progress, nil
}
