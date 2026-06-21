package handlers

import (
	"net/http"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/services"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// LeaderboardHandler exposes the public leaderboard and platform stats.
// No auth required on either route — public by design.
type LeaderboardHandler struct {
	leaderboardService *services.LeaderboardService
}

// NewLeaderboardHandler constructs a LeaderboardHandler.
func NewLeaderboardHandler(leaderboardService *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

// GetLeaderboard handles GET /api/leaderboard.
func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries, err := h.leaderboardService.GetLeaderboard(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to load leaderboard")
		return
	}
	if entries == nil {
		entries = []models.LeaderboardEntry{} // always return [], never JSON null
	}
	utils.WriteJSON(w, http.StatusOK, entries)
}

// GetPlatformStats handles GET /api/leaderboard/stats.
func (h *LeaderboardHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.leaderboardService.GetPlatformStats(r.Context())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to load platform stats")
		return
	}
	utils.WriteJSON(w, http.StatusOK, stats)
}
