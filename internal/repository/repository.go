// Package repository defines the data-access contracts used by services.
// Services depend only on these interfaces, never on pgx directly — so the
// storage backend (Postgres today, anything else tomorrow) can be swapped
// or mocked in tests without touching business logic.
package repository

import (
	"context"

	"github.com/yourname/dsa-tracker/internal/models"
)

// UserRepository handles persistence for user accounts.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// UsernameTaken reports whether username is already in use by a
	// *different* user than excludeUserID (pass "" when checking at
	// signup, when there's no existing user to exclude).
	UsernameTaken(ctx context.Context, username string, excludeUserID string) (bool, error)

	// Update persists changes to mutable profile fields: username,
	// leetcode_url, is_anonymous. Other fields (email, google_id) are
	// intentionally not touched here — they change through other, more
	// deliberate paths (OAuth login, admin review).
	Update(ctx context.Context, user *models.User) error

	// UpdateAccountStatus persists just the account_status transition
	// (e.g. blocked -> active when a dev flag is set, or pending -> active
	// when a college gets approved). Kept separate from Update so a
	// profile-edit code path can never accidentally change access status.
	UpdateAccountStatus(ctx context.Context, userID string, status models.AccountStatus) error
}

// CatalogRepository handles read access to the static topic/pattern catalog.
// "Catalog" because this data is the same for every user — it's the content
// of the revision sheet itself, not anyone's personal progress.
type CatalogRepository interface {
	ListTopics(ctx context.Context) ([]models.Topic, error)
	ListPatternsByTopic(ctx context.Context, topicID string) ([]models.Pattern, error)
	ListAllPatterns(ctx context.Context) ([]models.Pattern, error)
	GetPattern(ctx context.Context, patternID string) (*models.Pattern, error)
}

// ProgressRepository handles persistence for per-user checklist progress.
type ProgressRepository interface {
	// ListByUser returns every progress row a user has (i.e. every pattern
	// they've interacted with at least once).
	ListByUser(ctx context.Context, userID string) ([]models.Progress, error)

	// Upsert creates or updates the progress row for (userID, patternID).
	Upsert(ctx context.Context, p *models.Progress) error
}

// CollegeRepository handles persistence for the domain→college mapping.
type CollegeRepository interface {
	// GetByDomain looks up a college by exact domain match (e.g. "kiit.ac.in").
	GetByDomain(ctx context.Context, domain string) (*models.College, error)

	// GetByID looks up a college by its primary key — used when building a
	// profile from a user row that only stores college_id.
	GetByID(ctx context.Context, id string) (*models.College, error)

	// CreatePending inserts a new domain as a pending-review college. Used
	// the first time someone signs in with an unrecognized domain.
	CreatePending(ctx context.Context, domain string) (*models.College, error)

	// SetName updates a pending college's display name — used when the
	// first user from that domain names their college (it starts out
	// defaulted to the raw domain string). Only takes effect while the
	// college is still pending, so it can't be used to rename an
	// already-approved/reviewed college through the public API.
	SetName(ctx context.Context, collegeID string, name string) error
}

// StatsRepository handles read-only aggregate queries that span all users —
// the leaderboard and public platform stats. Kept separate from
// ProgressRepository/UserRepository because these are cross-cutting
// aggregation queries, not single-entity CRUD.
type StatsRepository interface {
	// Leaderboard returns one row per user with progress, sorted by
	// patterns solved (desc), then steps completed (desc).
	Leaderboard(ctx context.Context) ([]models.LeaderboardEntry, error)

	// PlatformStats returns the aggregate numbers shown on the public
	// leaderboard page (total users, total solved, signups by day, etc).
	PlatformStats(ctx context.Context) (*models.PlatformStats, error)
}
