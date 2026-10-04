package tests

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/internal/seoaudit"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func waitFor(t *testing.T, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !done() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSEOAuditIsOffWithoutASiteURL(t *testing.T) {
	svc := services.NewSEOAuditService(nil, quietLog())

	if err := svc.Start(); !errors.Is(err, services.ErrAuditNotConfigured) {
		t.Fatalf("expected ErrAuditNotConfigured, got %v", err)
	}
	if svc.Status().Configured {
		t.Error("an audit with no site URL reported itself configured")
	}
}

// One crawl at a time: otherwise the endpoint could be used to make the server
// crawl the site in a loop.
func TestSEOAuditRunsOneAtATimeAndKeepsTheReport(t *testing.T) {
	release := make(chan struct{})
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			<-release // hold the crawl open until the test has tried a second start
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><title>Home</title></head><body><h1>Home</h1></body></html>`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(site.Close)

	auditor, err := seoaudit.New(site.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc := services.NewSEOAuditService(auditor, quietLog())

	if err := svc.Start(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if !svc.Status().Running {
		t.Error("expected the audit to report itself running")
	}
	if err := svc.Start(); !errors.Is(err, services.ErrAuditRunning) {
		t.Fatalf("a second start while running: expected ErrAuditRunning, got %v", err)
	}

	close(release)
	waitFor(t, func() bool { return !svc.Status().Running })

	status := svc.Status()
	if status.Report == nil || status.Error != "" {
		t.Fatalf("no report after the run: %+v", status)
	}
	if status.Origin != site.URL {
		t.Errorf("origin %q, want %q", status.Origin, site.URL)
	}
	if err := svc.Start(); err != nil {
		t.Errorf("a new run after the first finished was refused: %v", err)
	}
	waitFor(t, func() bool { return !svc.Status().Running })
}

func TestLoadReadsTheSiteURL(t *testing.T) {
	t.Setenv("DB_ENABLED", "false")

	t.Setenv("SITE_URL", "")
	t.Setenv("NUXT_PUBLIC_SITE_URL", "https://example.com")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SiteURL != "https://example.com" {
		t.Errorf("fallback to NUXT_PUBLIC_SITE_URL: got %q", cfg.SiteURL)
	}

	t.Setenv("SITE_URL", "https://own.example")
	if cfg, _ := config.Load(); cfg.SiteURL != "https://own.example" {
		t.Errorf("SITE_URL should win, got %q", cfg.SiteURL)
	}

	for _, bad := range []string{"example.com", "/relative", "ftp://example.com"} {
		t.Setenv("SITE_URL", bad)
		if _, err := config.Load(); err == nil {
			t.Errorf("SITE_URL=%q was accepted", bad)
		}
	}
}
