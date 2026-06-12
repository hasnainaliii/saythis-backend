package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

const (
	dateLayout = "2006-01-02"
)

func (uc *StatsUseCase) GetStats(ctx context.Context, userID uuid.UUID, from, to *time.Time) (*statsdomain.StatsOverview, error) {
	today := todayUTC()
	fromDate := today.AddDate(0, 0, -30)
	toDate := today

	if from != nil {
		fromDate = startOfDayUTC(*from)
	}
	if to != nil {
		toDate = startOfDayUTC(*to)
	}
	if fromDate.After(toDate) {
		return nil, statsdomain.ErrInvalidDateRange
	}
	if toDate.After(today) {
		return nil, statsdomain.ErrFutureDate
	}

	dailyStats, err := uc.statsRepo.GetDailyStatsByRange(ctx, userID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get daily stats range: %w", err)
	}

	todayStat, err := uc.statsRepo.GetDailyStatByDate(ctx, userID, today)
	if err != nil {
		if !errors.Is(err, statsdomain.ErrDailyStatNotFound) {
			return nil, fmt.Errorf("get today stat: %w", err)
		}
		todayStat = nil
	}

	journalDates, err := uc.statsRepo.GetJournalEntryDates(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get journal dates: %w", err)
	}

	toolSessions, err := uc.statsRepo.GetToolSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get tool sessions: %w", err)
	}

	return &statsdomain.StatsOverview{
		DailyStats:      dailyStats,
		Today:           todayStat,
		JournalStreak:   calculateJournalStreak(journalDates, today),
		WellnessSummary: calculateWellnessSummary(dailyStats),
		StutterSummary:  calculateStutterSummary(dailyStats),
		ToolStats:       calculateToolStats(toolSessions, today),
		WeeklyActivity:  calculateWeeklyActivity(toolSessions, today),
		WeeklyTrend:     calculateWeeklyTrend(toolSessions, today),
		RecentSessions:  recentSessions(toolSessions, 10),
	}, nil
}

func calculateJournalStreak(dates []time.Time, today time.Time) int {
	dateSet := make(map[string]struct{}, len(dates))
	for _, date := range dates {
		dateSet[dateKey(date)] = struct{}{}
	}

	streak := 0
	for day := today; ; day = day.AddDate(0, 0, -1) {
		if _, ok := dateSet[dateKey(day)]; !ok {
			return streak
		}
		streak++
	}
}

func calculateWellnessSummary(stats []*statsdomain.DailyStat) statsdomain.WellnessSummary {
	sleep := avgCollector{}
	stress := avgCollector{}
	mindful := avgCollector{}
	moods := map[string]int{"Great": 0, "Happy": 0, "Neutral": 0, "Sad": 0, "Dizzy": 0}
	journalEntries := 0
	daysTracked := 0

	for _, stat := range stats {
		if stat.SleepHours != nil {
			sleep.add(*stat.SleepHours)
		}
		if stat.StressLevel != nil {
			stress.add(float64(*stat.StressLevel))
		}
		if stat.MindfulHours != nil {
			mindful.add(*stat.MindfulHours)
		}
		if stat.JournalEntry != nil {
			journalEntries++
		}
		if stat.Mood != nil {
			moods[*stat.Mood]++
		}
		if stat.HasTrackedData() {
			daysTracked++
		}
	}

	return statsdomain.WellnessSummary{
		AvgSleepHours:       sleep.avgPtr(),
		AvgStressLevel:      stress.avgPtr(),
		AvgMindfulHours:     mindful.avgPtr(),
		TotalJournalEntries: journalEntries,
		MoodDistribution:    moods,
		DaysTracked:         daysTracked,
	}
}

func calculateStutterSummary(stats []*statsdomain.DailyStat) statsdomain.StutterSummary {
	avgScore := avgCollector{}
	trend := make([]statsdomain.ScoreTrendPoint, 0)
	var bestScore *float64
	var worstScore *float64
	var latestDate time.Time
	var latestScore *float64

	for _, stat := range stats {
		if stat.StutterScore == nil {
			continue
		}

		score := *stat.StutterScore
		avgScore.add(score)
		trend = append(trend, statsdomain.ScoreTrendPoint{Date: stat.Date, Score: round1(score)})
		if bestScore == nil || score < *bestScore {
			bestScore = floatPtr(round1(score))
		}
		if worstScore == nil || score > *worstScore {
			worstScore = floatPtr(round1(score))
		}
		if latestScore == nil || stat.Date.After(latestDate) {
			latestDate = stat.Date
			latestScore = floatPtr(round1(score))
		}
	}

	sort.Slice(trend, func(i, j int) bool {
		return trend[i].Date.Before(trend[j].Date)
	})

	return statsdomain.StutterSummary{
		AvgScore:      avgScore.avgPtr(),
		BestScore:     bestScore,
		WorstScore:    worstScore,
		TotalAnalyses: avgScore.count,
		ScoreTrend:    trend,
		LatestScore:   latestScore,
	}
}

func calculateToolStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.ToolStats {
	dafSessions := filterToolSessions(sessions, "DAF")
	fafSessions := filterToolSessions(sessions, "FAF")
	breathingSessions := filterToolSessions(sessions, "BOX_BREATHING", "DIAPHRAGMATIC", "PRE_SPEECH")
	drillSessions := filterToolSessions(sessions, "GENTLE_ONSET", "PROLONGED_SPEECH")
	biofeedbackSessions := filterToolSessions(sessions, "STUTTER_TAP_COUNTER", "TIMED_READING_WPM")
	simulationSessions := filterToolSessions(sessions, "VIRTUAL_COFFEE_ORDER", "PHONE_CALL_SIMULATOR")

	return statsdomain.ToolStats{
		Combined:    calculateCombinedToolStats(sessions, today),
		DAF:         calculateDAFStats(dafSessions, today),
		FAF:         calculateFAFStats(fafSessions, today),
		Breathing:   calculateBreathingStats(breathingSessions, today),
		Drills:      calculateDrillStats(drillSessions, today),
		Biofeedback: calculateBiofeedbackStats(biofeedbackSessions, today),
		Simulation:  calculateSimulationStats(simulationSessions, today),
	}
}

func calculateCombinedToolStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.CombinedToolStats {
	dates := sessionDates(sessions)
	var lastSessionAt *time.Time
	if len(sessions) > 0 {
		latest := sessions[0].StartedAt
		lastSessionAt = &latest
	}

	return statsdomain.CombinedToolStats{
		TotalSessions: len(sessions),
		TotalMinutes:  totalMinutes(sessions),
		CurrentStreak: currentStreak(dates, today),
		BestStreak:    bestStreak(dates),
		ActiveDays:    len(dates),
		LastSessionAt: lastSessionAt,
	}
}

func calculateDAFStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.DAFToolStats {
	delay := avgCollector{}
	for _, session := range sessions {
		if value, ok := metadataNumber(session.Metadata, "delay_ms"); ok {
			delay.add(value)
		}
	}

	return statsdomain.DAFToolStats{
		TotalSessions:    len(sessions),
		TotalMinutes:     totalMinutes(sessions),
		AvgRating:        averageRating(sessions),
		AvgDelayMS:       delay.avgPtr(),
		SessionsThisWeek: sessionsThisWeek(sessions),
		BestStreak:       bestStreak(sessionDates(sessions)),
	}
}

func calculateFAFStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.FAFToolStats {
	semitones := avgCollector{}
	directions := map[string]int{}
	for _, session := range sessions {
		if value, ok := metadataNumber(session.Metadata, "pitch_semitones"); ok {
			semitones.add(value)
		}
		if direction, ok := metadataString(session.Metadata, "pitch_direction"); ok {
			directions[direction]++
		}
	}

	return statsdomain.FAFToolStats{
		TotalSessions:      len(sessions),
		TotalMinutes:       totalMinutes(sessions),
		AvgRating:          averageRating(sessions),
		PreferredDirection: modeStringPtr(directions),
		AvgSemitones:       semitones.avgPtr(),
		SessionsThisWeek:   sessionsThisWeek(sessions),
		BestStreak:         bestStreak(sessionDates(sessions)),
	}
}

func calculateBreathingStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.BreathingToolStats {
	situationBreakdown := map[string]int{}
	stats := statsdomain.BreathingToolStats{SituationBreakdown: situationBreakdown}

	for _, session := range sessions {
		switch session.ToolType {
		case "BOX_BREATHING":
			stats.BoxBreathingSessions++
		case "DIAPHRAGMATIC":
			stats.DiaphragmaticSessions++
		case "PRE_SPEECH":
			stats.PreSpeechSessions++
			if situation, ok := metadataString(session.Metadata, "situation"); ok {
				situationBreakdown[situation]++
			}
		}
	}

	stats.TotalSessions = len(sessions)
	stats.TotalMinutes = totalMinutes(sessions)
	stats.AvgRating = averageRating(sessions)
	stats.CurrentStreak = currentStreak(sessionDates(sessions), today)
	return stats
}

func calculateDrillStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.DrillToolStats {
	gentleScore := avgCollector{}
	prolongedWPM := avgCollector{}
	stats := statsdomain.DrillToolStats{}

	for _, session := range sessions {
		switch session.ToolType {
		case "GENTLE_ONSET":
			stats.GentleOnsetSessions++
			if value, ok := metadataNumber(session.Metadata, "average_score"); ok {
				gentleScore.add(value)
			}
		case "PROLONGED_SPEECH":
			stats.ProlongedSpeechSessions++
			if value, ok := metadataNumber(session.Metadata, "estimated_wpm"); ok {
				prolongedWPM.add(value)
			}
		}
	}

	stats.TotalSessions = len(sessions)
	stats.TotalMinutes = totalMinutes(sessions)
	stats.AvgRating = averageRating(sessions)
	stats.AvgGentleScore = gentleScore.avgPtr()
	stats.AvgProlongedWPM = prolongedWPM.avgPtr()
	stats.CurrentStreak = currentStreak(sessionDates(sessions), today)
	return stats
}

func calculateBiofeedbackStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.BiofeedbackToolStats {
	stuttersPerMin := avgCollector{}
	readingWPM := avgCollector{}
	stats := statsdomain.BiofeedbackToolStats{}

	for _, session := range sessions {
		switch session.ToolType {
		case "STUTTER_TAP_COUNTER":
			stats.StutterTapSessions++
			if taps, ok := metadataNumber(session.Metadata, "total_taps"); ok && session.DurationSeconds > 0 {
				stuttersPerMin.add(taps / (float64(session.DurationSeconds) / 60))
			}
		case "TIMED_READING_WPM":
			stats.TimedReadingSessions++
			if value, ok := metadataNumber(session.Metadata, "actual_wpm"); ok {
				readingWPM.add(value)
			}
		}
	}

	stats.TotalSessions = len(sessions)
	stats.TotalMinutes = totalMinutes(sessions)
	stats.AvgRating = averageRating(sessions)
	stats.AvgStuttersPerMin = stuttersPerMin.avgPtr()
	stats.AvgReadingWPM = readingWPM.avgPtr()
	stats.CurrentStreak = currentStreak(sessionDates(sessions), today)
	return stats
}

func calculateSimulationStats(sessions []*statsdomain.ToolSession, today time.Time) statsdomain.SimulationToolStats {
	completionScore := avgCollector{}
	stats := statsdomain.SimulationToolStats{}

	for _, session := range sessions {
		switch session.ToolType {
		case "VIRTUAL_COFFEE_ORDER":
			stats.CoffeeSessions++
		case "PHONE_CALL_SIMULATOR":
			stats.CallSessions++
		}
		if completed, ok := metadataBool(session.Metadata, "completed"); ok {
			if completed {
				completionScore.add(100)
			} else {
				completionScore.add(0)
			}
		}
	}

	stats.TotalSessions = len(sessions)
	stats.TotalMinutes = totalMinutes(sessions)
	stats.AvgRating = averageRating(sessions)
	stats.AvgCompletionScore = completionScore.avgPtr()
	stats.CurrentStreak = currentStreak(sessionDates(sessions), today)
	return stats
}

func calculateWeeklyActivity(sessions []*statsdomain.ToolSession, today time.Time) []statsdomain.WeeklyActivityDay {
	counts := make(map[string]int, 7)
	for _, session := range sessions {
		key := dateKey(session.StartedAt)
		counts[key]++
	}

	activity := make([]statsdomain.WeeklyActivityDay, 0, 7)
	start := today.AddDate(0, 0, -6)
	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i)
		activity = append(activity, statsdomain.WeeklyActivityDay{
			Date:     date,
			Day:      weekdayLabel(date),
			Sessions: counts[dateKey(date)],
		})
	}
	return activity
}

func calculateWeeklyTrend(sessions []*statsdomain.ToolSession, today time.Time) []statsdomain.WeeklyTrendWeek {
	trend := make([]statsdomain.WeeklyTrendWeek, 0, 6)
	start := today.AddDate(0, 0, -35)

	for week := 0; week < 6; week++ {
		weekStart := start.AddDate(0, 0, week*7)
		weekEnd := weekStart.AddDate(0, 0, 7)
		minutes := 0.0
		for _, session := range sessions {
			sessionDate := startOfDayUTC(session.StartedAt)
			if !sessionDate.Before(weekStart) && sessionDate.Before(weekEnd) {
				minutes += float64(session.DurationSeconds) / 60
			}
		}
		trend = append(trend, statsdomain.WeeklyTrendWeek{
			WeekLabel:    fmt.Sprintf("W%d", week+1),
			TotalMinutes: round1(minutes),
		})
	}
	return trend
}

func recentSessions(sessions []*statsdomain.ToolSession, limit int) []*statsdomain.ToolSession {
	if len(sessions) <= limit {
		return sessions
	}
	return sessions[:limit]
}

func filterToolSessions(sessions []*statsdomain.ToolSession, toolTypes ...string) []*statsdomain.ToolSession {
	allowed := make(map[string]struct{}, len(toolTypes))
	for _, toolType := range toolTypes {
		allowed[toolType] = struct{}{}
	}

	filtered := make([]*statsdomain.ToolSession, 0)
	for _, session := range sessions {
		if _, ok := allowed[session.ToolType]; ok {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func totalMinutes(sessions []*statsdomain.ToolSession) float64 {
	seconds := 0
	for _, session := range sessions {
		seconds += session.DurationSeconds
	}
	return round1(float64(seconds) / 60)
}

func averageRating(sessions []*statsdomain.ToolSession) *float64 {
	avg := avgCollector{}
	for _, session := range sessions {
		if session.SelfRating != nil {
			avg.add(float64(*session.SelfRating))
		}
	}
	return avg.avgPtr()
}

func sessionsThisWeek(sessions []*statsdomain.ToolSession) int {
	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	count := 0
	for _, session := range sessions {
		if !session.StartedAt.Before(cutoff) {
			count++
		}
	}
	return count
}

func sessionDates(sessions []*statsdomain.ToolSession) map[string]time.Time {
	dates := make(map[string]time.Time)
	for _, session := range sessions {
		date := startOfDayUTC(session.StartedAt)
		dates[dateKey(date)] = date
	}
	return dates
}

func currentStreak(dates map[string]time.Time, today time.Time) int {
	streak := 0
	for date := today; ; date = date.AddDate(0, 0, -1) {
		if _, ok := dates[dateKey(date)]; !ok {
			return streak
		}
		streak++
	}
}

func bestStreak(dates map[string]time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	ordered := make([]time.Time, 0, len(dates))
	for _, date := range dates {
		ordered = append(ordered, date)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Before(ordered[j])
	})

	best := 1
	current := 1
	for i := 1; i < len(ordered); i++ {
		if ordered[i].Equal(ordered[i-1].AddDate(0, 0, 1)) {
			current++
		} else {
			current = 1
		}
		if current > best {
			best = current
		}
	}
	return best
}

func metadataNumber(metadata map[string]any, key string) (float64, bool) {
	value, ok := metadata[key]
	if !ok || value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func metadataString(metadata map[string]any, key string) (string, bool) {
	value, ok := metadata[key]
	if !ok || value == nil {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	return text, text != ""
}

func metadataBool(metadata map[string]any, key string) (bool, bool) {
	value, ok := metadata[key]
	if !ok || value == nil {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		if lower == "true" {
			return true, true
		}
		if lower == "false" {
			return false, true
		}
	}
	return false, false
}

func modeStringPtr(counts map[string]int) *string {
	var mode string
	bestCount := 0
	for value, count := range counts {
		if count > bestCount || (count == bestCount && value < mode) {
			mode = value
			bestCount = count
		}
	}
	if bestCount == 0 {
		return nil
	}
	return &mode
}

type avgCollector struct {
	sum   float64
	count int
}

func (a *avgCollector) add(value float64) {
	a.sum += value
	a.count++
}

func (a avgCollector) avgPtr() *float64 {
	if a.count == 0 {
		return nil
	}
	return floatPtr(round1(a.sum / float64(a.count)))
}

func floatPtr(value float64) *float64 {
	return &value
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func dateKey(date time.Time) string {
	return startOfDayUTC(date).Format(dateLayout)
}

func weekdayLabel(date time.Time) string {
	return string(date.Weekday().String()[0])
}
