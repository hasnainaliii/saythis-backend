package domain

import "errors"

var (
	ErrDateRequired           = errors.New("date is required")
	ErrInvalidDate            = errors.New("invalid date")
	ErrFutureDate             = errors.New("date cannot be in the future")
	ErrInvalidDateRange       = errors.New("invalid date range")
	ErrInvalidMood            = errors.New("invalid mood value")
	ErrInvalidSleepHours      = errors.New("sleep_hours must be 0-12")
	ErrInvalidJournalEntry    = errors.New("journal_entry is too long")
	ErrInvalidStressLevel     = errors.New("stress_level must be 1-5")
	ErrInvalidMindfulHours    = errors.New("mindful_hours must be 0-8")
	ErrInvalidStutterScore    = errors.New("stutter_score must be 0-100")
	ErrInvalidStutterCount    = errors.New("stutter_count must be non-negative")
	ErrInvalidRepetitionCount = errors.New("repetition_count must be non-negative")
	ErrInvalidFillerCount     = errors.New("filler_count must be non-negative")
	ErrInvalidTotalWords      = errors.New("total_words must be non-negative")
	ErrInvalidTranscript      = errors.New("stutter_transcript is too long")
	ErrIncompleteStutterData  = errors.New("incomplete stutter data")
	ErrDailyStatNotFound      = errors.New("daily stat not found")
)
