package repository

import (
	"context"

	"github.com/google/uuid"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

type TherapyRepository interface {
	UpsertExerciseProgress(ctx context.Context, progress *therapydomain.ExerciseProgress) error

	GetProgressByUserID(ctx context.Context, userID uuid.UUID) ([]*therapydomain.ExerciseProgress, error)
}
