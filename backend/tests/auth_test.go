package tests

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// stubResolver stands in for AuthService so the middleware can be tested
// without a database.
type stubResolver struct {
	user *models.User
	err  error
}

func (s *stubResolver) Resolve(_ context.Context, token string) (*models.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	if token != "good-token" {
		return nil, services.ErrSessionInvalid
	}
	return s.user, nil
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
}

func TestRequireAuthRejectsMissingCookie(t *testing.T) {
	handler := middleware.RequireAuth(&stubResolver{}, quietLogger())(okHandler())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/admin/content", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthRejectsUnknownToken(t *testing.T) {
	handler := middleware.RequireAuth(&stubResolver{}, quietLogger())(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/content", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: "forged"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthAttachesUser(t *testing.T) {
	user := &models.User{ID: 1, Email: "editor@example.com", Role: models.RoleEditor, IsActive: true}

	var seen *models.User
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = middleware.UserFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequireAuth(&stubResolver{user: user}, quietLogger())(inner)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/content", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: "good-token"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if seen == nil || seen.Email != user.Email {
		t.Fatalf("expected the user on the context, got %#v", seen)
	}
}

func TestRequireAdminBlocksEditors(t *testing.T) {
	cases := map[string]struct {
		role models.Role
		want int
	}{
		"editor is refused": {models.RoleEditor, http.StatusForbidden},
		"admin is allowed":  {models.RoleAdmin, http.StatusOK},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			user := &models.User{ID: 1, Email: "user@example.com", Role: tc.role, IsActive: true}
			handler := middleware.RequireAuth(&stubResolver{user: user}, quietLogger())(
				middleware.RequireAdmin(okHandler()),
			)

			req := httptest.NewRequest(http.MethodGet, "/v1/admin/users", nil)
			req.AddCookie(&http.Cookie{Name: middleware.SessionCookie, Value: "good-token"})

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rec.Code)
			}
		})
	}
}

func TestCSRFAllowsSafeMethods(t *testing.T) {
	handler := middleware.CSRF(okHandler())

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(method, "/v1/admin/content", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", method, rec.Code)
		}
	}
}

func TestCSRFRejectsWritesWithoutAMatchingToken(t *testing.T) {
	handler := middleware.CSRF(okHandler())

	cases := map[string]func(*http.Request){
		"no cookie, no header": func(_ *http.Request) {},
		"cookie but no header": func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: middleware.CSRFCookie, Value: "token"})
		},
		"header but no cookie": func(r *http.Request) {
			r.Header.Set(middleware.CSRFHeader, "token")
		},
		"mismatched": func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: middleware.CSRFCookie, Value: "token"})
			r.Header.Set(middleware.CSRFHeader, "different")
		},
	}

	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/projects", strings.NewReader("{}"))
			setup(req)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d", rec.Code)
			}
		})
	}
}

func TestCSRFAcceptsMatchingToken(t *testing.T) {
	handler := middleware.CSRF(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/projects", strings.NewReader("{}"))
	req.AddCookie(&http.Cookie{Name: middleware.CSRFCookie, Value: "matching-token"})
	req.Header.Set(middleware.CSRFHeader, "matching-token")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCSRFMatchesRejectsEmptyValues(t *testing.T) {
	// An empty cookie must never satisfy an empty header, or a request with
	// neither would pass.
	if services.CSRFMatches("", "") {
		t.Error("two empty tokens must not match")
	}
	if services.CSRFMatches("token", "") || services.CSRFMatches("", "token") {
		t.Error("a missing side must not match")
	}
	if !services.CSRFMatches("token", "token") {
		t.Error("identical tokens must match")
	}
}

func TestPasswordPolicy(t *testing.T) {
	cases := map[string]struct {
		password string
		valid    bool
	}{
		"long passphrase with a space": {"correct horse battery staple", true},
		"letters and digits":           {"Portfolio2026Pass", true},
		"too short":                    {"Short1!", false},
		"exactly the minimum of 8":     {"abcd1234", true},
		"one under the minimum":        {"abc1234", false},
		"letters only":                 {"abcdefghijklmnop", false},
		"digits only":                  {"12345678901234", false},
		// bcrypt silently truncates beyond 72 bytes, so a longer password must be
		// refused rather than quietly shortened.
		"beyond bcrypt's limit": {strings.Repeat("a1", 40), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := services.ValidatePassword(tc.password)
			if tc.valid && err != nil {
				t.Fatalf("expected %q to be accepted, got %v", tc.password, err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected %q to be rejected", tc.password)
			}
		})
	}
}

func TestHashPasswordProducesVerifiableDistinctHashes(t *testing.T) {
	first, err := services.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}
	second, _ := services.HashPassword("correct horse battery staple")

	if first == second {
		t.Error("identical passwords must not produce identical hashes — the salt is missing")
	}
	if !strings.HasPrefix(first, "$2a$12$") && !strings.HasPrefix(first, "$2b$12$") {
		t.Errorf("expected a bcrypt cost-12 hash, got %q", first[:7])
	}
}

func TestNewCSRFTokenIsUnpredictable(t *testing.T) {
	seen := map[string]bool{}
	for range 50 {
		token, err := services.NewCSRFToken()
		if err != nil {
			t.Fatalf("generating token: %v", err)
		}
		if len(token) < 32 {
			t.Fatalf("token too short: %d characters", len(token))
		}
		if seen[token] {
			t.Fatal("generated a duplicate token")
		}
		seen[token] = true
	}
}

func TestAdminRoutesAreUnreachableWithoutASession(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	routes := []struct{ method, path string }{
		{http.MethodGet, "/v1/admin/content"},
		{http.MethodPost, "/v1/admin/projects"},
		{http.MethodPut, "/v1/admin/projects/1"},
		{http.MethodDelete, "/v1/admin/projects/1"},
		{http.MethodPost, "/v1/admin/experience"},
		{http.MethodPost, "/v1/admin/capabilities"},
		{http.MethodPost, "/v1/admin/services"},
		{http.MethodPut, "/v1/admin/settings"},
		{http.MethodGet, "/v1/admin/messages"},
		{http.MethodGet, "/v1/admin/messages/export"},
		{http.MethodGet, "/v1/admin/users"},
		{http.MethodGet, "/v1/admin/audit"},
		{http.MethodGet, "/v1/admin/seo/audit"},
		{http.MethodPost, "/v1/admin/seo/audit"},
		{http.MethodGet, "/v1/auth/session"},
		{http.MethodPost, "/v1/auth/logout"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}

			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("expected a JSON body, got %q", rec.Body.String())
			}
			if body["ok"] != false {
				t.Errorf("expected ok=false, got %v", body["ok"])
			}
		})
	}
}
