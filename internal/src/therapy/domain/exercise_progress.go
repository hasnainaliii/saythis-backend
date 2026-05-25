package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExerciseProgress records that a user has completed a specific exercise within
// a therapy chapter. It maps 1-to-1 with the exercise_progress table.
// A record only ever exists when the exercise is completed — completed is always true.
type ExerciseProgress struct {
	id          uuid.UUID
	userID      uuid.UUID
	chapterID   string
	exerciseID  string
	completed   bool
	rating      int
	remarks     string
	completedAt time.Time
}

// NewExerciseProgress constructs a fresh ExerciseProgress for a first-time or
// re-submission. completed is always set to true — a record only exists here
// because the exercise was finished.
func NewExerciseProgress(
	userID uuid.UUID,
	chapterID, exerciseID string,
	rating int,
	remarks string,
	now time.Time,
) *ExerciseProgress {
	return &ExerciseProgress{
		id:          uuid.New(),
		userID:      userID,
		chapterID:   chapterID,
		exerciseID:  exerciseID,
		completed:   true,
		rating:      rating,
		remarks:     remarks,
		completedAt: now,
	}
}

// ReconstitueExerciseProgress rebuilds an ExerciseProgress from a database row.
// All persisted fields are accepted directly — no re-validation because data
// coming from the DB is already trusted.
func ReconstitueExerciseProgress(
	id, userID uuid.UUID,
	chapterID, exerciseID string,
	completed bool,
	rating int,
	remarks string,
	completedAt time.Time,
) *ExerciseProgress {
	return &ExerciseProgress{
		id:          id,
		userID:      userID,
		chapterID:   chapterID,
		exerciseID:  exerciseID,
		completed:   completed,
		rating:      rating,
		remarks:     remarks,
		completedAt: completedAt,
	}
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (e *ExerciseProgress) ID() uuid.UUID          { return e.id }
func (e *ExerciseProgress) UserID() uuid.UUID      { return e.userID }
func (e *ExerciseProgress) ChapterID() string      { return e.chapterID }
func (e *ExerciseProgress) ExerciseID() string     { return e.exerciseID }
func (e *ExerciseProgress) Completed() bool        { return e.completed }
func (e *ExerciseProgress) Rating() int            { return e.rating }
func (e *ExerciseProgress) Remarks() string        { return e.remarks }
func (e *ExerciseProgress) CompletedAt() time.Time { return e.completedAt }
