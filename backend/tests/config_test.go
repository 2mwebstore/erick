package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/config"
)

func TestLoadRequiresDatabaseCredentialsWhenEnabled(t *testing.T) {
	t.Setenv("DB_ENABLED", "true")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected Load to refuse to start without database credentials")
	}
}

func TestLoadSucceedsWithoutDatabaseWhenDisabled(t *testing.T) {
	t.Setenv("DB_ENABLED", "false")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.DB.Enabled {
		t.Error("expected DB.Enabled to be false")
	}
}

func TestLoadParsesOverrides(t *testing.T) {
	t.Setenv("DB_ENABLED", "false")
	t.Setenv("PORT", "9090")
	t.Setenv("APP_ENV", "production")
	t.Setenv("RATE_LIMIT_REQUESTS", "3")
	t.Setenv("RATE_LIMIT_WINDOW", "30m")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction to be true")
	}
	if cfg.RateLimitRequests != 3 {
		t.Errorf("expected 3 requests, got %d", cfg.RateLimitRequests)
	}
	if cfg.RateLimitWindow != 30*time.Minute {
		t.Errorf("expected a 30m window, got %s", cfg.RateLimitWindow)
	}
}

func TestLoadRejectsZeroRateLimit(t *testing.T) {
	t.Setenv("DB_ENABLED", "false")
	t.Setenv("RATE_LIMIT_REQUESTS", "0")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected Load to reject a rate limit of 0")
	}
}

func TestLoadMigratesOnStartByDefault(t *testing.T) {
	t.Setenv("DB_ENABLED", "false")
	t.Setenv("DB_AUTO_MIGRATE", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cfg.DB.AutoMigrate {
		t.Error("a fresh database must get its schema without anyone opting in")
	}
	if cfg.SeedAdmin.Enabled() {
		t.Error("no admin account should be seeded unless one is configured")
	}
}

func TestLoadRejectsHalfASeedAdmin(t *testing.T) {
	for name, vars := range map[string][2]string{
		"email only":    {"you@example.com", ""},
		"password only": {"", "a long enough passphrase"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("DB_ENABLED", "false")
			t.Setenv("SEED_ADMIN_EMAIL", vars[0])
			t.Setenv("SEED_ADMIN_PASSWORD", vars[1])

			if _, err := config.Load(); err == nil {
				t.Fatal("expected Load to refuse a seed admin missing half its settings")
			}
		})
	}
}

func TestDSNContainsNoLoggableSecretByAccident(t *testing.T) {
	// The DSN necessarily contains the password; this test documents that fact so
	// nobody logs cfg.DB.DSN() casually. It must never be passed to a logger.
	dsn := config.DBConfig{
		Host: "db", Port: "3306", Name: "portfolio",
		User: "app", Password: "s3cret",
	}.DSN()

	if !strings.Contains(dsn, "s3cret") {
		t.Fatal("DSN unexpectedly omits the password — update the logging guidance")
	}
	if !strings.Contains(dsn, "parseTime=true") {
		t.Error("DSN must set parseTime=true so DATETIME scans into time.Time")
	}
}
