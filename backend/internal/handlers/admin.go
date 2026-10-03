package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// AdminHandler is the authenticated management surface (§28).
type AdminHandler struct {
	content    *services.ContentService
	messages   *repositories.ContactRepository
	auth       *services.AuthService
	audit      *services.AuditService
	log        *slog.Logger
	trustProxy bool
}

func NewAdminHandler(
	content *services.ContentService,
	messages *repositories.ContactRepository,
	auth *services.AuthService,
	audit *services.AuditService,
	log *slog.Logger,
	trustProxy bool,
) *AdminHandler {
	return &AdminHandler{
		content: content, messages: messages, auth: auth,
		audit: audit, log: log, trustProxy: trustProxy,
	}
}

func (h *AdminHandler) ip(r *http.Request) string {
	return middleware.ClientIP(r, h.trustProxy)
}

// fail maps a service error onto a response, keeping internals out of the body.
func (h *AdminHandler) fail(w http.ResponseWriter, r *http.Request, err error, action string) {
	var invalid *services.ValidationError

	switch {
	case errors.As(err, &invalid):
		writeJSON(w, h.log, http.StatusUnprocessableEntity,
			Response{Message: "Please correct the highlighted fields.", Errors: invalid.Fields})
	case errors.Is(err, repositories.ErrNotFound):
		writeJSON(w, h.log, http.StatusNotFound, Response{Message: "That item no longer exists."})
	default:
		h.log.Error(action,
			slog.String("request_id", middleware.RequestIDFrom(r.Context())),
			slog.String("error", err.Error()))
		writeJSON(w, h.log, http.StatusInternalServerError,
			Response{Message: "Something went wrong saving that. Please try again."})
	}
}

func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// decodeMessage turns a decoding failure into something the editor can act on.
//
// "Malformed request." is true and useless. These routes are behind a session,
// so naming the field or the position gives the person editing a way forward
// without telling an anonymous visitor anything. The most common cause in
// practice is a version skew — a browser tab loaded before a deploy sending a
// body the new handler does not recognise, or the reverse — and that is
// exactly the case the old message could not describe.
func decodeMessage(err error) string {
	var syntax *json.SyntaxError
	var typeErr *json.UnmarshalTypeError

	switch {
	case errors.Is(err, io.EOF):
		return "The request body was empty."
	case errors.Is(err, io.ErrUnexpectedEOF):
		// A body that stopped mid-value: usually a dropped connection rather
		// than a client sending nonsense.
		return "The request body is not valid JSON — it ended unexpectedly."
	case errors.As(err, &syntax):
		return fmt.Sprintf("The request body is not valid JSON (at byte %d).", syntax.Offset)
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return fmt.Sprintf("Field %q has the wrong type — expected %s.", typeErr.Field, typeErr.Type)
		}
		return "A value in the request has the wrong type."
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return "The request contained an unexpected field: " + field +
			". This usually means the page was loaded before the server was updated — reload and try again."
	default:
		return "Malformed request."
	}
}

// SavePillar handles POST /v1/admin/pillars and PUT /v1/admin/pillars/{id}.
func (h *AdminHandler) SavePillar(w http.ResponseWriter, r *http.Request) {
	var pillar models.Pillar
	if err := decode(r, &pillar); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	pillar.ID = pathID(r)

	id, err := h.content.SavePillar(r.Context(), &pillar)
	if err != nil {
		h.fail(w, r, err, "saving pillar")
		return
	}
	pillar.ID = id

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), saveAction(r), "pillar",
		pillar.Slug, map[string]any{"title": pillar.Title}, h.ip(r))

	writeJSON(w, h.log, saveStatus(r), map[string]any{"ok": true, "pillar": pillar})
}

func (h *AdminHandler) DeletePillar(w http.ResponseWriter, r *http.Request) {
	h.deleteEntity(w, r, "pillar", h.content.DeletePillar)
}

// SavePrinciple handles POST /v1/admin/principles and PUT /v1/admin/principles/{id}.
func (h *AdminHandler) SavePrinciple(w http.ResponseWriter, r *http.Request) {
	var principle models.Principle
	if err := decode(r, &principle); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	principle.ID = pathID(r)

	id, err := h.content.SavePrinciple(r.Context(), &principle)
	if err != nil {
		h.fail(w, r, err, "saving principle")
		return
	}
	principle.ID = id

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), saveAction(r), "principle",
		principle.Slug, map[string]any{"title": principle.Title}, h.ip(r))

	writeJSON(w, h.log, saveStatus(r), map[string]any{"ok": true, "principle": principle})
}

func (h *AdminHandler) DeletePrinciple(w http.ResponseWriter, r *http.Request) {
	h.deleteEntity(w, r, "principle", h.content.DeletePrinciple)
}

// ── Content ─────────────────────────────────────────────────────────────────

// Content handles GET /v1/admin/content, including unpublished projects.
func (h *AdminHandler) Content(w http.ResponseWriter, r *http.Request) {
	// Admin gets English plus every translation, so the editor can show both
	// languages side by side.
	content, err := h.content.SiteForAdmin(r.Context())
	if err != nil {
		h.fail(w, r, err, "loading admin content")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, h.log, http.StatusOK, content)
}

// SaveProject handles POST /v1/admin/projects and PUT /v1/admin/projects/{id}.
func (h *AdminHandler) SaveProject(w http.ResponseWriter, r *http.Request) {
	var project models.Project
	if err := decode(r, &project); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}

	// The id always comes from the path, never the body, so a payload cannot
	// redirect a save onto a different row.
	if raw := r.PathValue("id"); raw != "" {
		id, ok := parseID(r, "id")
		if !ok {
			writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Invalid id."})
			return
		}
		project.ID = id
	} else {
		project.ID = 0
	}

	id, err := h.content.SaveProject(r.Context(), &project)
	if err != nil {
		h.fail(w, r, err, "saving project")
		return
	}
	project.ID = id

	action := "update"
	status := http.StatusOK
	if r.Method == http.MethodPost {
		action, status = "create", http.StatusCreated
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), action, "project",
		project.Slug, map[string]any{"title": project.Title, "published": project.Published}, h.ip(r))

	writeJSON(w, h.log, status, map[string]any{"ok": true, "project": project})
}

// DeleteProject handles DELETE /v1/admin/projects/{id}.
func (h *AdminHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Invalid id."})
		return
	}

	if err := h.content.DeleteProject(r.Context(), id); err != nil {
		h.fail(w, r, err, "deleting project")
		return
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "delete", "project",
		strconv.FormatInt(id, 10), nil, h.ip(r))

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Project deleted."})
}

// SaveExperience handles POST and PUT for experience entries.
func (h *AdminHandler) SaveExperience(w http.ResponseWriter, r *http.Request) {
	var entry models.ExperienceEntry
	if err := decode(r, &entry); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	entry.ID = pathID(r)

	id, err := h.content.SaveExperience(r.Context(), &entry)
	if err != nil {
		h.fail(w, r, err, "saving experience")
		return
	}
	entry.ID = id

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), saveAction(r), "experience",
		strconv.FormatInt(id, 10), map[string]any{"title": entry.Title}, h.ip(r))

	writeJSON(w, h.log, saveStatus(r), map[string]any{"ok": true, "entry": entry})
}

func (h *AdminHandler) DeleteExperience(w http.ResponseWriter, r *http.Request) {
	h.deleteEntity(w, r, "experience", h.content.DeleteExperience)
}

// SaveCapability handles POST and PUT for capabilities.
func (h *AdminHandler) SaveCapability(w http.ResponseWriter, r *http.Request) {
	var capability models.Capability
	if err := decode(r, &capability); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	capability.ID = pathID(r)

	id, err := h.content.SaveCapability(r.Context(), &capability)
	if err != nil {
		h.fail(w, r, err, "saving capability")
		return
	}
	capability.ID = id

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), saveAction(r), "capability",
		capability.Slug, map[string]any{"title": capability.Title}, h.ip(r))

	writeJSON(w, h.log, saveStatus(r), map[string]any{"ok": true, "capability": capability})
}

func (h *AdminHandler) DeleteCapability(w http.ResponseWriter, r *http.Request) {
	h.deleteEntity(w, r, "capability", h.content.DeleteCapability)
}

// SaveService handles POST and PUT for services.
func (h *AdminHandler) SaveService(w http.ResponseWriter, r *http.Request) {
	var service models.Service
	if err := decode(r, &service); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}
	service.ID = pathID(r)

	id, err := h.content.SaveService(r.Context(), &service)
	if err != nil {
		h.fail(w, r, err, "saving service")
		return
	}
	service.ID = id

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), saveAction(r), "service",
		service.Slug, map[string]any{"title": service.Title}, h.ip(r))

	writeJSON(w, h.log, saveStatus(r), map[string]any{"ok": true, "service": service})
}

func (h *AdminHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	h.deleteEntity(w, r, "service", h.content.DeleteService)
}

// SaveSettings handles PUT /v1/admin/settings.
func (h *AdminHandler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Settings     map[string]string            `json:"settings"`
		Translations map[string]map[string]string `json:"translations"`
	}
	if err := decode(r, &body); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}

	// A body with no `settings` object used to decode cleanly, save nothing,
	// and answer "Settings saved." — the shape changed when translations were
	// added, and an older client silently lost every edit it made. Saying so is
	// the whole point.
	if body.Settings == nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{
			Message: "The request had no settings to save. Reload the page and try again.",
		})
		return
	}

	if err := h.content.SaveSettings(r.Context(), body.Settings); err != nil {
		h.fail(w, r, err, "saving settings")
		return
	}
	if err := h.content.SaveSettingTranslations(r.Context(), body.Translations); err != nil {
		h.fail(w, r, err, "saving setting translations")
		return
	}

	keys := make([]string, 0, len(body.Settings))
	for key := range body.Settings {
		keys = append(keys, key)
	}
	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "update", "settings", "",
		map[string]any{"keys": keys}, h.ip(r))

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Settings saved."})
}

// ── Messages ────────────────────────────────────────────────────────────────

// Messages handles GET /v1/admin/messages.
func (h *AdminHandler) Messages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := repositories.MessageFilter{
		UnreadOnly: query.Get("unread") == "1",
		Archived:   query.Get("archived") == "1",
	}
	filter.Limit, _ = strconv.Atoi(query.Get("limit"))
	filter.Offset, _ = strconv.Atoi(query.Get("offset"))

	listing, err := h.messages.List(r.Context(), filter)
	if err != nil {
		h.fail(w, r, err, "listing messages")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, h.log, http.StatusOK, listing)
}

// UpdateMessage handles PATCH /v1/admin/messages/{id}.
func (h *AdminHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Invalid id."})
		return
	}

	var req struct {
		Read     *bool `json:"read"`
		Archived *bool `json:"archived"`
	}
	if err := decode(r, &req); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}

	if req.Read != nil {
		if err := h.messages.SetRead(r.Context(), id, *req.Read); err != nil {
			h.fail(w, r, err, "updating message")
			return
		}
	}
	if req.Archived != nil {
		if err := h.messages.SetArchived(r.Context(), id, *req.Archived); err != nil {
			h.fail(w, r, err, "updating message")
			return
		}
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "update", "message",
		strconv.FormatInt(id, 10), map[string]any{"read": req.Read, "archived": req.Archived}, h.ip(r))

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Message updated."})
}

// DeleteMessage handles DELETE /v1/admin/messages/{id}.
func (h *AdminHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Invalid id."})
		return
	}

	if err := h.messages.Delete(r.Context(), id); err != nil {
		h.fail(w, r, err, "deleting message")
		return
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "delete", "message",
		strconv.FormatInt(id, 10), nil, h.ip(r))

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Message deleted."})
}

// ExportMessages handles GET /v1/admin/messages/export and streams CSV.
func (h *AdminHandler) ExportMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.messages.All(r.Context())
	if err != nil {
		h.fail(w, r, err, "exporting messages")
		return
	}

	filename := fmt.Sprintf("contact-messages-%s.csv", time.Now().UTC().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"id", "received", "name", "email", "phone", "subject", "project_type", "message", "read", "ip_address"})

	for _, m := range messages {
		read := "no"
		if m.ReadAt != nil {
			read = m.ReadAt.Format(time.RFC3339)
		}
		_ = writer.Write([]string{
			strconv.FormatInt(m.ID, 10),
			m.CreatedAt.Format(time.RFC3339),
			m.Name, m.Email, m.Phone, m.Subject, m.ProjectType, m.Message, read, m.IPAddress,
		})
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "export", "messages", "",
		map[string]any{"count": len(messages)}, h.ip(r))
}

// ── Users and audit ─────────────────────────────────────────────────────────

// Users handles GET /v1/admin/users. Admin only.
func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	users, err := h.auth.ListUsers(r.Context())
	if err != nil {
		h.fail(w, r, err, "listing users")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, h.log, http.StatusOK, map[string]any{"ok": true, "users": users})
}

// CreateUser handles POST /v1/admin/users. Admin only.
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := decode(r, &req); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: decodeMessage(err)})
		return
	}

	user, err := h.auth.CreateUser(r.Context(), req.Email, req.Name, req.Password, models.Role(req.Role))
	if err != nil {
		h.fail(w, r, err, "creating user")
		return
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "create", "user",
		user.Email, map[string]any{"role": user.Role}, h.ip(r))

	writeJSON(w, h.log, http.StatusCreated, map[string]any{"ok": true, "user": user})
}

// Audit handles GET /v1/admin/audit. Admin only.
func (h *AdminHandler) Audit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	listing, err := h.audit.Recent(r.Context(), limit, offset)
	if err != nil {
		h.fail(w, r, err, "listing audit log")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, h.log, http.StatusOK, map[string]any{
		"ok": true, "entries": listing.Entries, "total": listing.Total,
	})
}

// ── Shared ──────────────────────────────────────────────────────────────────

func (h *AdminHandler) deleteEntity(w http.ResponseWriter, r *http.Request, entity string, remove func(ctx context.Context, id int64) error) {
	id, ok := parseID(r, "id")
	if !ok {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Invalid id."})
		return
	}

	if err := remove(r.Context(), id); err != nil {
		h.fail(w, r, err, "deleting "+entity)
		return
	}

	h.audit.Record(r.Context(), middleware.UserFrom(r.Context()), "delete", entity,
		strconv.FormatInt(id, 10), nil, h.ip(r))

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Deleted."})
}

func pathID(r *http.Request) int64 {
	if id, ok := parseID(r, "id"); ok {
		return id
	}
	return 0
}

func saveAction(r *http.Request) string {
	if r.Method == http.MethodPost {
		return "create"
	}
	return "update"
}

func saveStatus(r *http.Request) int {
	if r.Method == http.MethodPost {
		return http.StatusCreated
	}
	return http.StatusOK
}
