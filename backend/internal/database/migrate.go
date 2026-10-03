package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/migrations"
)

const (
	// lockName serialises migrations across instances: during a deploy the new
	// container can start while the old one is still up, and two runners must
	// not both apply the same file.
	lockName    = "portfolio_migrate"
	lockTimeout = 60 // seconds

	createTrackingTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version    VARCHAR(191) NOT NULL,
    applied_at DATETIME     NOT NULL,
    PRIMARY KEY (version)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci`
)

// legacy lists the migrations written before schema_migrations existed, each
// with a query that counts what its last statement left behind.
//
// A database set up by hand — with scripts/migrate.sh or the Compose init
// directory — has no record of what ran. Re-running those files is not safe:
// 0002 and 0005 add columns and fail the second time, and the seed in 0003
// upserts, so it would overwrite anything edited in the admin panel since. On
// the first start these probes decide what is already there instead.
//
// They ran in order, so the first probe that finds nothing ends the check and
// everything from there on is applied normally. Migrations after 0007 need no
// entry: from now on every file is recorded when it runs.
var legacy = []struct {
	version string
	probe   string
}{
	{"0001_create_contact_messages", tableExists("contact_messages")},
	{"0002_cms_schema", columnExists("contact_messages", "read_at")},
	// Reached only once 0002 is confirmed, so the table exists.
	{"0003_seed_content", "SELECT COUNT(*) FROM site_settings"},
	{"0004_translations", tableExists("content_translations")},
	{"0005_contact_and_experience", columnExists("experience_entries", "technologies")},
	{"0006_principles", tableExists("principles")},
	{"0007_pillars", tableExists("pillars")},
}

func tableExists(table string) string {
	return "SELECT COUNT(*) FROM information_schema.tables" +
		" WHERE table_schema = DATABASE() AND table_name = '" + table + "'"
}

func columnExists(table, column string) string {
	return "SELECT COUNT(*) FROM information_schema.columns" +
		" WHERE table_schema = DATABASE() AND table_name = '" + table + "'" +
		" AND column_name = '" + column + "'"
}

// Migrate applies every embedded migration the database has not recorded yet.
//
// It runs at startup, so a new database — a fresh Railway MySQL, say — holds
// the schema and the seed content after the first deploy, with nobody running
// a script against it. Each file runs once and is recorded in
// schema_migrations.
//
// It uses its own connection, with multiStatements on: a migration file holds
// many statements, and the seed carries session variables (@pid) from one to
// the next. The application pool never gets that flag, because it widens what
// a SQL injection could do.
func Migrate(ctx context.Context, cfg config.DBConfig, log *slog.Logger) error {
	mc, err := mysql.ParseDSN(cfg.DSN())
	if err != nil {
		return fmt.Errorf("database: migrate: %w", err)
	}
	mc.MultiStatements = true
	// The pool's 5s read timeout suits a request, not an ALTER on a large table.
	mc.ReadTimeout = 5 * time.Minute
	mc.WriteTimeout = 5 * time.Minute

	connector, err := mysql.NewConnector(mc)
	if err != nil {
		return fmt.Errorf("database: migrate: %w", err)
	}
	db := sql.OpenDB(connector)
	defer func() { _ = db.Close() }()

	// One connection throughout: the advisory lock and the seed's session
	// variables both belong to a session.
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("database: migrate: connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	var locked sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", lockName, lockTimeout).Scan(&locked); err != nil {
		return fmt.Errorf("database: migrate: lock: %w", err)
	}
	if !locked.Valid || locked.Int64 != 1 {
		return errors.New("database: migrate: another instance held the migration lock for 60s")
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), "DO RELEASE_LOCK(?)", lockName) }()

	if _, err := conn.ExecContext(ctx, createTrackingTable); err != nil {
		return fmt.Errorf("database: migrate: tracking table: %w", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		found, err := baseline(ctx, conn)
		if err != nil {
			return err
		}
		for _, version := range found {
			if err := record(ctx, conn, version); err != nil {
				return err
			}
			applied[version] = true
		}
		if len(found) > 0 {
			log.Info("migrations: recorded what was already applied by hand",
				slog.String("through", found[len(found)-1]))
		}
	}

	versions, err := Versions()
	if err != nil {
		return err
	}

	var ran int
	for _, version := range versions {
		if applied[version] {
			continue
		}

		body, err := fs.ReadFile(migrations.FS, version+".up.sql")
		if err != nil {
			return fmt.Errorf("database: migrate: %w", err)
		}

		start := time.Now()
		if _, err := conn.ExecContext(ctx, string(body)); err != nil {
			// MySQL cannot roll DDL back, so a file that fails halfway leaves what
			// it already did. Refusing to start says so loudly, where serving on a
			// half-built schema would not.
			return fmt.Errorf("database: migration %s failed: %w", version, err)
		}
		if err := record(ctx, conn, version); err != nil {
			return err
		}

		ran++
		log.Info("migration applied",
			slog.String("version", version),
			slog.Duration("took", time.Since(start).Round(time.Millisecond)))
	}

	if ran == 0 {
		log.Info("migrations: schema is up to date", slog.String("version", versions[len(versions)-1]))
	}
	return nil
}

// Versions lists the embedded migrations in the order they apply, by name
// without the .up.sql suffix.
func Versions() ([]string, error) {
	names, err := fs.Glob(migrations.FS, "*.up.sql")
	if err != nil {
		return nil, fmt.Errorf("database: migrate: %w", err)
	}
	if len(names) == 0 {
		return nil, errors.New("database: migrate: no migrations embedded")
	}

	// The numeric prefix is zero-padded, so name order is apply order.
	sort.Strings(names)
	for i, name := range names {
		names[i] = strings.TrimSuffix(name, ".up.sql")
	}
	return names, nil
}

// LegacyVersions lists the migrations that predate schema_migrations, for the
// test that keeps the probes in step with the files.
func LegacyVersions() []string {
	out := make([]string, len(legacy))
	for i, l := range legacy {
		out[i] = l.version
	}
	return out
}

func appliedVersions(ctx context.Context, conn *sql.Conn) (map[string]bool, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("database: migrate: reading schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("database: migrate: reading schema_migrations: %w", err)
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func baseline(ctx context.Context, conn *sql.Conn) ([]string, error) {
	var found []string
	for _, l := range legacy {
		var count int
		if err := conn.QueryRowContext(ctx, l.probe).Scan(&count); err != nil {
			return nil, fmt.Errorf("database: migrate: checking %s: %w", l.version, err)
		}
		if count == 0 {
			break
		}
		found = append(found, l.version)
	}
	return found, nil
}

func record(ctx context.Context, conn *sql.Conn, version string) error {
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, UTC_TIMESTAMP())", version); err != nil {
		return fmt.Errorf("database: migrate: recording %s: %w", version, err)
	}
	return nil
}
