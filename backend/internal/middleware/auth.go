package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

const (
	userKey contextKey = "auth_user"

	// SessionCookie holds the opaque session token; CSRFCookie holds the
	// double-submit token, which must be readable by the client and so is not
	// HttpOnly.
	SessionCookie = "portfolio_session"
	CSRFCookie    = "portfolio_csrf"
	CSRFHeader    = "X-CSRF-Token"
)

// SessionResolver is the slice of AuthService this middleware needs.
type SessionResolver interface {
	Resolve(ctx context.Context, token string) (*models.User, error)
}

// UserFrom returns the authenticated user attached by RequireAuth.
func UserFrom(ctx context.Context) *models.User {
	if u, ok := ctx.Value(userKey).(*models.User); ok {
		return u
	}
	return nil
}

// RequireAuth rejects a request without a valid session.
func RequireAuth(auth SessionResolver, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookie)
			if err != nil || cookie.Value == "" {
				unauthorised(w)
				return
			}

			user, err := auth.Resolve(r.Context(), cookie.Value)
			if err != nil {
				if !errors.Is(err, services.ErrSessionInvalid) {
					log.Error("resolving session",
						slog.String("request_id", RequestIDFrom(r.Context())),
						slog.String("error", err.Error()))
				}
				unauthorised(w)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		})
	}
}

// RequireAdmin must be applied after RequireAuth.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFrom(r.Context())
		if user == nil || !user.Role.CanManageUsers() {
			writeJSON(w, http.StatusForbidden,
				`{"ok":false,"message":"This action requires an admin account."}`)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CSRF guards cookie-authenticated state-changing requests using the
// double-submit pattern: the token is set in a readable cookie and must be
// echoed in a header. A cross-site form post can send the cookie but cannot read
// it to set the header, and SameSite=Strict on the session cookie is the second
// layer.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(CSRFCookie)
		if err != nil || !services.CSRFMatches(cookie.Value, r.Header.Get(CSRFHeader)) {
			writeJSON(w, http.StatusForbidden,
				`{"ok":false,"message":"Your session has expired. Reload the page and try again."}`)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unauthorised(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, `{"ok":false,"message":"Please sign in."}`)
}
