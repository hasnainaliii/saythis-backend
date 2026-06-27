package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

func (uc *TherapyUseCase) GetProgress(ctx context.Context, userID uuid.UUID) ([]*therapydomain.ExerciseProgress, error) {
	progress, err := uc.therapyRepo.GetProgressByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}
	return progress, nil
}
