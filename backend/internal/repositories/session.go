package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

type SessionRepository struct{ db *sql.DB }

func NewSessionRepository(db *sql.DB) *SessionRepository { return &SessionRepository{db: db} }

func (r *SessionRepository) Create(ctx context.Context, s *models.Session) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now().UTC()
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at, created_at, last_seen_at, ip_address, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.TokenHash, s.UserID, s.ExpiresAt, now, now, nullable(s.IPAddress), nullable(s.UserAgent)); err != nil {
		return fmt.Errorf("repositories: create session: %w", err)
	}
	return nil
}

// Resolve returns the session and its user in one query, and only if the session
// has not expired and the account is still active — a deactivated account must
// lose access immediately, not at its next login.
func (r *SessionRepository) Resolve(ctx context.Context, tokenHash string) (*models.Session, *models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := r.db.QueryRowContext(ctx,
		`SELECT s.token_hash, s.user_id, s.expires_at, s.created_at, s.last_seen_at,
		        u.id, u.email, u.name, u.password_hash, u.role, u.is_active, u.created_at
		 FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = ? AND s.expires_at > UTC_TIMESTAMP() AND u.is_active = 1`,
		tokenHash)

	var s models.Session
	var u models.User
	if err := row.Scan(
		&s.TokenHash, &s.UserID, &s.ExpiresAt, &s.CreatedAt, &s.LastSeenAt,
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("repositories: resolve session: %w", err)
	}
	return &s, &u, nil
}

func (r *SessionRepository) Touch(ctx context.Context, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE token_hash = ?`, time.Now().UTC(), tokenHash)
	return err
}

func (r *SessionRepository) Delete(ctx context.Context, tokenHash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash); err != nil {
		return fmt.Errorf("repositories: delete session: %w", err)
	}
	return nil
}

// DeleteForUser revokes every session a user holds, used on password change.
func (r *SessionRepository) DeleteForUser(ctx context.Context, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("repositories: delete user sessions: %w", err)
	}
	return nil
}

// PurgeExpired keeps the table from growing without bound.
func (r *SessionRepository) PurgeExpired(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < UTC_TIMESTAMP()`)
	if err != nil {
		return 0, fmt.Errorf("repositories: purge sessions: %w", err)
	}
	return result.RowsAffected()
}
