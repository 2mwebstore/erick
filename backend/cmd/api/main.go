// Command api is the portfolio backend: a contact endpoint and a health check.
//
// Scope is deliberately small. The site's content is prerendered by the
// frontend, so the only thing that genuinely needs a server is the contact form
// (§29). Adding more here would be complexity without a requirement.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/internal/database"
	"github.com/kongchansila/portfolio/backend/internal/middleware"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
	"github.com/kongchansila/portfolio/backend/internal/routes"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// version is overridden at build time:
//
//	go build -ldflags "-X main.version=$(git rev-parse --short HEAD)"
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := newLogger(cfg)
	log.Info("starting",
		slog.String("version", version),
		slog.String("env", cfg.Env),
		slog.String("addr", cfg.Addr()),
		slog.Bool("db_enabled", cfg.DB.Enabled),
	)

	// Signal context: SIGTERM is what a container runtime sends on stop, so the
	// server must drain on it rather than be killed mid-request.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		db              *sql.DB
		messages        *repositories.ContactRepository
		contentRepo     *repositories.ContentRepository
		translationRepo *repositories.TranslationRepository
		authService     *services.AuthService
		audit           *services.AuditService
		sessions        *repositories.SessionRepository
	)

	if cfg.DB.Enabled {
		db, err = database.Connect(ctx, cfg.DB)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.Close(); cerr != nil {
				log.Error("closing database", slog.String("error", cerr.Error()))
			}
		}()

		messages = repositories.NewContactRepository(db)
		contentRepo = repositories.NewContentRepository(db)
		translationRepo = repositories.NewTranslationRepository(db)
		sessions = repositories.NewSessionRepository(db)
		authService = services.NewAuthService(repositories.NewUserRepository(db), sessions)
		audit = services.NewAuditService(repositories.NewAuditRepository(db), log)

		log.Info("database connected", slog.String("name", cfg.DB.Name))
	} else {
		// Without a database there is no content and no admin; the site falls
		// back to its bundled defaults and the contact form rejects submissions.
		log.Warn("database disabled — content, admin and contact submissions are unavailable")
	}

	limiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	defer limiter.Close()

	loginLimiter := middleware.NewRateLimiter(cfg.LoginRateLimit, cfg.LoginRateWindow)
	defer loginLimiter.Close()

	if sessions != nil {
		stopPurge := startSessionPurge(ctx, sessions, log)
		defer stopPurge()
	}

	var contactRepo services.ContactRepository
	if messages != nil {
		contactRepo = messages
	}

	handler := routes.New(routes.Dependencies{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		ContactService: services.NewContactService(contactRepo),
		ContentService: services.NewContentService(contentRepo, translationRepo),
		AuthService:    authService,
		AuditService:   audit,
		Messages:       messages,
		Version:        version,
		RateLimiter:    limiter,
		LoginLimiter:   loginLimiter,
	})

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 16,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", slog.String("addr", cfg.Addr()))
		if lerr := server.ListenAndServe(); lerr != nil && !errors.Is(lerr, http.ErrServerClosed) {
			errCh <- lerr
		}
	}()

	select {
	case lerr := <-errCh:
		return lerr
	case <-ctx.Done():
		log.Info("shutdown signal received", slog.Duration("grace", cfg.ShutdownTimeout))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Info("stopped cleanly")
	return nil
}

// startSessionPurge deletes expired sessions periodically.
//
// Sessions are server-side, so nothing removes an expired row on its own and the
// table would grow forever.
func startSessionPurge(ctx context.Context, sessions *repositories.SessionRepository, log *slog.Logger) func() {
	ticker := time.NewTicker(time.Hour)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				removed, err := sessions.PurgeExpired(ctx)
				if err != nil {
					log.Error("purging sessions", slog.String("error", err.Error()))
					continue
				}
				if removed > 0 {
					log.Info("purged expired sessions", slog.Int64("count", removed))
				}
			}
		}
	}()

	return func() { close(done) }
}

// newLogger emits JSON in production so log aggregation can parse it, and text
// locally because a human is reading it.
func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	if !cfg.IsProduction() {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(slog.String("service", "portfolio-api"))
}
