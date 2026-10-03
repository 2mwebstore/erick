package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/handlers"
)

/**
 * The admin panel refuses to trust a server below the contract it was built
 * against. That only works if the number moves when the wire format does, so
 * this is the reminder: if a change breaks an older admin build, bump
 * handlers.APIContract and REQUIRED_API_CONTRACT in composables/useAdmin.ts
 * together.
 */
func TestAPIContractIsPinned(t *testing.T) {
	const expected = 5

	if handlers.APIContract != expected {
		t.Fatalf("APIContract = %d, want %d.\n"+
			"If this is a deliberate bump, update REQUIRED_API_CONTRACT in "+
			"frontend/composables/useAdmin.ts to match, then update this test.",
			handlers.APIContract, expected)
	}
}

// A contract of zero would make every server look current, which is exactly
// the failure this is meant to catch — older builds omit the field entirely.
func TestAPIContractIsPositive(t *testing.T) {
	if handlers.APIContract < 1 {
		t.Fatal("APIContract must be positive, or an older server reads as current")
	}
}

// The frontend constant has to agree, and nothing else checks that.
func TestFrontendRequiresTheSameContract(t *testing.T) {
	source := readRepoFile(t, "frontend/composables/useAdmin.ts")

	want := "export const REQUIRED_API_CONTRACT = 5"
	if !strings.Contains(source, want) {
		t.Errorf("useAdmin.ts should declare %q to match handlers.APIContract = %d",
			want, handlers.APIContract)
	}
}

// readRepoFile reads a path relative to the repository root.
//
// The two constants live in different languages in different directories, so
// the only way to check they agree is to read the other one as text.
func readRepoFile(t *testing.T, relative string) string {
	t.Helper()

	// tests/ sits at backend/tests, so the repo root is two levels up.
	path := filepath.Join("..", "..", relative)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", relative, err)
	}
	return string(data)
}

/**
 * Every allowed setting has to exist in the seed.
 *
 * A key added to AllowedSettings but never seeded simply never appears in the
 * admin panel, which looks like the field is broken rather than absent. This
 * caught two: the hero headline pair, added by hand to a generated file and
 * silently dropped the next time it regenerated.
 */
func TestEverySettingIsSeeded(t *testing.T) {
	source := readRepoFile(t, "backend/internal/services/content.go")
	block := regexp.MustCompile(`(?s)var AllowedSettings = map\[string\]bool\{(.*?)\n\}`).FindStringSubmatch(source)
	if block == nil {
		t.Fatal("could not find AllowedSettings")
	}

	allowed := regexp.MustCompile(`"([a-z_]+)":`).FindAllStringSubmatch(block[1], -1)
	if len(allowed) == 0 {
		t.Fatal("AllowedSettings looks empty — the pattern probably drifted")
	}

	seed := readRepoFile(t, "backend/migrations/0003_seed_content.up.sql")
	for _, match := range allowed {
		key := match[1]
		needle := "VALUES ('" + key + "'"
		if !strings.Contains(seed, needle) {
			t.Errorf("setting %q is allowed but never seeded — add it to "+
				"frontend/scripts/export-seed.mjs and regenerate", key)
		}
	}
}
