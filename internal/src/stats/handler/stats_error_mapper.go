package handler

import (
	"errors"
	"net/http"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

var errInvalidRequestBody = errors.New("invalid request body")

func mapStatsError(err error) (int, string) {
	switch {
	case errors.Is(err, errInvalidRequestBody):
		return http.StatusBadRequest, "invalid request body"
	case errors.Is(err, statsdomain.ErrDateRequired):
		return http.StatusBadRequest, "date is required"
	case errors.Is(err, statsdomain.ErrInvalidDate):
		return http.StatusBadRequest, "invalid date"
	case errors.Is(err, statsdomain.ErrFutureDate):
		return http.StatusBadRequest, "date cannot be in the future"
	case errors.Is(err, statsdomain.ErrInvalidDateRange):
		return http.StatusBadRequest, "invalid date range"
	case errors.Is(err, statsdomain.ErrInvalidMood):
		return http.StatusBadRequest, "Invalid mood value"
	case errors.Is(err, statsdomain.ErrInvalidSleepHours):
		return http.StatusBadRequest, "sleep_hours must be 0-12"
	case errors.Is(err, statsdomain.ErrInvalidJournalEntry):
		return http.StatusBadRequest, "journal_entry is too long"
	case errors.Is(err, statsdomain.ErrInvalidStressLevel):
		return http.StatusBadRequest, "stress_level must be 1-5"
	case errors.Is(err, statsdomain.ErrInvalidMindfulHours):
		return http.StatusBadRequest, "mindful_hours must be 0-8"
	case errors.Is(err, statsdomain.ErrInvalidStutterScore):
		return http.StatusBadRequest, "stutter_score must be 0-100"
	case errors.Is(err, statsdomain.ErrInvalidStutterCount):
		return http.StatusBadRequest, "stutter_count must be non-negative"
	case errors.Is(err, statsdomain.ErrInvalidRepetitionCount):
		return http.StatusBadRequest, "repetition_count must be non-negative"
	case errors.Is(err, statsdomain.ErrInvalidFillerCount):
		return http.StatusBadRequest, "filler_count must be non-negative"
	case errors.Is(err, statsdomain.ErrInvalidTotalWords):
		return http.StatusBadRequest, "total_words must be non-negative"
	case errors.Is(err, statsdomain.ErrInvalidTranscript):
		return http.StatusBadRequest, "stutter_transcript is too long"
	case errors.Is(err, statsdomain.ErrIncompleteStutterData):
		return http.StatusBadRequest, "Incomplete stutter data"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
