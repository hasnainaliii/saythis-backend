package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

const (
	maxJournalEntryLength = 10000
	maxTranscriptLength   = 50000
)

var allowedMoods = map[string]struct{}{
	"Dizzy":   {},
	"Sad":     {},
	"Neutral": {},
	"Happy":   {},
	"Great":   {},
}

func (uc *StatsUseCase) UpdateDailyStat(ctx context.Context, userID uuid.UUID, patch statsdomain.DailyStatPatch) (*statsdomain.DailyStat, error) {
	if err := validateDailyStatPatch(patch); err != nil {
		return nil, err
	}

	stat, err := uc.statsRepo.UpsertDailyStat(ctx, userID, patch)
	if err != nil {
		return nil, fmt.Errorf("update daily stat: %w", err)
	}
	return stat, nil
}

func validateDailyStatPatch(patch statsdomain.DailyStatPatch) error {
	if patch.Date.IsZero() {
		return statsdomain.ErrDateRequired
	}
	if isFutureDate(patch.Date) {
		return statsdomain.ErrFutureDate
	}

	if patch.Mood.Present && patch.Mood.Value != nil {
		if _, ok := allowedMoods[*patch.Mood.Value]; !ok {
			return statsdomain.ErrInvalidMood
		}
	}
	if patch.SleepHours.Present && patch.SleepHours.Value != nil {
		if *patch.SleepHours.Value < 0 || *patch.SleepHours.Value > 12 || !isHalfStep(*patch.SleepHours.Value) {
			return statsdomain.ErrInvalidSleepHours
		}
	}
	if patch.JournalEntry.Present && patch.JournalEntry.Value != nil {
		if len(*patch.JournalEntry.Value) > maxJournalEntryLength {
			return statsdomain.ErrInvalidJournalEntry
		}
	}
	if patch.StressLevel.Present && patch.StressLevel.Value != nil {
		if *patch.StressLevel.Value < 1 || *patch.StressLevel.Value > 5 {
			return statsdomain.ErrInvalidStressLevel
		}
	}
	if patch.MindfulHours.Present && patch.MindfulHours.Value != nil {
		if *patch.MindfulHours.Value < 0 || *patch.MindfulHours.Value > 8 || !isHalfStep(*patch.MindfulHours.Value) {
			return statsdomain.ErrInvalidMindfulHours
		}
	}
	if patch.StutterScore.Present && patch.StutterScore.Value != nil {
		if *patch.StutterScore.Value < 0 || *patch.StutterScore.Value > 100 {
			return statsdomain.ErrInvalidStutterScore
		}
	}
	if patch.StutterCount.Present && patch.StutterCount.Value != nil && *patch.StutterCount.Value < 0 {
		return statsdomain.ErrInvalidStutterCount
	}
	if patch.RepetitionCount.Present && patch.RepetitionCount.Value != nil && *patch.RepetitionCount.Value < 0 {
		return statsdomain.ErrInvalidRepetitionCount
	}
	if patch.FillerCount.Present && patch.FillerCount.Value != nil && *patch.FillerCount.Value < 0 {
		return statsdomain.ErrInvalidFillerCount
	}
	if patch.TotalWords.Present && patch.TotalWords.Value != nil && *patch.TotalWords.Value < 0 {
		return statsdomain.ErrInvalidTotalWords
	}
	if patch.StutterTranscript.Present && patch.StutterTranscript.Value != nil {
		if len(*patch.StutterTranscript.Value) > maxTranscriptLength {
			return statsdomain.ErrInvalidTranscript
		}
	}

	if !allOrNonePresent(
		patch.StutterScore.Present,
		patch.StutterCount.Present,
		patch.RepetitionCount.Present,
		patch.FillerCount.Present,
		patch.TotalWords.Present,
	) {
		return statsdomain.ErrIncompleteStutterData
	}

	return nil
}

func allOrNonePresent(values ...bool) bool {
	presentCount := 0
	for _, present := range values {
		if present {
			presentCount++
		}
	}
	return presentCount == 0 || presentCount == len(values)
}

func isHalfStep(value float64) bool {
	return math.Abs(value*2-math.Round(value*2)) < 0.0000001
}

func isFutureDate(date time.Time) bool {
	return startOfDayUTC(date).After(todayUTC())
}

func todayUTC() time.Time {
	return startOfDayUTC(time.Now().UTC())
}

func startOfDayUTC(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
