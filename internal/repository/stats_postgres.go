package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/dsa-tracker/internal/models"
)

// PostgresStatsRepository implements StatsRepository against a Postgres pool.
// Both queries here are intentionally raw aggregate SQL rather than
// in-memory Go loops — they need to scan every user's progress, and the
// database is far better at that aggregation than pulling everything into
// the app and summing it in Go.
type PostgresStatsRepository struct {
	db *pgxpool.Pool
}

// NewPostgresStatsRepository constructs a PostgresStatsRepository.
func NewPostgresStatsRepository(db *pgxpool.Pool) *PostgresStatsRepository {
	return &PostgresStatsRepository{db: db}
}

// Leaderboard computes, per user, how many patterns they've fully
// completed (all 4 checklist steps) and how many total steps they've
// checked across all patterns. Sorted by patterns solved desc, then steps
// desc, then joined date asc (earlier joiners rank first on a dead-even tie).
func (r *PostgresStatsRepository) Leaderboard(ctx context.Context) ([]models.LeaderboardEntry, error) {
	const query = `
		SELECT
			u.username,
			u.display_name,
			u.avatar_url,
			COALESCE(c.name, '') AS college_name,
			u.is_anonymous,
			COALESCE(SUM(CASE WHEN up.understood AND up.explained AND up.solved_blind AND up.solved_after_gap THEN 1 ELSE 0 END), 0) AS patterns_solved,
			COALESCE(SUM(
				CASE WHEN up.understood THEN 1 ELSE 0 END +
				CASE WHEN up.explained THEN 1 ELSE 0 END +
				CASE WHEN up.solved_blind THEN 1 ELSE 0 END +
				CASE WHEN up.solved_after_gap THEN 1 ELSE 0 END
			), 0) AS steps_completed
		FROM users u
		LEFT JOIN user_progress up ON up.user_id = u.id
		LEFT JOIN colleges c ON c.id = u.college_id
		WHERE u.account_status = 'active'
		GROUP BY u.id, u.username, u.display_name, u.avatar_url, c.name, u.is_anonymous, u.created_at
		ORDER BY patterns_solved DESC, steps_completed DESC, u.created_at ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(
			&e.Username, &e.DisplayName, &e.AvatarURL, &e.CollegeName,
			&e.IsAnonymous, &e.PatternsSolved, &e.StepsCompleted,
		); err != nil {
			return nil, fmt.Errorf("scanning leaderboard row: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating leaderboard rows: %w", err)
	}
	return entries, nil
}

// PlatformStats computes the public, project-wide numbers: total active
// users, total patterns fully solved across everyone, total checklist
// steps checked across everyone, total approved colleges represented, and
// a day-by-day signup count for the last 30 days.
func (r *PostgresStatsRepository) PlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	stats := &models.PlatformStats{}

	const totalsQuery = `
		SELECT
			(SELECT COUNT(*) FROM users WHERE account_status = 'active') AS total_users,
			(SELECT COUNT(*) FROM user_progress WHERE understood AND explained AND solved_blind AND solved_after_gap) AS total_patterns_solved,
			(SELECT COALESCE(SUM(
				CASE WHEN understood THEN 1 ELSE 0 END +
				CASE WHEN explained THEN 1 ELSE 0 END +
				CASE WHEN solved_blind THEN 1 ELSE 0 END +
				CASE WHEN solved_after_gap THEN 1 ELSE 0 END
			), 0) FROM user_progress) AS total_steps_completed,
			(SELECT COUNT(*) FROM colleges WHERE status = 'approved') AS total_colleges`

	err := r.db.QueryRow(ctx, totalsQuery).Scan(
		&stats.TotalUsers, &stats.TotalPatternsSolved, &stats.TotalStepsCompleted, &stats.TotalColleges,
	)
	if err != nil {
		return nil, fmt.Errorf("querying platform totals: %w", err)
	}

	const signupsQuery = `
		SELECT to_char(created_at::date, 'YYYY-MM-DD') AS day, COUNT(*)
		FROM users
		WHERE account_status = 'active' AND created_at >= now() - interval '30 days'
		GROUP BY day
		ORDER BY day ASC`

	rows, err := r.db.Query(ctx, signupsQuery)
	if err != nil {
		return nil, fmt.Errorf("querying signups by day: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d models.DailySignupCount
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, fmt.Errorf("scanning signup day: %w", err)
		}
		stats.SignupsByDay = append(stats.SignupsByDay, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating signup days: %w", err)
	}

	return stats, nil
}
