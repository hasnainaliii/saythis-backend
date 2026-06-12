package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

var _ StatsRepository = (*PostgresStatsRepo)(nil)

type PostgresStatsRepo struct {
	db *pgxpool.Pool
}

func NewPostgresStatsRepo(db *pgxpool.Pool) *PostgresStatsRepo {
	return &PostgresStatsRepo{db: db}
}

func (r *PostgresStatsRepo) UpsertDailyStat(ctx context.Context, userID uuid.UUID, patch statsdomain.DailyStatPatch) (*statsdomain.DailyStat, error) {
	columns := []string{"user_id", "date"}
	placeholders := []string{"$1", "$2"}
	updates := make([]string, 0, 12)
	args := []any{userID, patch.Date}

	addOptionalColumn(&columns, &placeholders, &updates, &args, "mood", patch.Mood)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "sleep_hours", patch.SleepHours)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "journal_entry", patch.JournalEntry)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "stress_level", patch.StressLevel)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "mindful_hours", patch.MindfulHours)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "stutter_score", patch.StutterScore)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "stutter_count", patch.StutterCount)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "repetition_count", patch.RepetitionCount)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "filler_count", patch.FillerCount)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "total_words", patch.TotalWords)
	addOptionalColumn(&columns, &placeholders, &updates, &args, "stutter_transcript", patch.StutterTranscript)

	updates = append(updates, "updated_at = NOW()")

	query := fmt.Sprintf(`
		INSERT INTO user_daily_stats (%s)
		VALUES (%s)
		ON CONFLICT (user_id, date) DO UPDATE
		SET %s
		RETURNING %s
	`, strings.Join(columns, ", "), strings.Join(placeholders, ", "), strings.Join(updates, ", "), dailyStatSelectColumns())

	stat, err := scanDailyStat(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("upsert daily stat: %w", err)
	}
	return stat, nil
}

func (r *PostgresStatsRepo) GetDailyStatsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*statsdomain.DailyStat, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM user_daily_stats
		WHERE user_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC
	`, dailyStatSelectColumns())

	rows, err := r.db.Query(ctx, query, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	stats := make([]*statsdomain.DailyStat, 0)
	for rows.Next() {
		stat, err := scanDailyStat(rows)
		if err != nil {
			return nil, fmt.Errorf("scan daily stat: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily stats: %w", err)
	}
	return stats, nil
}

func (r *PostgresStatsRepo) GetDailyStatByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*statsdomain.DailyStat, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM user_daily_stats
		WHERE user_id = $1 AND date = $2
	`, dailyStatSelectColumns())

	stat, err := scanDailyStat(r.db.QueryRow(ctx, query, userID, date))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, statsdomain.ErrDailyStatNotFound
		}
		return nil, fmt.Errorf("get daily stat: %w", err)
	}
	return stat, nil
}

func (r *PostgresStatsRepo) GetJournalEntryDates(ctx context.Context, userID uuid.UUID) ([]time.Time, error) {
	query := `
		SELECT date
		FROM user_daily_stats
		WHERE user_id = $1 AND journal_entry IS NOT NULL
		ORDER BY date DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query journal dates: %w", err)
	}
	defer rows.Close()

	dates := make([]time.Time, 0)
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("scan journal date: %w", err)
		}
		dates = append(dates, date)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal dates: %w", err)
	}
	return dates, nil
}

func (r *PostgresStatsRepo) GetToolSessions(ctx context.Context, userID uuid.UUID) ([]*statsdomain.ToolSession, error) {
	exists, err := r.toolSessionsTableExists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []*statsdomain.ToolSession{}, nil
	}

	query := `
		SELECT id, user_id, tool_type, started_at, duration_seconds, self_rating, metadata
		FROM tool_sessions
		WHERE user_id = $1
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query tool sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]*statsdomain.ToolSession, 0)
	for rows.Next() {
		session, err := scanToolSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan tool session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tool sessions: %w", err)
	}
	return sessions, nil
}

func (r *PostgresStatsRepo) toolSessionsTableExists(ctx context.Context) (bool, error) {
	var tableName pgtype.Text
	if err := r.db.QueryRow(ctx, `SELECT to_regclass('public.tool_sessions')::text`).Scan(&tableName); err != nil {
		return false, fmt.Errorf("check tool_sessions table: %w", err)
	}
	return tableName.Valid, nil
}

func addOptionalColumn[T any](columns, placeholders, updates *[]string, args *[]any, column string, value statsdomain.Optional[T]) {
	if !value.Present {
		return
	}

	*columns = append(*columns, column)
	*placeholders = append(*placeholders, fmt.Sprintf("$%d", len(*args)+1))
	*updates = append(*updates, fmt.Sprintf("%s = EXCLUDED.%s", column, column))
	if value.Value == nil {
		*args = append(*args, nil)
		return
	}
	*args = append(*args, *value.Value)
}

func dailyStatSelectColumns() string {
	return `id, user_id, date, mood, sleep_hours::float8, journal_entry, stress_level, mindful_hours::float8,
		stutter_score::float8, stutter_count, repetition_count, filler_count, total_words, stutter_transcript,
		created_at, updated_at`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDailyStat(row rowScanner) (*statsdomain.DailyStat, error) {
	var (
		stat              statsdomain.DailyStat
		mood              pgtype.Text
		sleepHours        pgtype.Float8
		journalEntry      pgtype.Text
		stressLevel       pgtype.Int2
		mindfulHours      pgtype.Float8
		stutterScore      pgtype.Float8
		stutterCount      pgtype.Int4
		repetitionCount   pgtype.Int4
		fillerCount       pgtype.Int4
		totalWords        pgtype.Int4
		stutterTranscript pgtype.Text
	)

	if err := row.Scan(
		&stat.ID,
		&stat.UserID,
		&stat.Date,
		&mood,
		&sleepHours,
		&journalEntry,
		&stressLevel,
		&mindfulHours,
		&stutterScore,
		&stutterCount,
		&repetitionCount,
		&fillerCount,
		&totalWords,
		&stutterTranscript,
		&stat.CreatedAt,
		&stat.UpdatedAt,
	); err != nil {
		return nil, err
	}

	stat.Mood = textPtr(mood)
	stat.SleepHours = float8Ptr(sleepHours)
	stat.JournalEntry = textPtr(journalEntry)
	stat.StressLevel = int2Ptr(stressLevel)
	stat.MindfulHours = float8Ptr(mindfulHours)
	stat.StutterScore = float8Ptr(stutterScore)
	stat.StutterCount = int4Ptr(stutterCount)
	stat.RepetitionCount = int4Ptr(repetitionCount)
	stat.FillerCount = int4Ptr(fillerCount)
	stat.TotalWords = int4Ptr(totalWords)
	stat.StutterTranscript = textPtr(stutterTranscript)

	return &stat, nil
}

func scanToolSession(row rowScanner) (*statsdomain.ToolSession, error) {
	var (
		session      statsdomain.ToolSession
		selfRating   pgtype.Int4
		metadataJSON []byte
	)

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.ToolType,
		&session.StartedAt,
		&session.DurationSeconds,
		&selfRating,
		&metadataJSON,
	); err != nil {
		return nil, err
	}

	session.SelfRating = int4Ptr(selfRating)
	session.Metadata = map[string]any{}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
			return nil, err
		}
	}

	return &session, nil
}

func textPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func float8Ptr(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func int2Ptr(value pgtype.Int2) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int16)
	return &v
}

func int4Ptr(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int32)
	return &v
}
