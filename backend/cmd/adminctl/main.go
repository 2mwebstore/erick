// Command adminctl manages admin accounts from the server.
//
// The first account has to be created somewhere, and a self-service signup page
// on an admin panel is a liability. This runs beside the API with the same
// environment, so it needs no separate credentials:
//
//	go run ./cmd/adminctl create -email you@example.com -name "Your Name" -role admin
//	go run ./cmd/adminctl list
//	go run ./cmd/adminctl passwd -email you@example.com
//
// In Docker:
//
//	docker compose exec api adminctl create -email you@example.com -name "Your Name"
//
// The password is read from the terminal, never passed as a flag — a flag would
// land in the shell history and the process list.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/kongchansila/portfolio/backend/internal/config"
	"github.com/kongchansila/portfolio/backend/internal/database"
	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return errors.New("no command given")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.DB.Enabled {
		return errors.New("DB_ENABLED is false — adminctl needs the database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	users := repositories.NewUserRepository(db)
	auth := services.NewAuthService(users, repositories.NewSessionRepository(db))

	switch os.Args[1] {
	case "create":
		return create(ctx, auth)
	case "list":
		return list(ctx, auth)
	case "passwd":
		return passwd(ctx, users)
	default:
		usage()
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `adminctl — manage admin accounts

  create  -email <address> -name <name> [-role admin|editor]
  list
  passwd  -email <address>

The password is prompted for, never taken as a flag.
`)
}

func create(ctx context.Context, auth *services.AuthService) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	email := fs.String("email", "", "email address")
	name := fs.String("name", "", "display name")
	role := fs.String("role", "admin", "admin or editor")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *email == "" || *name == "" {
		return errors.New("both -email and -name are required")
	}

	password, err := readPassword(fmt.Sprintf("Password for %s: ", *email))
	if err != nil {
		return err
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}
	if password != confirm {
		return errors.New("passwords do not match")
	}

	user, err := auth.CreateUser(ctx, *email, *name, password, models.Role(*role))
	if err != nil {
		var invalid *services.ValidationError
		if errors.As(err, &invalid) {
			for field, message := range invalid.Fields {
				fmt.Fprintf(os.Stderr, "  %s: %s\n", field, message)
			}
			return errors.New("the account was not created")
		}
		return err
	}

	fmt.Printf("created %s (%s), role %s\n", user.Email, user.Name, user.Role)
	return nil
}

func list(ctx context.Context, auth *services.AuthService) error {
	users, err := auth.ListUsers(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		fmt.Println("no accounts yet — create one with: adminctl create -email … -name …")
		return nil
	}

	fmt.Printf("%-34s %-22s %-8s %-7s %s\n", "EMAIL", "NAME", "ROLE", "ACTIVE", "LAST LOGIN")
	for _, u := range users {
		lastLogin := "never"
		if u.LastLoginAt != nil {
			lastLogin = u.LastLoginAt.Format(time.RFC3339)
		}
		active := "yes"
		if !u.IsActive {
			active = "no"
		}
		fmt.Printf("%-34s %-22s %-8s %-7s %s\n", u.Email, u.Name, u.Role, active, lastLogin)
	}
	return nil
}

func passwd(ctx context.Context, users *repositories.UserRepository) error {
	fs := flag.NewFlagSet("passwd", flag.ExitOnError)
	email := fs.String("email", "", "email address")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if *email == "" {
		return errors.New("-email is required")
	}

	user, err := users.ByEmail(ctx, *email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return fmt.Errorf("no account for %s", *email)
		}
		return err
	}

	password, err := readPassword(fmt.Sprintf("New password for %s: ", user.Email))
	if err != nil {
		return err
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}
	if password != confirm {
		return errors.New("passwords do not match")
	}
	if err := services.ValidatePassword(password); err != nil {
		var invalid *services.ValidationError
		if errors.As(err, &invalid) {
			return errors.New(invalid.Fields["password"])
		}
		return err
	}

	hash, err := services.HashPassword(password)
	if err != nil {
		return err
	}
	if err := users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return err
	}

	fmt.Printf("password updated for %s\n", user.Email)
	fmt.Println("note: existing sessions remain valid — sign out elsewhere if that matters")
	return nil
}

// stdin is shared across calls. A fresh bufio.Reader per call would discard
// whatever the previous one had already buffered, so a piped password and its
// confirmation would read as one line followed by EOF.
var stdin = bufio.NewReader(os.Stdin)

// readPassword reads without echoing when attached to a terminal, and falls back
// to a plain read when piped (for scripted setup).
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	if term.IsTerminal(int(syscall.Stdin)) {
		raw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}

	line, err := stdin.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
