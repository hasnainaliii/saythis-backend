package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

// StatsRepository defines persistence needed by the stats use cases.
type StatsRepository interface {
	UpsertDailyStat(ctx context.Context, userID uuid.UUID, patch statsdomain.DailyStatPatch) (*statsdomain.DailyStat, error)
	GetDailyStatsByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*statsdomain.DailyStat, error)
	GetDailyStatByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*statsdomain.DailyStat, error)
	GetJournalEntryDates(ctx context.Context, userID uuid.UUID) ([]time.Time, error)
	GetToolSessions(ctx context.Context, userID uuid.UUID) ([]*statsdomain.ToolSession, error)
}
