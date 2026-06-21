package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/dsa-tracker/internal/models"
)

// PostgresCatalogRepository implements CatalogRepository against a Postgres pool.
type PostgresCatalogRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCatalogRepository constructs a PostgresCatalogRepository.
func NewPostgresCatalogRepository(db *pgxpool.Pool) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

func (r *PostgresCatalogRepository) ListTopics(ctx context.Context) ([]models.Topic, error) {
	const query = `
		SELECT id, slug, title, sort_order
		FROM topics
		ORDER BY sort_order ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying topics: %w", err)
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.Slug, &t.Title, &t.SortOrder); err != nil {
			return nil, fmt.Errorf("scanning topic: %w", err)
		}
		topics = append(topics, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating topics: %w", err)
	}
	return topics, nil
}

func (r *PostgresCatalogRepository) ListPatternsByTopic(ctx context.Context, topicID string) ([]models.Pattern, error) {
	const query = `
		SELECT id, topic_id, name, core_idea, question_title, question_url, sort_order
		FROM patterns
		WHERE topic_id = $1
		ORDER BY sort_order ASC`

	rows, err := r.db.Query(ctx, query, topicID)
	if err != nil {
		return nil, fmt.Errorf("querying patterns by topic: %w", err)
	}
	defer rows.Close()

	return scanPatterns(rows)
}

func (r *PostgresCatalogRepository) ListAllPatterns(ctx context.Context) ([]models.Pattern, error) {
	const query = `
		SELECT id, topic_id, name, core_idea, question_title, question_url, sort_order
		FROM patterns
		ORDER BY sort_order ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all patterns: %w", err)
	}
	defer rows.Close()

	return scanPatterns(rows)
}

func (r *PostgresCatalogRepository) GetPattern(ctx context.Context, patternID string) (*models.Pattern, error) {
	const query = `
		SELECT id, topic_id, name, core_idea, question_title, question_url, sort_order
		FROM patterns
		WHERE id = $1`

	var p models.Pattern
	err := r.db.QueryRow(ctx, query, patternID).Scan(
		&p.ID, &p.TopicID, &p.Name, &p.CoreIdea, &p.QuestionTitle, &p.QuestionURL, &p.SortOrder,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying pattern: %w", err)
	}
	return &p, nil
}

// scanPatterns is a small helper shared by the two list methods above to
// avoid duplicating the row-scanning loop.
func scanPatterns(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]models.Pattern, error) {
	var patterns []models.Pattern
	for rows.Next() {
		var p models.Pattern
		if err := rows.Scan(
			&p.ID, &p.TopicID, &p.Name, &p.CoreIdea, &p.QuestionTitle, &p.QuestionURL, &p.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scanning pattern: %w", err)
		}
		patterns = append(patterns, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating patterns: %w", err)
	}
	return patterns, nil
}
