package domain

import (
	"time"

	"github.com/google/uuid"
)

type Optional[T any] struct {
	Present bool
	Value   *T
}

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
