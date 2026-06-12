// Package usecase implements stats orchestration and aggregation.
package usecase

import statsrepo "saythis-backend/internal/src/stats/repository"

type StatsUseCase struct {
	statsRepo statsrepo.StatsRepository
}

func NewStatsUseCase(statsRepo statsrepo.StatsRepository) *StatsUseCase {
	return &StatsUseCase{statsRepo: statsRepo}
}
