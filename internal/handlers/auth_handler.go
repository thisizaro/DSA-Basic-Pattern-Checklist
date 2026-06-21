// Package handlers contains HTTP handlers. Handlers parse requests, call
// a service, and write a response — they hold no business logic themselves.
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/yourname/dsa-tracker/internal/middleware"
	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/services"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// oauthStateCookie is the short-lived cookie holding the CSRF state value
// between the redirect-to-Google step and the callback step.
const oauthStateCookie = "oauth_state"

// AuthHandler exposes the Google OAuth login flow and "who am I" endpoint.
// There is no password path — Google is the sole identity provider.
type AuthHandler struct {
	authService  *services.AuthService
	isProd       bool   // controls the Secure flag on cookies
	frontendBase string // where to redirect the browser after login completes
}

// NewAuthHandler constructs an AuthHandler. frontendBase is the URL to
// send the browser back to once login succeeds or fails (e.g. "" for
// same-origin relative redirects, or a full origin if the frontend is
// hosted separately).
func NewAuthHandler(authService *services.AuthService, isProd bool, frontendBase string) *AuthHandler {
	return &AuthHandler{authService: authService, isProd: isProd, frontendBase: frontendBase}
}

// GoogleLogin handles GET /api/auth/google/login — generates a CSRF state,
// stores it in a short-lived cookie, and redirects the browser to Google's
// consent screen.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := utils.RandomToken(24)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "failed to start login")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes — plenty of time to complete the Google consent screen
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, h.authService.GoogleAuthCodeURL(state), http.StatusFound)
}

// GoogleCallback handles GET /api/auth/google/callback — Google redirects
// here with ?code=...&state=.... Verifies state against the cookie set in
// GoogleLogin, exchanges the code, finds-or-creates the local user, sets
// the session cookie, then redirects back into the app.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oauthStateCookie)
	if err != nil || r.URL.Query().Get("state") != stateCookie.Value {
		h.redirectWithError(w, r, "invalid_state")
		return
	}
	// State is single-use — clear it immediately regardless of outcome.
	http.SetCookie(w, &http.Cookie{
		Name: oauthStateCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.isProd, SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectWithError(w, r, "missing_code")
		return
	}

	user, token, err := h.authService.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		if errors.Is(err, services.ErrEmailNotVerified) {
			h.redirectWithError(w, r, "email_not_verified")
			return
		}
		h.redirectWithError(w, r, "login_failed")
		return
	}

	h.setAuthCookie(w, token)

	// Send active users to the checklist; anyone else to the home page,
	// which reads /api/auth/me and renders the right pending/blocked screen.
	dest := h.frontendBase + "/index.html"
	if user.AccountStatus == models.AccountActive {
		dest = h.frontendBase + "/checklist.html"
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

func (h *AuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, reason string) {
	http.Redirect(w, r, h.frontendBase+"/index.html?auth_error="+reason, http.StatusFound)
}

// Logout handles POST /api/auth/logout by clearing the auth cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// meResponse is the shape returned by GET /api/auth/me. account_status is
// included explicitly as a reminder that the frontend's first job after
// login is to branch on this field.
type meResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	AvatarURL     string `json:"avatar_url"`
	LeetCodeURL   string `json:"leetcode_url"`
	IsAnonymous   bool   `json:"is_anonymous"`
	AccountStatus string `json:"account_status"`
}

// Me handles GET /api/auth/me — returns the currently authenticated user,
// regardless of account status. Mounted behind AuthMiddleware only (NOT
// RequireActive) so pending/blocked users can still discover their status.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	utils.WriteJSON(w, http.StatusOK, meResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		AvatarURL:     user.AvatarURL,
		LeetCodeURL:   user.LeetCodeURL,
		IsAnonymous:   user.IsAnonymous,
		AccountStatus: string(user.AccountStatus),
	})
}

type nameCollegeRequest struct {
	Name string `json:"name"`
}

// NameMyCollege handles POST /api/auth/college-name — lets a
// pending_review user supply their college's display name. Mounted behind
// AuthMiddleware only (not RequireActive), since pending users are
// precisely who needs to reach this endpoint.
func (h *AuthHandler) NameMyCollege(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req nameCollegeRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "malformed request body")
		return
	}

	err := h.authService.NameMyPendingCollege(r.Context(), user.ID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCollegeNameInvalid):
			utils.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrNoCollegeToName):
			utils.WriteError(w, http.StatusConflict, err.Error())
		default:
			utils.WriteError(w, http.StatusInternalServerError, "failed to save college name")
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// setAuthCookie stores the JWT in an httpOnly cookie so it can't be read by
// client-side JS (mitigates XSS token theft).
func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
}
