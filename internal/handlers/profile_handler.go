package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/dsa-tracker/internal/middleware"
	"github.com/yourname/dsa-tracker/internal/services"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// ProfileHandler exposes profile updates (authenticated) and public
// profile lookups (no auth required — the /u/<username> page).
type ProfileHandler struct {
	profileService *services.ProfileService
}

// NewProfileHandler constructs a ProfileHandler.
func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

type updateProfileRequest struct {
	Username    *string `json:"username"`
	LeetCodeURL *string `json:"leetcode_url"`
	IsAnonymous *bool   `json:"is_anonymous"`
}

// GetMyProfile handles GET /api/profile — returns the current user's own
// profile, fully unmasked even if is_anonymous is set. Powers the
// edit-profile screen, which must show real data to its owner regardless
// of what's hidden from everyone else.
func (h *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	profile, err := h.profileService.GetMyProfile(r.Context(), user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/profile — partial update of the current
// user's editable fields (username, leetcode_url, is_anonymous). Fields
// omitted from the request body are left unchanged.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req updateProfileRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	updated, err := h.profileService.UpdateProfile(r.Context(), user.ID, services.UpdateProfileInput{
		Username:    req.Username,
		LeetCodeURL: req.LeetCodeURL,
		IsAnonymous: req.IsAnonymous,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUsernameTaken):
			utils.WriteError(w, http.StatusConflict, err.Error())
		case errors.Is(err, services.ErrUsernameInvalid):
			utils.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			utils.WriteError(w, http.StatusInternalServerError, "failed to update profile")
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, meResponse{
		ID:            updated.ID,
		Email:         updated.Email,
		Username:      updated.Username,
		DisplayName:   updated.DisplayName,
		AvatarURL:     updated.AvatarURL,
		LeetCodeURL:   updated.LeetCodeURL,
		IsAnonymous:   updated.IsAnonymous,
		AccountStatus: string(updated.AccountStatus),
	})
}

// GetPublicProfile handles GET /api/users/{username} — the public,
// no-auth-required /u/<username> page data.
func (h *ProfileHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		utils.WriteError(w, http.StatusBadRequest, "username is required")
		return
	}

	profile, err := h.profileService.GetPublicProfile(r.Context(), username)
	if err != nil {
		if errors.Is(err, services.ErrProfileNotFound) {
			utils.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	utils.WriteJSON(w, http.StatusOK, profile)
}
