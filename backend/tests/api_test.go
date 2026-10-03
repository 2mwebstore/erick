package tests

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/routes"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

func testServer(t *testing.T, repo services.ContactRepository, rateLimit int) (http.Handler, *middleware.RateLimiter) {
	t.Helper()

	cfg := config.Config{
		Env:               "test",
		MaxBodyBytes:      16 * 1024,
		MaxAdminBodyBytes: 256 * 1024,
		RateLimitRequests: rateLimit,
		RateLimitWindow:   time.Hour,
		LoginRateLimit:    10,
		LoginRateWindow:   time.Hour,
		TrustedProxy:      true,
	}

	limiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	t.Cleanup(limiter.Close)

	loginLimiter := middleware.NewRateLimiter(cfg.LoginRateLimit, cfg.LoginRateWindow)
	t.Cleanup(loginLimiter.Close)

	handler := routes.New(routes.Dependencies{
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:             nil,
		ContactService: services.NewContactService(repo),
		ContentService: services.NewContentService(nil, nil),
		Version:        "test",
		RateLimiter:    limiter,
		LoginLimiter:   loginLimiter,
	})

	return handler, limiter
}

func post(t *testing.T, handler http.Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.7:54321"
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

const validBody = `{"name":"Test Visitor","email":"visitor@example.com","project_type":"Web Application","message":"We need an internal ordering system for about forty staff."}`

func TestContactAcceptsValidSubmission(t *testing.T) {
	repo := &stubRepo{}
	handler, _ := testServer(t, repo, 10)

	rec := post(t, handler, validBody, nil)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if res["ok"] != true {
		t.Errorf("expected ok=true, got %v", res["ok"])
	}
	if len(repo.created) != 1 {
		t.Errorf("expected the message to be stored, got %d", len(repo.created))
	}
}

func TestContactReturnsFieldErrors(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	rec := post(t, handler, `{"name":"A","email":"nope","project_type":"","message":"short"}`, nil)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}

	var res struct {
		OK     bool              `json:"ok"`
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if res.OK {
		t.Error("expected ok=false")
	}
	for _, field := range []string{"name", "email", "project_type", "message"} {
		if _, ok := res.Errors[field]; !ok {
			t.Errorf("expected an error for %q, got %v", field, res.Errors)
		}
	}
}

func TestContactRejectsUnknownFields(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	rec := post(t, handler, `{"name":"Test Visitor","email":"visitor@example.com","project_type":"Other","message":"A long enough message here.","is_admin":true}`, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown field, got %d", rec.Code)
	}
}

func TestContactRejectsNonJSONContentType(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	rec := post(t, handler, validBody, map[string]string{"Content-Type": "text/plain"})

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestContactRejectsOversizedBody(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	oversized := `{"name":"Test Visitor","email":"visitor@example.com","project_type":"Other","message":"` +
		strings.Repeat("a", 20*1024) + `"}`

	rec := post(t, handler, oversized, nil)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestContactNeverLeaksInternalErrors(t *testing.T) {
	handler, _ := testServer(t, nil, 10) // no repository configured

	rec := post(t, handler, validBody, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, leak := range []string{"no storage", "sql", "contact:", "panic"} {
		if strings.Contains(strings.ToLower(body), leak) {
			t.Errorf("response leaks internal detail %q: %s", leak, body)
		}
	}
}

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 2)

	for i := 1; i <= 2; i++ {
		if rec := post(t, handler, validBody, nil); rec.Code != http.StatusCreated {
			t.Fatalf("request %d: expected 201, got %d", i, rec.Code)
		}
	}

	rec := post(t, handler, validBody, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on the third request, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected a Retry-After header on 429")
	}
}

func TestRateLimitIsPerClientAddress(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 1)

	first := post(t, handler, validBody, map[string]string{"X-Real-IP": "198.51.100.1"})
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 for the first address, got %d", first.Code)
	}

	// Same address again: blocked.
	if again := post(t, handler, validBody, map[string]string{"X-Real-IP": "198.51.100.1"}); again.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for a repeat from the same address, got %d", again.Code)
	}

	// A different address has its own window.
	other := post(t, handler, validBody, map[string]string{"X-Real-IP": "198.51.100.2"})
	if other.Code != http.StatusCreated {
		t.Fatalf("expected 201 for a different address, got %d", other.Code)
	}
}

func TestSecurityHeadersPresent(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	rec := post(t, handler, validBody, nil)

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-store",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s: expected %q, got %q", header, value, got)
		}
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("expected a Content-Security-Policy header")
	}
}

func TestRequestIDIsReturned(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	rec := post(t, handler, validBody, nil)
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected an X-Request-ID header")
	}

	// A caller-supplied id is preserved so a report can be traced end to end.
	rec = post(t, handler, validBody, map[string]string{"X-Request-ID": "abc123"})
	if got := rec.Header().Get("X-Request-ID"); got != "abc123" {
		t.Errorf("expected the supplied request id to be echoed, got %q", got)
	}
}

func TestHealthReportsOKWithoutDatabase(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if res.Status != "ok" {
		t.Errorf("expected status ok, got %q", res.Status)
	}
	if res.Checks["database"] != "disabled" {
		t.Errorf("expected database=disabled, got %q", res.Checks["database"])
	}
}

func TestHealthIsNotRateLimited(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 1)

	// An uptime monitor polls far more often than the contact limit allows.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("health poll %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestUnknownRouteReturns404JSON(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected a JSON 404, got content-type %q", ct)
	}
}

func TestWrongMethodIsRejected(t *testing.T) {
	handler, _ := testServer(t, &stubRepo{}, 10)

	req := httptest.NewRequest(http.MethodGet, "/v1/contact", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatalf("GET /v1/contact must not succeed, got %d", rec.Code)
	}
}
