// Package router wires together every handler, middleware, and static
// file route into a single http.Handler. This is the only place that
// knows the full URL map of the app.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/yourname/dsa-tracker/internal/handlers"
	"github.com/yourname/dsa-tracker/internal/middleware"
	"github.com/yourname/dsa-tracker/internal/services"
)

// Dependencies bundles everything the router needs to construct handlers.
// Passed in from main() after all services are built.
type Dependencies struct {
	AuthService        *services.AuthService
	ChecklistService   *services.ChecklistService
	ProfileService     *services.ProfileService
	LeaderboardService *services.LeaderboardService
	IsProduction       bool
	StaticDir          string // path to the frontend static files
	AllowedOrigins     []string
	FrontendBase       string // base URL the OAuth callback redirects back to
}

// New builds and returns the fully configured router.
func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// ── Global middleware ────────────────────────────────────────────────
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true, // required for the httpOnly auth cookie to be sent cross-origin
		MaxAge:           300,
	}))

	authHandler := handlers.NewAuthHandler(deps.AuthService, deps.IsProduction, deps.FrontendBase)
	checklistHandler := handlers.NewChecklistHandler(deps.ChecklistService)
	profileHandler := handlers.NewProfileHandler(deps.ProfileService)
	leaderboardHandler := handlers.NewLeaderboardHandler(deps.LeaderboardService)

	requireAuth := middleware.AuthMiddleware(deps.AuthService)

	// ── API routes ────────────────────────────────────────────────────────
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handlers.HealthCheck)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/google/login", authHandler.GoogleLogin)
			r.Get("/google/callback", authHandler.GoogleCallback)
			r.Post("/logout", authHandler.Logout)

			r.Group(func(r chi.Router) {
				// AuthMiddleware only, deliberately NOT RequireActive — a
				// pending/blocked user must still be able to ask "who am I
				// and why am I blocked" via this endpoint.
				r.Use(requireAuth)
				r.Get("/me", authHandler.Me)
				r.Post("/college-name", authHandler.NameMyCollege)
			})
		})

		// Checklist and profile-editing require a fully active account —
		// pending/blocked users are stopped here with a 403 the frontend
		// turns into the appropriate status screen.
		r.Group(func(r chi.Router) {
			r.Use(requireAuth)
			r.Use(middleware.RequireActive)

			r.Route("/checklist", func(r chi.Router) {
				r.Get("/", checklistHandler.GetChecklist)
				r.Get("/summary", checklistHandler.GetSummary)
				r.Put("/patterns/{patternID}/progress", checklistHandler.UpdateProgress)
			})

			r.Get("/profile", profileHandler.GetMyProfile)
			r.Put("/profile", profileHandler.UpdateProfile)
		})

		// Public, no-auth-required routes: profile pages and the
		// leaderboard are intentionally visible to everyone, including
		// logged-out visitors — see product decision in project notes.
		r.Get("/users/{username}", profileHandler.GetPublicProfile)
		r.Get("/leaderboard", leaderboardHandler.GetLeaderboard)
		r.Get("/leaderboard/stats", leaderboardHandler.GetPlatformStats)
	})

	// ── Static frontend ──────────────────────────────────────────────────
	// Serves web/static (html/css/js) directly from the Go binary's host
	// filesystem. If you choose to host the frontend separately instead,
	// you can remove this block entirely — the /api routes above are
	// already a complete, CORS-enabled backend on their own.

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) { ===============================puased temporarily... causing infinite redirect loop.
	// 	http.Redirect(w, r, "/index.html", http.StatusFound)
	// })

	// /u/{username} is a real, shareable URL. profile.html is a static
	// shell — its JS reads the username from the path and calls
	// GET /api/users/{username} to render the page client-side.
	fileServer := http.FileServer(http.Dir(deps.StaticDir))
	r.Get("/u/{username}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, deps.StaticDir+"/profile.html")
	})

	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
	r.Handle("/*", fileServer)

	return r
}
