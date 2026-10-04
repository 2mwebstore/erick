package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/seoaudit"
)

// auditTimeout bounds one crawl. A site this size takes seconds; the bound is
// for a site that has stopped answering.
const auditTimeout = 2 * time.Minute

var (
	ErrAuditNotConfigured = errors.New("SEO audit: SITE_URL is not set")
	ErrAuditRunning       = errors.New("SEO audit: one is already running")
)

// SEOAuditStatus is what the admin panel polls.
type SEOAuditStatus struct {
	Configured bool             `json:"configured"`
	Origin     string           `json:"origin,omitempty"`
	Running    bool             `json:"running"`
	Error      string           `json:"error,omitempty"`
	Report     *seoaudit.Report `json:"report,omitempty"`
}

// SEOAuditService runs the crawl in the background and keeps the latest result.
//
// A crawl takes longer than a request may stay open (the Nuxt proxy gives up
// after 8 seconds), so starting one returns at once and the panel polls. One
// runs at a time: a second request while one is running is refused, so the
// endpoint cannot be used to make the server crawl the site over and over.
type SEOAuditService struct {
	auditor *seoaudit.Auditor // nil when no site URL is configured
	log     *slog.Logger

	mu      sync.Mutex
	running bool
	report  *seoaudit.Report
	lastErr string
}

func NewSEOAuditService(auditor *seoaudit.Auditor, log *slog.Logger) *SEOAuditService {
	return &SEOAuditService{auditor: auditor, log: log}
}

// Start begins a crawl in the background.
func (s *SEOAuditService) Start() error {
	if s == nil || s.auditor == nil {
		return ErrAuditNotConfigured
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return ErrAuditRunning
	}
	s.running = true
	s.mu.Unlock()

	go s.run()
	return nil
}

func (s *SEOAuditService) run() {
	var (
		report *seoaudit.Report
		err    error
	)

	// A panic here would otherwise take the whole API down: the HTTP recovery
	// middleware only covers request goroutines.
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("the audit stopped unexpectedly: %v", p)
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.running = false
		if err != nil {
			s.lastErr = "The check did not finish: " + err.Error()
			s.log.Error("seo audit failed", slog.String("error", err.Error()))
			return
		}
		s.report, s.lastErr = report, ""
		s.log.Info("seo audit finished",
			slog.Int("score", report.Score),
			slog.Int("pages", len(report.Pages)),
			slog.Int64("ms", report.DurationMs))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
	defer cancel()
	report, err = s.auditor.Run(ctx)
}

// Status is the latest result and whether a crawl is running.
func (s *SEOAuditService) Status() SEOAuditStatus {
	if s == nil || s.auditor == nil {
		return SEOAuditStatus{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return SEOAuditStatus{
		Configured: true,
		Origin:     s.auditor.Origin(),
		Running:    s.running,
		Error:      s.lastErr,
		Report:     s.report,
	}
}
