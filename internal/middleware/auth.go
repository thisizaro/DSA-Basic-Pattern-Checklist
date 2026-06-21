// Package middleware holds HTTP middleware shared across routes:
// authentication, request logging, and panic recovery.
package middleware

import (
	"context"
	"net/http"

	"github.com/yourname/dsa-tracker/internal/models"
	"github.com/yourname/dsa-tracker/internal/services"
	"github.com/yourname/dsa-tracker/internal/utils"
)

// contextKey is a private type to avoid collisions with other packages'
// context keys, per Go's standard advice for context.WithValue.
type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware reads the "auth_token" cookie, validates it via
// AuthService, and stores the resolved user on the request context.
// Requests without a valid token get a 401 and never reach the handler.
func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			user, err := authService.ValidateToken(r.Context(), cookie.Value)
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext retrieves the authenticated user stored by AuthMiddleware.
// Safe to call only on routes mounted behind AuthMiddleware.
func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

// RequireActive blocks any request from a user whose account isn't fully
// active (pending_review or blocked_unrecognized_domain) with a 403 and a
// machine-readable reason the frontend uses to render the right screen.
// Must be mounted *after* AuthMiddleware, since it reads the user from
// context that AuthMiddleware populates.
//
// GET /api/auth/me intentionally does NOT sit behind this middleware —
// it's how the frontend discovers *why* a non-active user is blocked in
// the first place, so it can't itself be blocked.
func RequireActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if user.IsBlocked() {
			utils.WriteJSON(w, http.StatusForbidden, map[string]string{
				"error":          "account not active",
				"account_status": string(user.AccountStatus),
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
