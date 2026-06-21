package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/dsa-tracker/internal/models"
)

// PostgresCollegeRepository implements CollegeRepository against a Postgres pool.
type PostgresCollegeRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCollegeRepository constructs a PostgresCollegeRepository.
func NewPostgresCollegeRepository(db *pgxpool.Pool) *PostgresCollegeRepository {
	return &PostgresCollegeRepository{db: db}
}

func (r *PostgresCollegeRepository) GetByDomain(ctx context.Context, domain string) (*models.College, error) {
	const query = `SELECT id, domain, name, status, created_at FROM colleges WHERE domain = $1`

	var c models.College
	err := r.db.QueryRow(ctx, query, domain).Scan(&c.ID, &c.Domain, &c.Name, &c.Status, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying college by domain: %w", err)
	}
	return &c, nil
}

func (r *PostgresCollegeRepository) GetByID(ctx context.Context, id string) (*models.College, error) {
	const query = `SELECT id, domain, name, status, created_at FROM colleges WHERE id = $1`

	var c models.College
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Domain, &c.Name, &c.Status, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying college by id: %w", err)
	}
	return &c, nil
}

// SetName updates a pending college's display name. Restricted to
// `status = 'pending'` in the WHERE clause so it can never be used to
// rename a college that's already been through manual review — at that
// point the name is considered finalized and only direct DB access should
// change it.
func (r *PostgresCollegeRepository) SetName(ctx context.Context, collegeID string, name string) error {
	const query = `UPDATE colleges SET name = $1 WHERE id = $2 AND status = 'pending'`

	tag, err := r.db.Exec(ctx, query, name, collegeID)
	if err != nil {
		return fmt.Errorf("setting college name: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreatePending inserts domain as a brand new pending-review college. The
// display name defaults to the raw domain itself (e.g. "xyz.edu") until
// someone — the first user from that domain, per the product's "be the
// first to enter the name" flow, or you on manual review — sets a proper
// display name and flips status to approved directly in the DB.
//
// ON CONFLICT handles the race where two people from the same brand-new
// domain sign up at nearly the same time — whoever wins the insert race,
// the other request just reads back the same row instead of erroring.
func (r *PostgresCollegeRepository) CreatePending(ctx context.Context, domain string) (*models.College, error) {
	const query = `
		INSERT INTO colleges (domain, name, status)
		VALUES ($1, $1, 'pending')
		ON CONFLICT (domain) DO UPDATE SET domain = EXCLUDED.domain
		RETURNING id, domain, name, status, created_at`

	var c models.College
	err := r.db.QueryRow(ctx, query, domain).Scan(&c.ID, &c.Domain, &c.Name, &c.Status, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating pending college: %w", err)
	}
	return &c, nil
}
