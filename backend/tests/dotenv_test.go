package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/config"
)

func writeEnv(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

func TestLoadDotEnvSetsVariables(t *testing.T) {
	path := writeEnv(t, "DOTENV_A=hello\nDOTENV_B=world\n")
	t.Setenv("DOTENV_A", "")
	os.Unsetenv("DOTENV_A")
	t.Setenv("DOTENV_B", "")
	os.Unsetenv("DOTENV_B")

	if err := config.LoadDotEnv(path); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := os.Getenv("DOTENV_A"); got != "hello" {
		t.Errorf("DOTENV_A = %q, want hello", got)
	}
	if got := os.Getenv("DOTENV_B"); got != "world" {
		t.Errorf("DOTENV_B = %q, want world", got)
	}
}

func TestLoadDotEnvDoesNotOverrideRealEnvironment(t *testing.T) {
	path := writeEnv(t, "DOTENV_PRESET=from-file\n")
	t.Setenv("DOTENV_PRESET", "from-environment")

	if err := config.LoadDotEnv(path); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Compose and systemd set real variables; a stale file must not win.
	if got := os.Getenv("DOTENV_PRESET"); got != "from-environment" {
		t.Errorf("DOTENV_PRESET = %q, want from-environment", got)
	}
}

func TestLoadDotEnvIgnoresMissingFile(t *testing.T) {
	if err := config.LoadDotEnv(filepath.Join(t.TempDir(), "absent")); err != nil {
		t.Fatalf("a missing .env must not be an error, got %v", err)
	}
}

func TestLoadDotEnvParsesRealWorldLines(t *testing.T) {
	path := writeEnv(t, `
# a comment
   # an indented comment

export DOTENV_EXPORTED=yes
DOTENV_SPACED = padded
DOTENV_DQUOTE="a value # not a comment"
DOTENV_SQUOTE='literal $DOLLAR \not escaped'
DOTENV_INLINE=value # trailing comment
DOTENV_HASH_IN_PASSWORD=aB3#xY9$zQ
DOTENV_EMPTY=
DOTENV_DURATION=1h
DOTENV_ESCAPES="line\nbreak"
`)

	for _, key := range []string{
		"DOTENV_EXPORTED", "DOTENV_SPACED", "DOTENV_DQUOTE", "DOTENV_SQUOTE",
		"DOTENV_INLINE", "DOTENV_HASH_IN_PASSWORD", "DOTENV_EMPTY", "DOTENV_DURATION",
		"DOTENV_ESCAPES",
	} {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })
	}

	if err := config.LoadDotEnv(path); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cases := map[string]string{
		"DOTENV_EXPORTED":         "yes",
		"DOTENV_SPACED":           "padded",
		"DOTENV_DQUOTE":           "a value # not a comment",
		"DOTENV_SQUOTE":           `literal $DOLLAR \not escaped`,
		"DOTENV_INLINE":           "value",
		"DOTENV_HASH_IN_PASSWORD": "aB3#xY9$zQ",
		"DOTENV_EMPTY":            "",
		"DOTENV_DURATION":         "1h",
		"DOTENV_ESCAPES":          "line\nbreak",
	}

	for key, want := range cases {
		if got := os.Getenv(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestLoadDotEnvRejectsMalformedLines(t *testing.T) {
	cases := map[string]string{
		"no equals sign":     "JUST_A_WORD\n",
		"invalid name":       "9INVALID=value\n",
		"name with a hyphen": "BAD-NAME=value\n",
		"unterminated quote": "BROKEN=\"unclosed\n",
	}

	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			if err := config.LoadDotEnv(writeEnv(t, contents)); err == nil {
				t.Fatal("expected a parse error")
			}
		})
	}
}

func TestLoadRespectsEnvFileOverride(t *testing.T) {
	path := writeEnv(t, "DOTENV_VIA_ENVFILE=found\n")
	os.Unsetenv("DOTENV_VIA_ENVFILE")
	t.Cleanup(func() { os.Unsetenv("DOTENV_VIA_ENVFILE") })
	t.Setenv("ENV_FILE", path)

	if err := config.LoadDotEnv(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := os.Getenv("DOTENV_VIA_ENVFILE"); got != "found" {
		t.Errorf("DOTENV_VIA_ENVFILE = %q, want found", got)
	}
}

func TestLoadReadsCredentialsFromDotEnv(t *testing.T) {
	// The failure this fixes: a filled-in .env while Load still reported
	// "DB_USER is required".
	path := writeEnv(t, "DB_ENABLED=true\nDB_USER=portfolio_app\nDB_PASSWORD=secret\n")

	for _, key := range []string{"DB_ENABLED", "DB_USER", "DB_PASSWORD"} {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })
	}
	t.Setenv("ENV_FILE", path)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected Load to succeed using .env, got %v", err)
	}
	if cfg.DB.User != "portfolio_app" {
		t.Errorf("DB.User = %q, want portfolio_app", cfg.DB.User)
	}
}
