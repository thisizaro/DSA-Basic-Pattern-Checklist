package services

import (
	"context"
	"fmt"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/repository"
)

// LeaderboardService exposes the public leaderboard and platform-wide
// stats. Both are public-by-design (no auth required) per product
// decision, so this service does no per-viewer access checks — anonymity
// masking happens once, here, for every viewer alike.
type LeaderboardService struct {
	stats repository.StatsRepository
}

// NewLeaderboardService constructs a LeaderboardService.
func NewLeaderboardService(stats repository.StatsRepository) *LeaderboardService {
	return &LeaderboardService{stats: stats}
}

// GetLeaderboard returns every active user's ranking. Anonymous users keep
// their row (so they still occupy their earned rank) but have their
// identity fields blanked.
func (s *LeaderboardService) GetLeaderboard(ctx context.Context) ([]models.LeaderboardEntry, error) {
	entries, err := s.stats.Leaderboard(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading leaderboard: %w", err)
	}

	for i := range entries {
		if entries[i].IsAnonymous {
			entries[i].Username = ""
			entries[i].DisplayName = "Anonymous"
			entries[i].AvatarURL = ""
			entries[i].CollegeName = ""
		}
	}

	return entries, nil
}

// GetPlatformStats returns the public, project-wide numbers shown at the
// top of the leaderboard page.
func (s *LeaderboardService) GetPlatformStats(ctx context.Context) (*models.PlatformStats, error) {
	stats, err := s.stats.PlatformStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading platform stats: %w", err)
	}
	return stats, nil
}
