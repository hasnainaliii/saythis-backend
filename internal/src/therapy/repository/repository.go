package repository

import (
	"context"

	"github.com/google/uuid"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

// TherapyRepository defines the persistence contract for therapy progress data.
type TherapyRepository interface {
	// UpsertExerciseProgress inserts a new exercise progress record or updates the
	// existing one for the same (user_id, exercise_id) pair with fresh rating,
	// remarks, and completed_at. This allows users to re-submit an exercise.
	UpsertExerciseProgress(ctx context.Context, progress *therapydomain.ExerciseProgress) error

	// GetProgressByUserID returns all completed exercise records for the given user,
	// ordered chronologically (oldest first) so the client can derive unlock state.
	GetProgressByUserID(ctx context.Context, userID uuid.UUID) ([]*therapydomain.ExerciseProgress, error)
}
