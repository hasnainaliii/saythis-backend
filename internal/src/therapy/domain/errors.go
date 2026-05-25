package domain

import "errors"

var (
	// Input validation errors — returned by the use-case layer before any DB call.
	ErrInvalidChapterID  = errors.New("chapter_id must not be empty")
	ErrInvalidExerciseID = errors.New("exercise_id must not be empty")
	ErrInvalidRating     = errors.New("rating must be between 1 and 5")
)
