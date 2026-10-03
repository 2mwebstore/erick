package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

// ErrNotFound is returned instead of sql.ErrNoRows so callers do not need to
// import database/sql to handle a miss.
var ErrNotFound = errors.New("not found")

type UserRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

const userColumns = `id, email, name, password_hash, role, is_active, last_login_at, created_at`

func scanUser(row interface{ Scan(...any) error }) (*models.User, error) {
	var u models.User
	var lastLogin sql.NullTime

	if err := row.Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}
	return &u, nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ?`,
		strings.ToLower(strings.TrimSpace(email)))

	user, err := scanUser(row)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("repositories: user by email: %w", err)
	}
	return user, err
}

func (r *UserRepository) ByID(ctx context.Context, id int64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = ?`, id))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("repositories: user by id: %w", err)
	}
	return user, err
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `SELECT `+userColumns+` FROM users ORDER BY email`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	users := []models.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan user: %w", err)
		}
		users = append(users, *user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, name, password_hash, role, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		strings.ToLower(strings.TrimSpace(u.Email)), u.Name, u.PasswordHash, u.Role, u.IsActive, now, now)
	if err != nil {
		return 0, fmt.Errorf("repositories: create user: %w", err)
	}
	return result.LastInsertId()
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, hash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		hash, time.Now().UTC(), id); err != nil {
		return fmt.Errorf("repositories: update password: %w", err)
	}
	return nil
}

func (r *UserRepository) TouchLogin(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = ? WHERE id = ?`, time.Now().UTC(), id); err != nil {
		return fmt.Errorf("repositories: touch login: %w", err)
	}
	return nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("repositories: count users: %w", err)
	}
	return n, nil
}
