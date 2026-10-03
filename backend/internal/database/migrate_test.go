package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/migrations"
)

// The probes are keyed by file name, so a renamed or renumbered legacy file
// would silently stop being recognised — and be re-run against a database
// that already has it.
func TestLegacyProbesMatchTheFiles(t *testing.T) {
	versions, err := Versions()
	if err != nil {
		t.Fatal(err)
	}

	legacyVersions := LegacyVersions()
	if len(legacyVersions) > len(versions) {
		t.Fatalf("%d legacy probes but only %d migrations", len(legacyVersions), len(versions))
	}
	for i, v := range legacyVersions {
		if versions[i] != v {
			t.Errorf("legacy probe %d is %q, but migration %d is %q", i, v, i, versions[i])
		}
	}
}

func TestVersionsAreOrderedAndUnique(t *testing.T) {
	versions, err := Versions()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.IsSorted(versions) {
		t.Errorf("not in apply order: %v", versions)
	}

	seen := map[string]bool{}
	for _, v := range versions {
		prefix := v[:4]
		if seen[prefix] {
			t.Errorf("two migrations share the prefix %s", prefix)
		}
		seen[prefix] = true
	}
}

// The migration tests below need a real MySQL. They create and drop their own
// databases, named portfolio_migrate_test_*, and never touch DB_NAME:
//
//	MIGRATE_TEST_MYSQL=1 DB_USER=root DB_PASSWORD=… go test ./internal/database
func testServer(t *testing.T) config.DBConfig {
	t.Helper()
	if os.Getenv("MIGRATE_TEST_MYSQL") == "" {
		t.Skip("set MIGRATE_TEST_MYSQL=1 to run against a real MySQL")
	}

	cfg := config.DBConfig{
		Host:     envOr("DB_HOST", "127.0.0.1"),
		Port:     envOr("DB_PORT", "3306"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}
	if cfg.User == "" {
		t.Fatal("DB_USER is required")
	}
	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// scratch creates an empty database and returns a config pointing at it.
func scratch(t *testing.T, server config.DBConfig, name string) config.DBConfig {
	t.Helper()

	admin := open(t, server, false)
	name = "portfolio_migrate_test_" + name
	exec(t, admin, "DROP DATABASE IF EXISTS "+name)
	exec(t, admin, "CREATE DATABASE "+name+" CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE IF EXISTS " + name) })

	cfg := server
	cfg.Name = name
	return cfg
}

func open(t *testing.T, cfg config.DBConfig, multi bool) *sql.DB {
	t.Helper()
	mc, err := mysql.ParseDSN(cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	mc.MultiStatements = multi
	connector, err := mysql.NewConnector(mc)
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func exec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

// applyByHand runs files the way scripts/migrate.sh did: no record kept.
func applyByHand(t *testing.T, cfg config.DBConfig, versions ...string) {
	t.Helper()
	db := open(t, cfg, true)
	for _, v := range versions {
		body, err := fs.ReadFile(migrations.FS, v+".up.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(body)); err != nil {
			t.Fatalf("applying %s by hand: %v", v, err)
		}
	}
}

func recorded(t *testing.T, cfg config.DBConfig) []string {
	t.Helper()
	rows, err := open(t, cfg, false).Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		out = append(out, v)
	}
	return out
}

func count(t *testing.T, cfg config.DBConfig, query string) int {
	t.Helper()
	var n int
	if err := open(t, cfg, false).QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func migrate(t *testing.T, cfg config.DBConfig) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := Migrate(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
}

func TestMigrateFreshDatabase(t *testing.T) {
	cfg := scratch(t, testServer(t), "fresh")
	versions, _ := Versions()

	migrate(t, cfg)

	if got := recorded(t, cfg); !slices.Equal(got, versions) {
		t.Fatalf("recorded %v, want %v", got, versions)
	}
	if n := count(t, cfg, "SELECT COUNT(*) FROM site_settings"); n == 0 {
		t.Error("the seed did not run: site_settings is empty")
	}
	if n := count(t, cfg, "SELECT COUNT(*) FROM projects"); n == 0 {
		t.Error("the seed did not run: projects is empty")
	}

	// The connection must be UTF-8 end to end, or the em dashes in the seed are
	// stored as "â€”" for good — see scripts/_db.sh.
	if n := count(t, cfg, "SELECT COUNT(*) FROM site_settings WHERE setting_value LIKE '%—%'"); n == 0 {
		t.Error("no em dash survived the seed")
	}
	if n := count(t, cfg, "SELECT COUNT(*) FROM site_settings WHERE setting_value LIKE '%â€%'"); n != 0 {
		t.Errorf("%d settings hold mojibake", n)
	}

	// A restart is a no-op, and does not re-run the seed over edited content.
	exec(t, open(t, cfg, false), "UPDATE site_settings SET setting_value = 'edited in the panel' WHERE setting_key = (SELECT k FROM (SELECT MIN(setting_key) AS k FROM site_settings) AS first)")
	migrate(t, cfg)
	if n := count(t, cfg, "SELECT COUNT(*) FROM site_settings WHERE setting_value = 'edited in the panel'"); n != 1 {
		t.Error("a second start overwrote content edited since the first")
	}
}

func TestMigrateRecognisesAHandMigratedDatabase(t *testing.T) {
	cfg := scratch(t, testServer(t), "by_hand")
	legacyVersions := LegacyVersions()
	applyByHand(t, cfg, legacyVersions...)

	// Re-running 0002 would fail on its duplicate column, so getting through at
	// all means nothing was re-applied.
	migrate(t, cfg)

	versions, _ := Versions()
	if got := recorded(t, cfg); !slices.Equal(got, versions) {
		t.Fatalf("recorded %v, want %v", got, versions)
	}
}

func TestMigrateFinishesAPartlyMigratedDatabase(t *testing.T) {
	cfg := scratch(t, testServer(t), "partial")
	// What the Compose init directory used to apply.
	applyByHand(t, cfg, "0001_create_contact_messages", "0002_cms_schema", "0003_seed_content")

	migrate(t, cfg)

	versions, _ := Versions()
	if got := recorded(t, cfg); !slices.Equal(got, versions) {
		t.Fatalf("recorded %v, want %v", got, versions)
	}
	for _, table := range []string{"content_translations", "principles", "pillars"} {
		if n := count(t, cfg, tableExists(table)); n != 1 {
			t.Errorf("table %s was not created", table)
		}
	}
	if n := count(t, cfg, columnExists("experience_entries", "technologies")); n != 1 {
		t.Error("0005 was not applied")
	}
}

// Migrate sends each file as one multi-statement query. That is only safe if
// a failure in a later statement still comes back as an error, rather than
// the file being recorded as applied.
func TestMultiStatementErrorsAreReported(t *testing.T) {
	cfg := scratch(t, testServer(t), "errors")
	db := open(t, cfg, true)

	_, err := db.Exec("CREATE TABLE ok_first (id INT); SELECT * FROM missing_table; CREATE TABLE never (id INT)")
	if err == nil {
		t.Fatal("an error in the second statement was swallowed")
	}
	if n := count(t, cfg, tableExists("never")); n != 0 {
		t.Error(fmt.Sprintf("statements after the failure still ran (%d)", n))
	}
}
