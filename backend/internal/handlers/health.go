package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"
)

// HealthHandler answers GET /health (§33).
type HealthHandler struct {
	db      *sql.DB
	log     *slog.Logger
	version string
	started time.Time
}

func NewHealthHandler(db *sql.DB, log *slog.Logger, version string) *HealthHandler {
	return &HealthHandler{db: db, log: log, version: version, started: time.Now()}
}

type healthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Uptime  string            `json:"uptime"`
	Checks  map[string]string `json:"checks"`
	Time    string            `json:"time"`
}

// Check reports liveness plus dependency status.
//
// It returns 503 when a dependency is down so an uptime monitor treats a
// database outage as an outage, rather than seeing 200 from a process that
// cannot actually serve requests.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{"api": "ok"}
	status := http.StatusOK
	overall := "ok"

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unreachable"
			overall = "degraded"
			status = http.StatusServiceUnavailable
			h.log.Error("health check: database unreachable", slog.String("error", err.Error()))
		} else {
			checks["database"] = "ok"
		}
	} else {
		checks["database"] = "disabled"
	}

	writeJSON(w, h.log, status, healthResponse{
		Status:  overall,
		Version: h.version,
		Uptime:  time.Since(h.started).Round(time.Second).String(),
		Checks:  checks,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
