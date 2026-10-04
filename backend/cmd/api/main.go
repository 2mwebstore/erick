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
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
	"github.com/kongchansila/portfolio/backend/internal/routes"
	"github.com/kongchansila/portfolio/backend/internal/seoaudit"
	"github.com/kongchansila/portfolio/backend/internal/services"
	"github.com/kongchansila/portfolio/backend/internal/uploads"
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

		// Before anything reads a table: a fresh database gets its schema and
		// seed content here, and an existing one gets whatever is new.
		if cfg.DB.AutoMigrate {
			if err := database.Migrate(ctx, cfg.DB, log); err != nil {
				return err
			}
		} else {
			log.Warn("DB_AUTO_MIGRATE is false — pending migrations were not applied")
		}

		messages = repositories.NewContactRepository(db)
		contentRepo = repositories.NewContentRepository(db)
		translationRepo = repositories.NewTranslationRepository(db)
		sessions = repositories.NewSessionRepository(db)
		authService = services.NewAuthService(repositories.NewUserRepository(db), sessions)
		audit = services.NewAuditService(repositories.NewAuditRepository(db), log)

		log.Info("database connected", slog.String("name", cfg.DB.Name))

		seedAdmin(ctx, cfg.SeedAdmin, authService, log)
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

	// The SEO audit crawls the public site; without its address it stays off.
	var auditor *seoaudit.Auditor
	if cfg.SiteURL != "" {
		if auditor, err = seoaudit.New(cfg.SiteURL, nil); err != nil {
			return err
		}
	} else {
		log.Warn("SITE_URL is not set — the SEO audit in the admin panel is off")
	}

	// Image uploads go to R2. Without it, image fields still take a pasted link.
	var uploader *uploads.Uploader
	switch {
	case cfg.R2.Configured():
		store := uploads.NewR2Store(cfg.R2.EndpointURL(), cfg.R2.AccessKeyID, cfg.R2.SecretAccessKey, cfg.R2.Bucket, nil)
		uploader = uploads.NewUploader(store, cfg.R2.PublicURL, cfg.R2.MaxUploadBytes)
		log.Info("image uploads on", slog.String("bucket", cfg.R2.Bucket), slog.String("public_url", cfg.R2.PublicURL))
	case cfg.R2.Started():
		// Half set up: off, and loudly, rather than failing on the first upload.
		log.Error("R2 is partly configured — image uploads are off until these are set",
			slog.Any("missing", cfg.R2.Missing()))
	default:
		log.Info("R2 is not configured — image fields take links only")
	}

	contentService := services.NewContentService(contentRepo, translationRepo)
	// Deletes an uploaded image from the bucket once nothing shows it. Off
	// (nil) when uploads are.
	imageCleanup := services.NewImageCleanup(uploader, contentService, audit, log)

	handler := routes.New(routes.Dependencies{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		ContactService: services.NewContactService(contactRepo),
		ContentService: contentService,
		AuthService:    authService,
		AuditService:   audit,
		SEOAudit:       services.NewSEOAuditService(auditor, log),
		Uploader:       uploader,
		ImageCleanup:   imageCleanup,
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
	// Let image deletions started by the last requests finish (each is bounded
	// to 30 seconds), rather than cutting one off halfway.
	imageCleanup.Wait()

	log.Info("stopped cleanly")
	return nil
}

// seedAdmin creates the first admin account from SEED_ADMIN_*, and only while
// there are no accounts at all — so a restart can never reset a password that
// was changed in the panel since.
//
// A failure is logged rather than fatal: the public site does not need an
// account, and should not go down over a seed password that is too short.
func seedAdmin(ctx context.Context, seed config.SeedAdminConfig, auth *services.AuthService, log *slog.Logger) {
	if !seed.Enabled() {
		return
	}

	count, err := auth.UserCount(ctx)
	if err != nil {
		log.Error("seeding admin: counting accounts", slog.String("error", err.Error()))
		return
	}
	if count > 0 {
		log.Warn("SEED_ADMIN_PASSWORD is still set but accounts already exist, so it is unused — remove it")
		return
	}

	user, err := auth.CreateUser(ctx, seed.Email, seed.Name, seed.Password, models.RoleAdmin)
	if err != nil {
		var invalid *services.ValidationError
		if errors.As(err, &invalid) {
			log.Error("seeding admin: account not created", slog.Any("fields", invalid.Fields))
			return
		}
		log.Error("seeding admin: account not created", slog.String("error", err.Error()))
		return
	}

	log.Info("seeded admin account — sign in, change the password in the panel, then remove SEED_ADMIN_PASSWORD",
		slog.String("email", user.Email))
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
