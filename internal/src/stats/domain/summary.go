package domain

import "time"

type StatsOverview struct {
	DailyStats      []*DailyStat
	Today           *DailyStat
	JournalStreak   int
	WellnessSummary WellnessSummary
	StutterSummary  StutterSummary
	ToolStats       ToolStats
	WeeklyActivity  []WeeklyActivityDay
	WeeklyTrend     []WeeklyTrendWeek
	RecentSessions  []*ToolSession
}

type WellnessSummary struct {
	AvgSleepHours       *float64
	AvgStressLevel      *float64
	AvgMindfulHours     *float64
	TotalJournalEntries int
	MoodDistribution    map[string]int
	DaysTracked         int
}

type StutterSummary struct {
	AvgScore      *float64
	BestScore     *float64
	WorstScore    *float64
	TotalAnalyses int
	ScoreTrend    []ScoreTrendPoint
	LatestScore   *float64
}

type ScoreTrendPoint struct {
	Date  time.Time
	Score float64
}

type ToolStats struct {
	Combined    CombinedToolStats
	DAF         DAFToolStats
	FAF         FAFToolStats
	Breathing   BreathingToolStats
	Drills      DrillToolStats
	Biofeedback BiofeedbackToolStats
	Simulation  SimulationToolStats
}

type CombinedToolStats struct {
	TotalSessions int
	TotalMinutes  float64
	CurrentStreak int
	BestStreak    int
	ActiveDays    int
	LastSessionAt *time.Time
}

type DAFToolStats struct {
	TotalSessions    int
	TotalMinutes     float64
	AvgRating        *float64
	AvgDelayMS       *float64
	SessionsThisWeek int
	BestStreak       int
}

type FAFToolStats struct {
	TotalSessions      int
	TotalMinutes       float64
	AvgRating          *float64
	PreferredDirection *string
	AvgSemitones       *float64
	SessionsThisWeek   int
	BestStreak         int
}

type BreathingToolStats struct {
	TotalSessions         int
	TotalMinutes          float64
	AvgRating             *float64
	BoxBreathingSessions  int
	DiaphragmaticSessions int
	PreSpeechSessions     int
	SituationBreakdown    map[string]int
	CurrentStreak         int
}

type DrillToolStats struct {
	TotalSessions           int
	TotalMinutes            float64
	AvgRating               *float64
	GentleOnsetSessions     int
	ProlongedSpeechSessions int
	AvgGentleScore          *float64
	AvgProlongedWPM         *float64
	CurrentStreak           int
}

type BiofeedbackToolStats struct {
	TotalSessions        int
	TotalMinutes         float64
	AvgRating            *float64
	StutterTapSessions   int
	TimedReadingSessions int
	AvgStuttersPerMin    *float64
	AvgReadingWPM        *float64
	CurrentStreak        int
}

type SimulationToolStats struct {
	TotalSessions      int
	TotalMinutes       float64
	AvgRating          *float64
	CoffeeSessions     int
	CallSessions       int
	AvgCompletionScore *float64
	CurrentStreak      int
}

type WeeklyActivityDay struct {
	Date     time.Time
	Day      string
	Sessions int
}

type WeeklyTrendWeek struct {
	WeekLabel    string
	TotalMinutes float64
}
