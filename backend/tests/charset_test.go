package tests

import (
	"regexp"
	"strings"
	"testing"
)

// TestDBScriptsForceUTF8 guards the character set the shell scripts talk to
// MySQL with.
//
// The client's own default is `auto`, resolved from the shell's locale, and on
// a machine that reports no UTF-8 locale it lands on latin1. Every migration in
// this repository is a UTF-8 file, so a latin1 connection hands MySQL three
// latin1 characters where the file had one em dash, and stores the mojibake
// permanently: "maintenance — not just writing code" was live on the site as
// "maintenance â€” not just writing code" until this flag was added. Khmer goes
// the same way, and a dump taken over such a connection is a backup that cannot
// be restored faithfully.
func TestDBScriptsForceUTF8(t *testing.T) {
	script := readRepoFile(t, "scripts/_db.sh")

	for _, client := range []string{"mysql", "mysqldump"} {
		// An actual invocation: the client's name followed by a flag. Skips the
		// `command -v mysql` probe and the sentence in the error message, which
		// mention it without running it.
		invocation := regexp.MustCompile(`(^|\s)` + client + `\s+--`)

		var checked int
		for _, line := range strings.Split(script, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || !invocation.MatchString(trimmed) {
				continue
			}
			checked++
			if !strings.Contains(trimmed, "--default-character-set=utf8mb4") {
				t.Errorf("%s is invoked without --default-character-set=utf8mb4:\n  %s", client, trimmed)
			}
		}
		if checked == 0 {
			t.Errorf("no %s invocation found in scripts/_db.sh — has it moved?", client)
		}
	}
}
