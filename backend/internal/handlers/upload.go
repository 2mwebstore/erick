package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/services"
	"github.com/kongchansila/portfolio/backend/internal/uploads"
)

// UploadHandler stores images for the admin panel's image fields.
type UploadHandler struct {
	uploader     *uploads.Uploader // nil when R2 is not configured
	images       *services.ImageCleanup
	auditLog     *services.AuditService
	log          *slog.Logger
	trustedProxy bool
}

func NewUploadHandler(uploader *uploads.Uploader, images *services.ImageCleanup, auditLog *services.AuditService, log *slog.Logger, trustedProxy bool) *UploadHandler {
	return &UploadHandler{uploader: uploader, images: images, auditLog: auditLog, log: log, trustedProxy: trustedProxy}
}

// Discard handles POST /v1/admin/uploads/discard with {"url": "…"}: an upload
// replaced or cleared before it was ever saved, which no save would otherwise
// release. The usual rules hold — only a file this site uploaded, and only when
// nothing uses it — so naming a file that is saved somewhere deletes nothing.
func (h *UploadHandler) Discard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := decode(r, &body); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	h.images.Release(middleware.UserFrom(r.Context()), middleware.ClientIP(r, h.trustedProxy),
		"upload discarded before saving", body.URL)
	writeJSON(w, h.log, http.StatusAccepted, Response{OK: true, Message: "Discarded."})
}

type uploadSettings struct {
	Enabled  bool     `json:"enabled"`
	MaxBytes int64    `json:"maxBytes,omitempty"`
	Types    []string `json:"types,omitempty"`
}

// Settings handles GET /v1/admin/uploads: whether uploads are on, and what
// they accept. Image fields offer only a link when they are off.
func (h *UploadHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if h.uploader == nil {
		writeJSON(w, h.log, http.StatusOK, uploadSettings{})
		return
	}
	writeJSON(w, h.log, http.StatusOK, uploadSettings{Enabled: true, MaxBytes: h.uploader.MaxBytes(), Types: uploads.Types})
}

type uploadResponse struct {
	Response
	*uploads.Image
}

// Image handles POST /v1/admin/uploads/image, a multipart form with a
// "folder" field (projects, portraits, profiles) and a "file" field.
func (h *UploadHandler) Image(w http.ResponseWriter, r *http.Request) {
	if h.uploader == nil {
		writeJSON(w, h.log, http.StatusServiceUnavailable, Response{
			Message: "Image uploads are off: set the R2_* variables on the API. You can still paste an image link.",
		})
		return
	}

	parts, err := r.MultipartReader()
	if err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Send the image as a multipart form."})
		return
	}

	// Read both fields before uploading, so their order in the form does not matter.
	var folder string
	var file []byte
	for {
		part, err := parts.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			h.bodyError(w, err)
			return
		}
		switch part.FormName() {
		case "folder":
			b, _ := io.ReadAll(io.LimitReader(part, 64))
			folder = strings.TrimSpace(string(b))
		case "file":
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, io.LimitReader(part, h.uploader.MaxBytes()+1)); err != nil {
				h.bodyError(w, err)
				return
			}
			file = buf.Bytes()
		}
		_ = part.Close()
	}
	if file == nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{
			Message: "Choose an image to upload.", Errors: map[string]string{"file": "Choose an image to upload."},
		})
		return
	}

	img, err := h.uploader.Upload(r.Context(), folder, bytes.NewReader(file))
	switch {
	case errors.Is(err, uploads.ErrFolder):
		writeJSON(w, h.log, http.StatusBadRequest, Response{
			Message: "Unknown upload folder.", Errors: map[string]string{"folder": "Unknown upload folder."},
		})
		return
	case errors.Is(err, uploads.ErrTooLarge):
		h.tooLarge(w)
		return
	case errors.Is(err, uploads.ErrNotImage), errors.Is(err, uploads.ErrEmpty):
		msg := "Only JPEG, PNG, WebP, GIF or AVIF images can be uploaded."
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: msg, Errors: map[string]string{"file": msg}})
		return
	case err != nil:
		h.log.Error("storing upload", slog.String("error", err.Error()))
		writeJSON(w, h.log, http.StatusBadGateway, Response{Message: "The image could not be stored. Please try again."})
		return
	}

	if h.auditLog != nil {
		h.auditLog.Record(r.Context(), middleware.UserFrom(r.Context()), "upload", "image", img.Key,
			map[string]any{"type": img.ContentType, "size": img.Size}, middleware.ClientIP(r, h.trustedProxy))
	}
	writeJSON(w, h.log, http.StatusCreated, uploadResponse{Response: Response{OK: true, Message: "Uploaded."}, Image: img})
}

func (h *UploadHandler) bodyError(w http.ResponseWriter, err error) {
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		h.tooLarge(w)
		return
	}
	writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "The upload could not be read. Please try again."})
}

func (h *UploadHandler) tooLarge(w http.ResponseWriter) {
	msg := "Images can be up to " + humanBytes(h.uploader.MaxBytes()) + "."
	writeJSON(w, h.log, http.StatusRequestEntityTooLarge, Response{Message: msg, Errors: map[string]string{"file": msg}})
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<20 && n%(1<<20) == 0:
		return fmt.Sprintf("%d MB", n>>20)
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	default:
		return fmt.Sprintf("%d KB", (n+1023)>>10)
	}
}
