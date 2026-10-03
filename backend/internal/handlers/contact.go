package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

type ContactHandler struct {
	service    *services.ContactService
	log        *slog.Logger
	trustProxy bool
}

func NewContactHandler(service *services.ContactService, log *slog.Logger, trustProxy bool) *ContactHandler {
	return &ContactHandler{service: service, log: log, trustProxy: trustProxy}
}

// Create handles POST /v1/contact.
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "" && !isJSON(ct) {
		writeJSON(w, h.log, http.StatusUnsupportedMediaType, Response{
			Message: "Expected application/json.",
		})
		return
	}

	var req models.ContactRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // reject padded payloads rather than silently ignoring them

	if err := decoder.Decode(&req); err != nil {
		status := http.StatusBadRequest
		message := "Malformed request."

		// MaxBytesReader surfaces as a generic error; distinguish it so the
		// visitor gets an accurate reason.
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) || errors.Is(err, io.ErrUnexpectedEOF) {
			status = http.StatusRequestEntityTooLarge
			message = "That message is too large to send."
		}

		writeJSON(w, h.log, status, Response{Message: message})
		return
	}

	ip := middleware.ClientIP(r, h.trustProxy)

	msg, err := h.service.Submit(r.Context(), req, ip, r.UserAgent())
	if err != nil {
		var invalid *services.ValidationError
		if errors.As(err, &invalid) {
			writeJSON(w, h.log, http.StatusUnprocessableEntity, Response{
				Message: "Please check the highlighted fields.",
				Errors:  invalid.Fields,
			})
			return
		}

		// Anything else is our fault. It is logged in full and reported
		// generically, so no internal detail reaches the client (§34).
		h.log.Error("contact submission failed",
			slog.String("request_id", middleware.RequestIDFrom(r.Context())),
			slog.String("error", err.Error()),
		)

		writeJSON(w, h.log, http.StatusInternalServerError, Response{
			Message: "The message could not be saved. Please try again shortly.",
		})
		return
	}

	h.log.Info("contact message stored",
		slog.String("request_id", middleware.RequestIDFrom(r.Context())),
		slog.Int64("id", msg.ID),
		slog.String("project_type", msg.ProjectType),
	)

	writeJSON(w, h.log, http.StatusCreated, Response{
		OK:      true,
		Message: "Thanks — your message has been sent. I'll reply by email.",
	})
}

// isJSON ignores any charset or boundary parameter after the media type.
func isJSON(contentType string) bool {
	mediaType, _, _ := strings.Cut(contentType, ";")
	return strings.EqualFold(strings.TrimSpace(mediaType), "application/json")
}
