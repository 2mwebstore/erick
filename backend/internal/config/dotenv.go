package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// LoadDotEnv reads a .env file into the process environment.
//
// Go has no built-in support for this, and the alternative — a dependency, or
// remembering `set -a; source .env` before every run — is worse than forty
// lines of parsing.
//
// Two rules make it safe to call unconditionally:
//
//   - A variable already present in the real environment is never overwritten.
//     Docker Compose, systemd and `DB_USER=x go run ...` therefore always win
//     over a stale file left in the working directory.
//   - A missing file is not an error. Containers have no .env, and that is the
//     expected case in production.
//
// Set ENV_FILE to load from somewhere other than ./.env.
func LoadDotEnv(paths ...string) error {
	if len(paths) == 0 {
		if custom := strings.TrimSpace(os.Getenv("ENV_FILE")); custom != "" {
			paths = []string{custom}
		} else {
			paths = []string{".env"}
		}
	}

	for _, path := range paths {
		if err := loadFile(path); err != nil {
			return err
		}
	}
	return nil
}

func loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config: open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	line := 0

	for scanner.Scan() {
		line++

		key, value, ok, err := parseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("config: %s line %d: %w", path, line, err)
		}
		if !ok {
			continue
		}

		// The real environment is authoritative.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("config: set %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	return nil
}

// parseLine returns the key and value for an assignment, or ok=false for a
// blank line or comment.
func parseLine(raw string) (key, value string, ok bool, err error) {
	line := strings.TrimSpace(raw)

	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}

	line = strings.TrimPrefix(line, "export ")

	name, rest, found := strings.Cut(line, "=")
	if !found {
		return "", "", false, fmt.Errorf("expected KEY=VALUE, got %q", raw)
	}

	key = strings.TrimSpace(name)
	if !validKey(key) {
		return "", "", false, fmt.Errorf("invalid variable name %q", key)
	}

	value, err = parseValue(strings.TrimSpace(rest))
	if err != nil {
		return "", "", false, err
	}

	return key, value, true, nil
}

func parseValue(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	switch raw[0] {
	case '\'':
		// Single quotes are literal, so a password full of $ and \ survives intact.
		closing := strings.LastIndexByte(raw, '\'')
		if closing == 0 {
			return "", errors.New("unterminated single quote")
		}
		return raw[1:closing], nil

	case '"':
		closing := strings.LastIndexByte(raw, '"')
		if closing == 0 {
			return "", errors.New("unterminated double quote")
		}
		return unescape(raw[1:closing]), nil

	default:
		// Only strip a comment that follows whitespace: a value may legitimately
		// contain '#', as generated passwords often do.
		if idx := strings.Index(raw, " #"); idx >= 0 {
			raw = raw[:idx]
		}
		return strings.TrimSpace(raw), nil
	}
}

func unescape(s string) string {
	replacer := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, `"`, `\\`, `\`)
	return replacer.Replace(s)
}

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}
