package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	therapydomain "saythis-backend/internal/src/therapy/domain"
)

// Compile-time check: PostgresTherapyRepo must satisfy TherapyRepository.
var _ TherapyRepository = (*PostgresTherapyRepo)(nil)

// PostgresTherapyRepo is the Postgres-backed implementation of TherapyRepository.
type PostgresTherapyRepo struct {
	db *pgxpool.Pool
}

func NewPostgresTherapyRepo(db *pgxpool.Pool) *PostgresTherapyRepo {
	return &PostgresTherapyRepo{db: db}
}

// UpsertExerciseProgress inserts a new exercise progress record. If a record
// already exists for the same (user_id, exercise_id) pair it updates the
// rating, remarks, and completed_at in-place — preserving the original id.
func (r *PostgresTherapyRepo) UpsertExerciseProgress(ctx context.Context, p *therapydomain.ExerciseProgress) error {
	query := `
		INSERT INTO exercise_progress (id, user_id, chapter_id, exercise_id, completed, rating, remarks, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, exercise_id) DO UPDATE
		    SET chapter_id   = EXCLUDED.chapter_id,
		        rating       = EXCLUDED.rating,
		        remarks      = EXCLUDED.remarks,
		        completed_at = EXCLUDED.completed_at
	`
	_, err := r.db.Exec(ctx, query,
		p.ID(), p.UserID(), p.ChapterID(), p.ExerciseID(),
		p.Completed(), p.Rating(), p.Remarks(), p.CompletedAt(),
	)
	if err != nil {
		return fmt.Errorf("upsert exercise progress: %w", err)
	}
	return nil
}

// GetProgressByUserID fetches every completed exercise record for the given user,
// ordered by completed_at ascending so clients can reconstruct the unlock chain.
func (r *PostgresTherapyRepo) GetProgressByUserID(ctx context.Context, userID uuid.UUID) ([]*therapydomain.ExerciseProgress, error) {
	query := `
		SELECT id, user_id, chapter_id, exercise_id, completed, rating, remarks, completed_at
		FROM exercise_progress
		WHERE user_id = $1
		ORDER BY completed_at ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*therapydomain.ExerciseProgress{}, nil
		}
		return nil, fmt.Errorf("query exercise progress: %w", err)
	}
	defer rows.Close()

	var results []*therapydomain.ExerciseProgress
	for rows.Next() {
		var (
			id          uuid.UUID
			dbUserID    uuid.UUID
			chapterID   string
			exerciseID  string
			completed   bool
			rating      int
			remarks     string
			completedAt time.Time
		)
		if err := rows.Scan(
			&id, &dbUserID, &chapterID, &exerciseID,
			&completed, &rating, &remarks, &completedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exercise progress row: %w", err)
		}
		results = append(results, therapydomain.ReconstitueExerciseProgress(
			id, dbUserID, chapterID, exerciseID, completed, rating, remarks, completedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercise progress rows: %w", err)
	}
	return results, nil
}
