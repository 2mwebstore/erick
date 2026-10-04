package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// SEOHandler serves the admin SEO audit.
type SEOHandler struct {
	audit        *services.SEOAuditService
	auditLog     *services.AuditService
	log          *slog.Logger
	trustedProxy bool
}

func NewSEOHandler(audit *services.SEOAuditService, auditLog *services.AuditService, log *slog.Logger, trustedProxy bool) *SEOHandler {
	return &SEOHandler{audit: audit, auditLog: auditLog, log: log, trustedProxy: trustedProxy}
}

// Status handles GET /v1/admin/seo/audit: the latest report, and whether a
// crawl is running.
func (h *SEOHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.log, http.StatusOK, h.audit.Status())
}

// Run handles POST /v1/admin/seo/audit: starts a crawl and answers at once.
func (h *SEOHandler) Run(w http.ResponseWriter, r *http.Request) {
	switch err := h.audit.Start(); {
	case errors.Is(err, services.ErrAuditNotConfigured):
		writeJSON(w, h.log, http.StatusServiceUnavailable, Response{
			Message: "The SEO check is off: set SITE_URL on the API to the public site address.",
		})
		return
	case errors.Is(err, services.ErrAuditRunning):
		writeJSON(w, h.log, http.StatusConflict, Response{Message: "A check is already running."})
		return
	case err != nil:
		h.log.Error("starting seo audit", slog.String("error", err.Error()))
		writeJSON(w, h.log, http.StatusInternalServerError, Response{Message: "The check could not be started."})
		return
	}

	if h.auditLog != nil {
		h.auditLog.Record(r.Context(), middleware.UserFrom(r.Context()), "run", "seo_audit", "", nil,
			middleware.ClientIP(r, h.trustedProxy))
	}
	writeJSON(w, h.log, http.StatusAccepted, h.audit.Status())
}
