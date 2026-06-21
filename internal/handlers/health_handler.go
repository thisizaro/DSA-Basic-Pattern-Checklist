package handlers

import (
	"net/http"

	"github.com/yourname/dsa-tracker/internal/utils"
)

// HealthCheck handles GET /api/health. Most hosting platforms (Render,
// Fly.io, Railway, etc.) ping an endpoint like this to confirm the service
// is alive before routing traffic to it.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
