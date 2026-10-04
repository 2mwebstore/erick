// Package config loads runtime configuration from the environment.
//
// Nothing is read from a committed file and nothing has a production-safe
// hard-coded default: the process refuses to start if a required secret is
// missing, which is preferable to starting with a guessable one (§31, §34).
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	Host            string
	Port            string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration

	DB DBConfig

	// RateLimit is applied per client IP to POST /v1/contact.
	RateLimitRequests int
	RateLimitWindow   time.Duration

	MaxBodyBytes int64
	// Admin payloads carry full case-study text, so they need a larger cap than
	// a contact message.
	MaxAdminBodyBytes int64

	// SecureCookies marks session cookies Secure. Must be true in production;
	// false only for plain-HTTP local development, where the browser would
	// otherwise refuse to store them.
	SecureCookies bool

	// LoginRateLimit is deliberately separate from and tighter than the contact
	// form limit — this is the endpoint where guessing is the attack.
	LoginRateLimit  int
	LoginRateWindow time.Duration

	// TrustedProxy enables reading the client IP from X-Forwarded-For.
	// Only enable it when the service genuinely sits behind a proxy you control,
	// otherwise the header is attacker-controlled and defeats rate limiting.
	TrustedProxy bool

	// SeedAdmin is the first admin account, created at startup only while the
	// database has no accounts at all.
	SeedAdmin SeedAdminConfig

	// SiteURL is the public site's origin, the one address the SEO audit
	// crawls. Empty turns the audit off.
	SiteURL string

	// R2 is where uploaded images are stored. Unset, image fields take links only.
	R2 R2Config
}

// R2Config is a Cloudflare R2 bucket, read from the same variables the other
// projects on this account use.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	// Endpoint is the S3 API address; derived from AccountID when unset.
	Endpoint string
	// PublicURL is the bucket's public address — an r2.dev URL or a custom
	// domain — that uploaded images are served from.
	PublicURL string
	// MaxUploadBytes caps one image.
	MaxUploadBytes int64
}

// EndpointURL is the S3 API address to use.
func (r R2Config) EndpointURL() string {
	if r.Endpoint != "" {
		return r.Endpoint
	}
	if r.AccountID != "" {
		return "https://" + r.AccountID + ".r2.cloudflarestorage.com"
	}
	return ""
}

// Missing names the variables still needed before uploads can be turned on.
func (r R2Config) Missing() []string {
	var missing []string
	if r.EndpointURL() == "" {
		missing = append(missing, "R2_ACCOUNT_ID (or R2_ENDPOINT)")
	}
	if r.AccessKeyID == "" {
		missing = append(missing, "R2_ACCESS_KEY_ID")
	}
	if r.SecretAccessKey == "" {
		missing = append(missing, "R2_SECRET_ACCESS_KEY")
	}
	if r.Bucket == "" {
		missing = append(missing, "R2_BUCKET")
	}
	if r.PublicURL == "" {
		missing = append(missing, "R2_PUBLIC_URL")
	}
	return missing
}

// Configured reports whether every variable is set.
func (r R2Config) Configured() bool { return len(r.Missing()) == 0 }

// Started reports whether any variable is set, so a half-finished setup can be
// told apart from no setup at all.
func (r R2Config) Started() bool {
	return r.AccountID != "" || r.AccessKeyID != "" || r.SecretAccessKey != "" ||
		r.Bucket != "" || r.Endpoint != "" || r.PublicURL != ""
}

// SeedAdminConfig creates the first account without a shell on the server,
// which a host like Railway does not readily give. There is deliberately no
// default password: a guessable one would be live on every fresh deploy.
type SeedAdminConfig struct {
	Email    string
	Name     string
	Password string
}

// Enabled reports whether an account should be seeded.
func (s SeedAdminConfig) Enabled() bool { return s.Email != "" }

type DBConfig struct {
	Host         string
	Port         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int
	MaxIdleConns int
	ConnLifetime time.Duration
	// Enabled is false when the service runs without persistence.
	Enabled bool
	// AutoMigrate applies pending migrations at startup.
	AutoMigrate bool
}

func (c Config) Addr() string { return c.Host + ":" + c.Port }

func (c Config) IsProduction() bool { return c.Env == "production" }

// DSN returns a go-sql-driver compatible DSN. Kept out of logs deliberately.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=5s&readTimeout=5s&writeTimeout=5s",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// Load reads configuration from the environment and validates it.
//
// A .env file in the working directory is loaded first, for local development.
// It never overrides a variable that is already set, so containers and systemd
// are unaffected by one being present.
func Load() (Config, error) {
	if err := LoadDotEnv(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Env:               env("APP_ENV", "development"),
		Host:              env("HOST", "0.0.0.0"),
		Port:              env("PORT", "8080"),
		ShutdownTimeout:   duration("SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:       duration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      duration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:       duration("IDLE_TIMEOUT", 60*time.Second),
		RateLimitRequests: integer("RATE_LIMIT_REQUESTS", 5),
		RateLimitWindow:   duration("RATE_LIMIT_WINDOW", time.Hour),
		MaxBodyBytes:      int64(integer("MAX_BODY_BYTES", 16*1024)),
		MaxAdminBodyBytes: int64(integer("MAX_ADMIN_BODY_BYTES", 256*1024)),
		SecureCookies:     boolean("SECURE_COOKIES", env("APP_ENV", "development") == "production"),
		LoginRateLimit:    integer("LOGIN_RATE_LIMIT", 10),
		LoginRateWindow:   duration("LOGIN_RATE_WINDOW", 15*time.Minute),
		TrustedProxy:      boolean("TRUSTED_PROXY", true),
		SeedAdmin: SeedAdminConfig{
			Email:    env("SEED_ADMIN_EMAIL", ""),
			Name:     env("SEED_ADMIN_NAME", "Admin"),
			Password: env("SEED_ADMIN_PASSWORD", ""),
		},
		// The frontend's variable is accepted too: on a host where both services
		// share variables, the site address is then set in one place only.
		SiteURL: env("SITE_URL", env("NUXT_PUBLIC_SITE_URL", "")),
		R2: R2Config{
			AccountID:       env("R2_ACCOUNT_ID", ""),
			AccessKeyID:     env("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: env("R2_SECRET_ACCESS_KEY", ""),
			Bucket:          env("R2_BUCKET", ""),
			Endpoint:        env("R2_ENDPOINT", ""),
			PublicURL:       strings.TrimRight(env("R2_PUBLIC_URL", ""), "/"),
			MaxUploadBytes:  int64(integer("UPLOAD_MAX_BYTES", 5<<20)),
		},
		DB: DBConfig{
			Host:         env("DB_HOST", "127.0.0.1"),
			Port:         env("DB_PORT", "3306"),
			Name:         env("DB_NAME", "portfolio"),
			User:         env("DB_USER", ""),
			Password:     env("DB_PASSWORD", ""),
			MaxOpenConns: integer("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns: integer("DB_MAX_IDLE_CONNS", 5),
			ConnLifetime: duration("DB_CONN_LIFETIME", 30*time.Minute),
			Enabled:      boolean("DB_ENABLED", true),
			AutoMigrate:  boolean("DB_AUTO_MIGRATE", true),
		},
	}

	if cfg.DB.Enabled {
		if cfg.DB.User == "" {
			return cfg, fmt.Errorf("config: DB_USER is required when DB_ENABLED is true")
		}
		if cfg.DB.Password == "" {
			return cfg, fmt.Errorf("config: DB_PASSWORD is required when DB_ENABLED is true")
		}
	}

	// Half a seed is a mistake, not a choice: an address without a password
	// would otherwise be skipped silently and leave the panel with no account.
	if (cfg.SeedAdmin.Email == "") != (cfg.SeedAdmin.Password == "") {
		return cfg, fmt.Errorf("config: SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set together")
	}

	if cfg.SiteURL != "" {
		u, err := url.Parse(cfg.SiteURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return cfg, fmt.Errorf("config: SITE_URL must be an absolute http(s) URL, such as https://example.com")
		}
	}

	// A missing R2 variable only turns uploads off (main logs which one). A
	// malformed address is a mistake worth stopping for, as with SITE_URL.
	for name, value := range map[string]string{"R2_PUBLIC_URL": cfg.R2.PublicURL, "R2_ENDPOINT": cfg.R2.Endpoint} {
		if value == "" {
			continue
		}
		if u, err := url.Parse(value); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return cfg, fmt.Errorf("config: %s must be an absolute http(s) URL", name)
		}
	}
	if cfg.R2.MaxUploadBytes < 1 {
		return cfg, fmt.Errorf("config: UPLOAD_MAX_BYTES must be positive")
	}

	if cfg.RateLimitRequests < 1 {
		return cfg, fmt.Errorf("config: RATE_LIMIT_REQUESTS must be at least 1")
	}

	if cfg.LoginRateLimit < 1 {
		return cfg, fmt.Errorf("config: LOGIN_RATE_LIMIT must be at least 1")
	}

	// A zero cap would reject every request that carries a body, since the
	// middleware compares Content-Length against it.
	if cfg.MaxBodyBytes < 1 {
		return cfg, fmt.Errorf("config: MAX_BODY_BYTES must be positive")
	}
	if cfg.MaxAdminBodyBytes < 1 {
		return cfg, fmt.Errorf("config: MAX_ADMIN_BODY_BYTES must be positive")
	}

	// A session cookie without Secure in production would travel in plaintext on
	// any non-TLS hop, so refuse to start rather than issue one.
	if cfg.IsProduction() && !cfg.SecureCookies {
		return cfg, fmt.Errorf("config: SECURE_COOKIES cannot be false when APP_ENV=production")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func integer(key string, fallback int) int {
	if v, err := strconv.Atoi(env(key, "")); err == nil {
		return v
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(env(key, "")); err == nil {
		return v
	}
	return fallback
}

func boolean(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(env(key, "")); err == nil {
		return v
	}
	return fallback
}
