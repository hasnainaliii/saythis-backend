package usecase

import (
	therapyrepo "saythis-backend/internal/src/therapy/repository"
)

type TherapyUseCase struct {
	therapyRepo therapyrepo.TherapyRepository
}

func NewTherapyUseCase(therapyRepo therapyrepo.TherapyRepository) *TherapyUseCase {
	return &TherapyUseCase{therapyRepo: therapyRepo}
}
