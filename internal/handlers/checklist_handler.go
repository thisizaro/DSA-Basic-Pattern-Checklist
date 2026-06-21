package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/dsa-tracker/internal/middleware"
	"github.com/yourname/dsa-tracker/internal/services"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// ChecklistHandler exposes the topic/pattern checklist and progress updates.
// Every route here is mounted behind AuthMiddleware.
type ChecklistHandler struct {
	checklistService *services.ChecklistService
}

// NewChecklistHandler constructs a ChecklistHandler.
func NewChecklistHandler(checklistService *services.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{checklistService: checklistService}
}

// GetChecklist handles GET /api/checklist — returns every topic with its
// patterns and the current user's progress on each.
func (h *ChecklistHandler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	checklist, err := h.checklistService.GetFullChecklist(r.Context(), user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to load checklist")
		return
	}

	utils.WriteJSON(w, http.StatusOK, checklist)
}

// GetSummary handles GET /api/checklist/summary — returns aggregate stats
// used for a progress bar / completion count on the UI.
func (h *ChecklistHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	summary, err := h.checklistService.GetSummary(r.Context(), user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to load summary")
		return
	}

	utils.WriteJSON(w, http.StatusOK, summary)
}

type updateProgressRequest struct {
	Understood     bool   `json:"understood"`
	Explained      bool   `json:"explained"`
	SolvedBlind    bool   `json:"solved_blind"`
	SolvedAfterGap bool   `json:"solved_after_gap"`
	Notes          string `json:"notes"`
}

// UpdateProgress handles PUT /api/checklist/patterns/{patternID}/progress
// Saves the checklist state for one pattern for the current user.
func (h *ChecklistHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	patternID := chi.URLParam(r, "patternID")
	if patternID == "" {
		utils.WriteError(w, http.StatusBadRequest, "pattern id is required")
		return
	}

	var req updateProgressRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	progress, err := h.checklistService.UpdateProgress(
		r.Context(), user.ID, patternID,
		req.Understood, req.Explained, req.SolvedBlind, req.SolvedAfterGap, req.Notes,
	)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "pattern not found or update failed")
		return
	}

	utils.WriteJSON(w, http.StatusOK, progress)
}
