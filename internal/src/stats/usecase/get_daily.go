package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

func (uc *StatsUseCase) GetDailyStat(ctx context.Context, userID uuid.UUID, date time.Time) (*statsdomain.DailyStat, error) {
	if date.IsZero() {
		return nil, statsdomain.ErrInvalidDate
	}
	if isFutureDate(date) {
		return nil, statsdomain.ErrFutureDate
	}

	stat, err := uc.statsRepo.GetDailyStatByDate(ctx, userID, startOfDayUTC(date))
	if err != nil {
		if errors.Is(err, statsdomain.ErrDailyStatNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get daily stat: %w", err)
	}
	return stat, nil
}
