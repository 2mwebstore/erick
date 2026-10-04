// Package routes wires handlers and middleware into one http.Handler.
package routes

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/internal/handlers"
	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

type Dependencies struct {
	Config         config.Config
	Logger         *slog.Logger
	DB             *sql.DB
	ContactService *services.ContactService
	ContentService *services.ContentService
	AuthService    *services.AuthService
	AuditService   *services.AuditService
	SEOAudit       *services.SEOAuditService
	Messages       *repositories.ContactRepository
	Version        string
	RateLimiter    *middleware.RateLimiter
	LoginLimiter   *middleware.RateLimiter
}

// New builds the router.
//
// Note what is absent: there is no CORS middleware. The browser never calls this
// service directly — the Nuxt server proxies to it over the container network —
// so exposing a cross-origin policy would only widen the surface (§34).
func New(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	contact := handlers.NewContactHandler(deps.ContactService, deps.Logger, deps.Config.TrustedProxy)
	health := handlers.NewHealthHandler(deps.DB, deps.Logger, deps.Version)
	content := handlers.NewContentHandler(deps.ContentService, deps.Logger)
	auth := handlers.NewAuthHandler(
		deps.AuthService, deps.AuditService, deps.Logger,
		deps.Config.SecureCookies, deps.Config.TrustedProxy,
	)
	admin := handlers.NewAdminHandler(
		deps.ContentService, deps.Messages, deps.AuthService,
		deps.AuditService, deps.Logger, deps.Config.TrustedProxy,
	)
	seo := handlers.NewSEOHandler(deps.SEOAudit, deps.AuditService, deps.Logger, deps.Config.TrustedProxy)

	// ── Public ───────────────────────────────────────────────────────────────

	// Health is deliberately outside every limiter: an uptime monitor polling
	// each minute must never be throttled into a false alarm.
	mux.HandleFunc("GET /health", health.Check)

	// Content is read on every page render, so it gets no rate limit either —
	// the Nuxt server caches it, and throttling it would break the site rather
	// than protect it.
	mux.HandleFunc("GET /v1/content", content.Site)

	contactChain := middleware.Chain(
		middleware.MaxBody(deps.Config.MaxBodyBytes),
		middleware.Limit(deps.RateLimiter, deps.Config.TrustedProxy, deps.Logger),
	)
	mux.Handle("POST /v1/contact", contactChain(http.HandlerFunc(contact.Create)))

	// Login is rate limited separately and much more tightly than the contact
	// form: this is the one endpoint where guessing is the attack.
	loginChain := middleware.Chain(
		middleware.MaxBody(8*1024),
		middleware.Limit(deps.LoginLimiter, deps.Config.TrustedProxy, deps.Logger),
	)
	mux.Handle("POST /v1/auth/login", loginChain(http.HandlerFunc(auth.Login)))

	// ── Authenticated ────────────────────────────────────────────────────────

	// Order matters: the session is resolved first, then CSRF is checked, then
	// the body size is bounded before any handler decodes it.
	protected := middleware.Chain(
		middleware.MaxBody(deps.Config.MaxAdminBodyBytes),
		middleware.RequireAuth(deps.AuthService, deps.Logger),
		middleware.CSRF,
	)

	adminOnly := middleware.Chain(
		middleware.MaxBody(deps.Config.MaxAdminBodyBytes),
		middleware.RequireAuth(deps.AuthService, deps.Logger),
		middleware.CSRF,
		middleware.RequireAdmin,
	)

	protect := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, protected(handler))
	}
	protectAdmin := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, adminOnly(handler))
	}

	protect("GET /v1/auth/session", auth.Session)
	protect("POST /v1/auth/logout", auth.Logout)
	protect("POST /v1/auth/password", auth.ChangePassword)

	protect("GET /v1/admin/content", admin.Content)

	protect("POST /v1/admin/projects", admin.SaveProject)
	protect("PUT /v1/admin/projects/{id}", admin.SaveProject)
	protect("DELETE /v1/admin/projects/{id}", admin.DeleteProject)

	protect("POST /v1/admin/experience", admin.SaveExperience)
	protect("PUT /v1/admin/experience/{id}", admin.SaveExperience)
	protect("DELETE /v1/admin/experience/{id}", admin.DeleteExperience)

	protect("POST /v1/admin/capabilities", admin.SaveCapability)
	protect("PUT /v1/admin/capabilities/{id}", admin.SaveCapability)
	protect("DELETE /v1/admin/capabilities/{id}", admin.DeleteCapability)

	protect("POST /v1/admin/services", admin.SaveService)
	protect("PUT /v1/admin/services/{id}", admin.SaveService)
	protect("DELETE /v1/admin/services/{id}", admin.DeleteService)

	protect("POST /v1/admin/principles", admin.SavePrinciple)
	protect("PUT /v1/admin/principles/{id}", admin.SavePrinciple)
	protect("DELETE /v1/admin/principles/{id}", admin.DeletePrinciple)

	protect("POST /v1/admin/pillars", admin.SavePillar)
	protect("PUT /v1/admin/pillars/{id}", admin.SavePillar)
	protect("DELETE /v1/admin/pillars/{id}", admin.DeletePillar)

	protect("PUT /v1/admin/settings", admin.SaveSettings)

	// Read-only diagnostics, so editors get it too. It only ever crawls the
	// configured site, and one run at a time.
	protect("GET /v1/admin/seo/audit", seo.Status)
	protect("POST /v1/admin/seo/audit", seo.Run)

	protect("GET /v1/admin/messages", admin.Messages)
	protect("GET /v1/admin/messages/export", admin.ExportMessages)
	protect("PATCH /v1/admin/messages/{id}", admin.UpdateMessage)
	protect("DELETE /v1/admin/messages/{id}", admin.DeleteMessage)

	// Managing who has access is an admin concern, not an editor one.
	protectAdmin("GET /v1/admin/users", admin.Users)
	protectAdmin("POST /v1/admin/users", admin.CreateUser)
	protectAdmin("GET /v1/admin/audit", admin.Audit)

	mux.HandleFunc("/", notFound)

	global := middleware.Chain(
		middleware.RequestID,
		middleware.Recover(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.SecurityHeaders,
	)

	return global(mux)
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"ok":false,"message":"Not found."}`))
}
