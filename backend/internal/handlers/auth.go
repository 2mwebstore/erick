package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

type AuthHandler struct {
	auth       *services.AuthService
	audit      *services.AuditService
	log        *slog.Logger
	secure     bool
	trustProxy bool
}

func NewAuthHandler(auth *services.AuthService, audit *services.AuditService, log *slog.Logger, secure, trustProxy bool) *AuthHandler {
	return &AuthHandler{auth: auth, audit: audit, log: log, secure: secure, trustProxy: trustProxy}
}

// APIContract is the shape of what the admin panel and this API agree on.
//
// It is bumped whenever a request or response the admin panel depends on
// changes in a way an older build cannot satisfy. The panel refuses to trust a
// server below the version it was built against and says so, because the
// alternative is what actually happened: a server left running through a
// deploy answered a changed request body with "Malformed request.", which
// describes the symptom and not one useful thing about the cause.
//
//	1 — original CMS
//	2 — translations; PUT /v1/admin/settings takes {settings, translations}
//	3 — contact phone/subject, experience company/period/location/technologies
//	4 — principles are a CMS entity with their own admin routes
//	5 — positioning pillars likewise
const APIContract = 5

type sessionResponse struct {
	OK bool `json:"ok"`
	// Absent from any build before contract 3, which is itself the signal.
	Contract int          `json:"contract,omitempty"`
	User     *models.User `json:"user,omitempty"`
	CSRF     string       `json:"csrf,omitempty"`
	Expires  *time.Time   `json:"expires,omitempty"`
	Message  string       `json:"message,omitempty"`
}

// Login handles POST /v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Malformed request."})
		return
	}

	ip := middleware.ClientIP(r, h.trustProxy)

	session, err := h.auth.Login(r.Context(), req.Email, req.Password, ip, r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			// One message for both causes, so this cannot be used to discover
			// which email addresses have accounts.
			h.log.Warn("failed login",
				slog.String("request_id", middleware.RequestIDFrom(r.Context())),
				slog.String("ip", ip))
			writeJSON(w, h.log, http.StatusUnauthorized,
				Response{Message: "Those details do not match an account."})
		case errors.Is(err, services.ErrAccountDisabled):
			writeJSON(w, h.log, http.StatusForbidden,
				Response{Message: "That account has been disabled."})
		default:
			h.log.Error("login failed",
				slog.String("request_id", middleware.RequestIDFrom(r.Context())),
				slog.String("error", err.Error()))
			writeJSON(w, h.log, http.StatusInternalServerError,
				Response{Message: "Could not sign you in. Please try again."})
		}
		return
	}

	csrf, err := services.NewCSRFToken()
	if err != nil {
		writeJSON(w, h.log, http.StatusInternalServerError, Response{Message: "Could not start a session."})
		return
	}

	h.setCookie(w, middleware.SessionCookie, session.Token, session.Expires, true)
	h.setCookie(w, middleware.CSRFCookie, csrf, session.Expires, false)

	h.audit.Record(r.Context(), session.User, "login", "session", "", nil, ip)

	writeJSON(w, h.log, http.StatusOK, sessionResponse{
		OK: true, Contract: APIContract, User: session.User, CSRF: csrf, Expires: &session.Expires,
	})
}

// Logout handles POST /v1/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookie); err == nil {
		if err := h.auth.Logout(r.Context(), cookie.Value); err != nil {
			h.log.Error("logout failed", slog.String("error", err.Error()))
		}
	}

	if user := middleware.UserFrom(r.Context()); user != nil {
		h.audit.Record(r.Context(), user, "logout", "session", "", nil,
			middleware.ClientIP(r, h.trustProxy))
	}

	h.clearCookie(w, middleware.SessionCookie, true)
	h.clearCookie(w, middleware.CSRFCookie, false)

	writeJSON(w, h.log, http.StatusOK, Response{OK: true, Message: "Signed out."})
}

// Session handles GET /v1/auth/session, used by the admin UI on load.
//
// It reissues the CSRF token so a reload after the cookie was cleared still
// yields a working session rather than a silent 403 on the next save.
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())
	if user == nil {
		writeJSON(w, h.log, http.StatusUnauthorized, Response{Message: "Please sign in."})
		return
	}

	csrf := ""
	if cookie, err := r.Cookie(middleware.CSRFCookie); err == nil && cookie.Value != "" {
		csrf = cookie.Value
	} else if token, err := services.NewCSRFToken(); err == nil {
		csrf = token
		h.setCookie(w, middleware.CSRFCookie, token, time.Now().Add(services.SessionDuration), false)
	}

	writeJSON(w, h.log, http.StatusOK, sessionResponse{
		OK: true, Contract: APIContract, User: user, CSRF: csrf,
	})
}

// ChangePassword handles POST /v1/auth/password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFrom(r.Context())

	var req struct {
		Current string `json:"current"`
		Next    string `json:"next"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, h.log, http.StatusBadRequest, Response{Message: "Malformed request."})
		return
	}

	if err := h.auth.ChangePassword(r.Context(), user.ID, req.Current, req.Next); err != nil {
		var invalid *services.ValidationError
		switch {
		case errors.As(err, &invalid):
			writeJSON(w, h.log, http.StatusUnprocessableEntity,
				Response{Message: "Please check the form.", Errors: invalid.Fields})
		case errors.Is(err, services.ErrInvalidCredentials):
			writeJSON(w, h.log, http.StatusUnprocessableEntity, Response{
				Message: "Please check the form.",
				Errors:  map[string]string{"current": "That is not your current password."},
			})
		default:
			h.log.Error("password change failed", slog.String("error", err.Error()))
			writeJSON(w, h.log, http.StatusInternalServerError,
				Response{Message: "Could not change the password."})
		}
		return
	}

	h.audit.Record(r.Context(), user, "change_password", "user", "", nil,
		middleware.ClientIP(r, h.trustProxy))

	// Every session was revoked, including this one.
	h.clearCookie(w, middleware.SessionCookie, true)
	h.clearCookie(w, middleware.CSRFCookie, false)

	writeJSON(w, h.log, http.StatusOK, Response{
		OK: true, Message: "Password changed. Please sign in again.",
	})
}

// setCookie applies the session cookie policy in one place.
//
// SameSite=Strict because the admin is only ever reached by typing the URL or
// following a link from itself, so there is no legitimate cross-site navigation
// to preserve — and it removes most CSRF exposure on its own.
func (h *AuthHandler) setCookie(w http.ResponseWriter, name, value string, expires time.Time, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: httpOnly,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter, name string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: httpOnly,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}
