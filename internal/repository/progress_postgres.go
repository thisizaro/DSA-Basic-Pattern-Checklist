package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/dsa-tracker/internal/models"
)

// PostgresProgressRepository implements ProgressRepository against a Postgres pool.
type PostgresProgressRepository struct {
	db *pgxpool.Pool
}

// NewPostgresProgressRepository constructs a PostgresProgressRepository.
func NewPostgresProgressRepository(db *pgxpool.Pool) *PostgresProgressRepository {
	return &PostgresProgressRepository{db: db}
}

func (r *PostgresProgressRepository) ListByUser(ctx context.Context, userID string) ([]models.Progress, error) {
	const query = `
		SELECT id, user_id, pattern_id, understood, explained, solved_blind, solved_after_gap, notes, updated_at
		FROM user_progress
		WHERE user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying progress by user: %w", err)
	}
	defer rows.Close()

	var result []models.Progress
	for rows.Next() {
		var p models.Progress
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.PatternID,
			&p.Understood, &p.Explained, &p.SolvedBlind, &p.SolvedAfterGap,
			&p.Notes, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning progress row: %w", err)
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating progress rows: %w", err)
	}
	return result, nil
}

// Upsert creates the progress row if it doesn't exist, or updates the
// checklist fields if it does. ON CONFLICT relies on the unique
// (user_id, pattern_id) constraint defined in the schema.
func (r *PostgresProgressRepository) Upsert(ctx context.Context, p *models.Progress) error {
	const query = `
		INSERT INTO user_progress (user_id, pattern_id, understood, explained, solved_blind, solved_after_gap, notes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id, pattern_id)
		DO UPDATE SET
			understood       = EXCLUDED.understood,
			explained        = EXCLUDED.explained,
			solved_blind     = EXCLUDED.solved_blind,
			solved_after_gap = EXCLUDED.solved_after_gap,
			notes            = EXCLUDED.notes,
			updated_at       = now()
		RETURNING id, updated_at`

	err := r.db.QueryRow(ctx, query,
		p.UserID, p.PatternID, p.Understood, p.Explained, p.SolvedBlind, p.SolvedAfterGap, p.Notes,
	).Scan(&p.ID, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting progress: %w", err)
	}
	return nil
}
