package domain

import (
	"time"

	"github.com/google/uuid"
)

// ToolSession is the read model used to aggregate already-recorded practice sessions.
type ToolSession struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ToolType        string
	StartedAt       time.Time
	DurationSeconds int
	SelfRating      *int
	Metadata        map[string]any
}
