package domain

import (
	"time"

	"github.com/google/uuid"
)

// Optional distinguishes between an omitted field and an explicit null/value.
// It is used by PATCH /stats/daily so partial updates never erase fields unless
// the client intentionally sends null for that field.
type Optional[T any] struct {
	Present bool
	Value   *T
}

// DailyStatPatch contains the validated partial update for one user/day.
type DailyStatPatch struct {
	Date              time.Time
	Mood              Optional[string]
	SleepHours        Optional[float64]
	JournalEntry      Optional[string]
	StressLevel       Optional[int]
	MindfulHours      Optional[float64]
	StutterScore      Optional[float64]
	StutterCount      Optional[int]
	RepetitionCount   Optional[int]
	FillerCount       Optional[int]
	TotalWords        Optional[int]
	StutterTranscript Optional[string]
}

// DailyStat is the unified per-user, per-day wellness and stutter-analysis row.
type DailyStat struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Date              time.Time
	Mood              *string
	SleepHours        *float64
	JournalEntry      *string
	StressLevel       *int
	MindfulHours      *float64
	StutterScore      *float64
	StutterCount      *int
	RepetitionCount   *int
	FillerCount       *int
	TotalWords        *int
	StutterTranscript *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *DailyStat) HasTrackedData() bool {
	return s.Mood != nil ||
		s.SleepHours != nil ||
		s.JournalEntry != nil ||
		s.StressLevel != nil ||
		s.MindfulHours != nil ||
		s.StutterScore != nil ||
		s.StutterCount != nil ||
		s.RepetitionCount != nil ||
		s.FillerCount != nil ||
		s.TotalWords != nil ||
		s.StutterTranscript != nil
}
