// Package usecase implements the therapy progress use-case layer.
// Each file covers one distinct operation:
//
//   - complete_exercise.go — records (or updates) a completed exercise for a user
//   - get_progress.go      — fetches all completed exercises for a user
//
// All use cases share the TherapyUseCase struct defined here.
package usecase

import (
	therapyrepo "saythis-backend/internal/src/therapy/repository"
)

// TherapyUseCase orchestrates all therapy progress operations.
type TherapyUseCase struct {
	therapyRepo therapyrepo.TherapyRepository
}

// NewTherapyUseCase constructs a TherapyUseCase. Call this once at startup and
// share the result across handlers — it is safe for concurrent use.
func NewTherapyUseCase(therapyRepo therapyrepo.TherapyRepository) *TherapyUseCase {
	return &TherapyUseCase{therapyRepo: therapyRepo}
}
