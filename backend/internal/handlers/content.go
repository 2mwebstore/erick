package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/kongchansila/portfolio/backend/internal/services"
)

// ContentHandler serves the public content payload.
type ContentHandler struct {
	content *services.ContentService
	log     *slog.Logger
}

func NewContentHandler(content *services.ContentService, log *slog.Logger) *ContentHandler {
	return &ContentHandler{content: content, log: log}
}

// Site handles GET /v1/content.
//
// Unpublished projects are excluded here and included only on the admin route,
// so a draft cannot leak through the public cache.
func (h *ContentHandler) Site(w http.ResponseWriter, r *http.Request) {
	// An unknown or missing locale falls back to English inside the service
	// rather than erroring: a bad ?locale= should not be able to blank the site.
	content, err := h.content.Site(r.Context(), false, r.URL.Query().Get("locale"))
	if err != nil {
		h.log.Error("loading site content", slog.String("error", err.Error()))
		writeJSON(w, h.log, http.StatusServiceUnavailable,
			Response{Message: "Content is temporarily unavailable."})
		return
	}

	// Short cache with a longer stale window: the frontend and any CDN in front
	// can keep serving the last good copy while this refreshes, so an API blip
	// does not blank the site.
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=600")
	w.Header().Set("Last-Modified", content.GeneratedAt.Format(http.TimeFormat))
	// The locale lives in the query string, so it is already part of any cache
	// key; this is for anything keying on content negotiation instead.
	w.Header().Set("Vary", "Accept-Language")

	writeJSON(w, h.log, http.StatusOK, content)
}

// parseID reads a path value as a positive integer id.
func parseID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
