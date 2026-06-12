package handler

import (
	"time"

	"github.com/google/uuid"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

type dailyStatResponse struct {
	ID                uuid.UUID `json:"id"`
	Date              string    `json:"date"`
	Mood              *string   `json:"mood"`
	SleepHours        *float64  `json:"sleep_hours"`
	JournalEntry      *string   `json:"journal_entry"`
	StressLevel       *int      `json:"stress_level"`
	MindfulHours      *float64  `json:"mindful_hours"`
	StutterScore      *float64  `json:"stutter_score"`
	StutterCount      *int      `json:"stutter_count"`
	RepetitionCount   *int      `json:"repetition_count"`
	FillerCount       *int      `json:"filler_count"`
	TotalWords        *int      `json:"total_words"`
	StutterTranscript *string   `json:"stutter_transcript"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type dailySnapshotResponse struct {
	Date              string   `json:"date"`
	Mood              *string  `json:"mood"`
	SleepHours        *float64 `json:"sleep_hours"`
	JournalEntry      *string  `json:"journal_entry"`
	StressLevel       *int     `json:"stress_level"`
	MindfulHours      *float64 `json:"mindful_hours"`
	StutterScore      *float64 `json:"stutter_score"`
	StutterCount      *int     `json:"stutter_count"`
	RepetitionCount   *int     `json:"repetition_count"`
	FillerCount       *int     `json:"filler_count"`
	TotalWords        *int     `json:"total_words"`
	StutterTranscript *string  `json:"stutter_transcript"`
}

type statsResponse struct {
	DailyStats      []dailySnapshotResponse  `json:"daily_stats"`
	Today           *dailySnapshotResponse   `json:"today"`
	JournalStreak   int                      `json:"journal_streak"`
	WellnessSummary wellnessSummaryResponse  `json:"wellness_summary"`
	StutterSummary  stutterSummaryResponse   `json:"stutter_summary"`
	ToolStats       toolStatsResponse        `json:"tool_stats"`
	WeeklyActivity  []weeklyActivityResponse `json:"weekly_activity"`
	WeeklyTrend     []weeklyTrendResponse    `json:"weekly_trend"`
	RecentSessions  []recentSessionResponse  `json:"recent_sessions"`
}

type wellnessSummaryResponse struct {
	AvgSleepHours       *float64       `json:"avg_sleep_hours"`
	AvgStressLevel      *float64       `json:"avg_stress_level"`
	AvgMindfulHours     *float64       `json:"avg_mindful_hours"`
	TotalJournalEntries int            `json:"total_journal_entries"`
	MoodDistribution    map[string]int `json:"mood_distribution"`
	DaysTracked         int            `json:"days_tracked"`
}

type stutterSummaryResponse struct {
	AvgScore      *float64             `json:"avg_score"`
	BestScore     *float64             `json:"best_score"`
	WorstScore    *float64             `json:"worst_score"`
	TotalAnalyses int                  `json:"total_analyses"`
	ScoreTrend    []scoreTrendResponse `json:"score_trend"`
	LatestScore   *float64             `json:"latest_score"`
}

type scoreTrendResponse struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
}

type toolStatsResponse struct {
	Combined    combinedToolStatsResponse    `json:"combined"`
	DAF         dafToolStatsResponse         `json:"daf"`
	FAF         fafToolStatsResponse         `json:"faf"`
	Breathing   breathingToolStatsResponse   `json:"breathing"`
	Drills      drillToolStatsResponse       `json:"drills"`
	Biofeedback biofeedbackToolStatsResponse `json:"biofeedback"`
	Simulation  simulationToolStatsResponse  `json:"simulation"`
}

type combinedToolStatsResponse struct {
	TotalSessions int        `json:"total_sessions"`
	TotalMinutes  float64    `json:"total_minutes"`
	CurrentStreak int        `json:"current_streak"`
	BestStreak    int        `json:"best_streak"`
	ActiveDays    int        `json:"active_days"`
	LastSessionAt *time.Time `json:"last_session_at"`
}

type dafToolStatsResponse struct {
	TotalSessions    int      `json:"total_sessions"`
	TotalMinutes     float64  `json:"total_minutes"`
	AvgRating        *float64 `json:"avg_rating"`
	AvgDelayMS       *float64 `json:"avg_delay_ms"`
	SessionsThisWeek int      `json:"sessions_this_week"`
	BestStreak       int      `json:"best_streak"`
}

type fafToolStatsResponse struct {
	TotalSessions      int      `json:"total_sessions"`
	TotalMinutes       float64  `json:"total_minutes"`
	AvgRating          *float64 `json:"avg_rating"`
	PreferredDirection *string  `json:"preferred_direction"`
	AvgSemitones       *float64 `json:"avg_semitones"`
	SessionsThisWeek   int      `json:"sessions_this_week"`
	BestStreak         int      `json:"best_streak"`
}

type breathingToolStatsResponse struct {
	TotalSessions         int            `json:"total_sessions"`
	TotalMinutes          float64        `json:"total_minutes"`
	AvgRating             *float64       `json:"avg_rating"`
	BoxBreathingSessions  int            `json:"box_breathing_sessions"`
	DiaphragmaticSessions int            `json:"diaphragmatic_sessions"`
	PreSpeechSessions     int            `json:"pre_speech_sessions"`
	SituationBreakdown    map[string]int `json:"situation_breakdown"`
	CurrentStreak         int            `json:"current_streak"`
}

type drillToolStatsResponse struct {
	TotalSessions           int      `json:"total_sessions"`
	TotalMinutes            float64  `json:"total_minutes"`
	AvgRating               *float64 `json:"avg_rating"`
	GentleOnsetSessions     int      `json:"gentle_onset_sessions"`
	ProlongedSpeechSessions int      `json:"prolonged_speech_sessions"`
	AvgGentleScore          *float64 `json:"avg_gentle_score"`
	AvgProlongedWPM         *float64 `json:"avg_prolonged_wpm"`
	CurrentStreak           int      `json:"current_streak"`
}

type biofeedbackToolStatsResponse struct {
	TotalSessions        int      `json:"total_sessions"`
	TotalMinutes         float64  `json:"total_minutes"`
	AvgRating            *float64 `json:"avg_rating"`
	StutterTapSessions   int      `json:"stutter_tap_sessions"`
	TimedReadingSessions int      `json:"timed_reading_sessions"`
	AvgStuttersPerMin    *float64 `json:"avg_stutters_per_min"`
	AvgReadingWPM        *float64 `json:"avg_reading_wpm"`
	CurrentStreak        int      `json:"current_streak"`
}

type simulationToolStatsResponse struct {
	TotalSessions      int      `json:"total_sessions"`
	TotalMinutes       float64  `json:"total_minutes"`
	AvgRating          *float64 `json:"avg_rating"`
	CoffeeSessions     int      `json:"coffee_sessions"`
	CallSessions       int      `json:"call_sessions"`
	AvgCompletionScore *float64 `json:"avg_completion_score"`
	CurrentStreak      int      `json:"current_streak"`
}

type weeklyActivityResponse struct {
	Date     string `json:"date"`
	Day      string `json:"day"`
	Sessions int    `json:"sessions"`
}

type weeklyTrendResponse struct {
	WeekLabel    string  `json:"week_label"`
	TotalMinutes float64 `json:"total_minutes"`
}

type recentSessionResponse struct {
	ID              uuid.UUID `json:"id"`
	ToolType        string    `json:"tool_type"`
	StartedAt       time.Time `json:"started_at"`
	DurationSeconds int       `json:"duration_seconds"`
	SelfRating      *int      `json:"self_rating"`
}

func toStatsResponse(stats *statsdomain.StatsOverview) statsResponse {
	dailyStats := make([]dailySnapshotResponse, 0, len(stats.DailyStats))
	for _, stat := range stats.DailyStats {
		dailyStats = append(dailyStats, toDailySnapshotResponse(stat))
	}

	var today *dailySnapshotResponse
	if stats.Today != nil {
		mapped := toDailySnapshotResponse(stats.Today)
		today = &mapped
	}

	return statsResponse{
		DailyStats:      dailyStats,
		Today:           today,
		JournalStreak:   stats.JournalStreak,
		WellnessSummary: toWellnessSummaryResponse(stats.WellnessSummary),
		StutterSummary:  toStutterSummaryResponse(stats.StutterSummary),
		ToolStats:       toToolStatsResponse(stats.ToolStats),
		WeeklyActivity:  toWeeklyActivityResponse(stats.WeeklyActivity),
		WeeklyTrend:     toWeeklyTrendResponse(stats.WeeklyTrend),
		RecentSessions:  toRecentSessionResponse(stats.RecentSessions),
	}
}

func toDailyStatResponse(stat *statsdomain.DailyStat) dailyStatResponse {
	return dailyStatResponse{
		ID:                stat.ID,
		Date:              formatDate(stat.Date),
		Mood:              stat.Mood,
		SleepHours:        stat.SleepHours,
		JournalEntry:      stat.JournalEntry,
		StressLevel:       stat.StressLevel,
		MindfulHours:      stat.MindfulHours,
		StutterScore:      stat.StutterScore,
		StutterCount:      stat.StutterCount,
		RepetitionCount:   stat.RepetitionCount,
		FillerCount:       stat.FillerCount,
		TotalWords:        stat.TotalWords,
		StutterTranscript: stat.StutterTranscript,
		CreatedAt:         stat.CreatedAt,
		UpdatedAt:         stat.UpdatedAt,
	}
}

func toDailySnapshotResponse(stat *statsdomain.DailyStat) dailySnapshotResponse {
	return dailySnapshotResponse{
		Date:              formatDate(stat.Date),
		Mood:              stat.Mood,
		SleepHours:        stat.SleepHours,
		JournalEntry:      stat.JournalEntry,
		StressLevel:       stat.StressLevel,
		MindfulHours:      stat.MindfulHours,
		StutterScore:      stat.StutterScore,
		StutterCount:      stat.StutterCount,
		RepetitionCount:   stat.RepetitionCount,
		FillerCount:       stat.FillerCount,
		TotalWords:        stat.TotalWords,
		StutterTranscript: stat.StutterTranscript,
	}
}

func toWellnessSummaryResponse(summary statsdomain.WellnessSummary) wellnessSummaryResponse {
	return wellnessSummaryResponse{
		AvgSleepHours:       summary.AvgSleepHours,
		AvgStressLevel:      summary.AvgStressLevel,
		AvgMindfulHours:     summary.AvgMindfulHours,
		TotalJournalEntries: summary.TotalJournalEntries,
		MoodDistribution:    summary.MoodDistribution,
		DaysTracked:         summary.DaysTracked,
	}
}

func toStutterSummaryResponse(summary statsdomain.StutterSummary) stutterSummaryResponse {
	trend := make([]scoreTrendResponse, 0, len(summary.ScoreTrend))
	for _, point := range summary.ScoreTrend {
		trend = append(trend, scoreTrendResponse{Date: formatDate(point.Date), Score: point.Score})
	}

	return stutterSummaryResponse{
		AvgScore:      summary.AvgScore,
		BestScore:     summary.BestScore,
		WorstScore:    summary.WorstScore,
		TotalAnalyses: summary.TotalAnalyses,
		ScoreTrend:    trend,
		LatestScore:   summary.LatestScore,
	}
}

func toToolStatsResponse(stats statsdomain.ToolStats) toolStatsResponse {
	return toolStatsResponse{
		Combined:    combinedToolStatsResponse(stats.Combined),
		DAF:         dafToolStatsResponse(stats.DAF),
		FAF:         fafToolStatsResponse(stats.FAF),
		Breathing:   breathingToolStatsResponse(stats.Breathing),
		Drills:      drillToolStatsResponse(stats.Drills),
		Biofeedback: biofeedbackToolStatsResponse(stats.Biofeedback),
		Simulation:  simulationToolStatsResponse(stats.Simulation),
	}
}

func toWeeklyActivityResponse(activity []statsdomain.WeeklyActivityDay) []weeklyActivityResponse {
	response := make([]weeklyActivityResponse, 0, len(activity))
	for _, day := range activity {
		response = append(response, weeklyActivityResponse{
			Date:     formatDate(day.Date),
			Day:      day.Day,
			Sessions: day.Sessions,
		})
	}
	return response
}

func toWeeklyTrendResponse(trend []statsdomain.WeeklyTrendWeek) []weeklyTrendResponse {
	response := make([]weeklyTrendResponse, 0, len(trend))
	for _, week := range trend {
		response = append(response, weeklyTrendResponse{
			WeekLabel:    week.WeekLabel,
			TotalMinutes: week.TotalMinutes,
		})
	}
	return response
}

func toRecentSessionResponse(sessions []*statsdomain.ToolSession) []recentSessionResponse {
	response := make([]recentSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, recentSessionResponse{
			ID:              session.ID,
			ToolType:        session.ToolType,
			StartedAt:       session.StartedAt,
			DurationSeconds: session.DurationSeconds,
			SelfRating:      session.SelfRating,
		})
	}
	return response
}

func formatDate(date time.Time) string {
	return date.UTC().Format(dateLayout)
}
