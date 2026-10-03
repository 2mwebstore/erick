// Package repositories holds SQL access. No business rules live here.
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create inserts a contact message and returns its id.
//
// Every value is a bound parameter — the statement text never contains user
// input, so there is no injection surface here (§34).
func (r *ContactRepository) Create(ctx context.Context, msg *models.ContactMessage) (int64, error) {
	const query = `
		INSERT INTO contact_messages
			(name, email, phone, subject, project_type, message, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg.CreatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		msg.Name, msg.Email, nullable(msg.Phone), nullable(msg.Subject), msg.ProjectType, msg.Message,
		nullable(msg.IPAddress), nullable(msg.UserAgent), msg.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("repositories: insert contact message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("repositories: last insert id: %w", err)
	}

	return id, nil
}

// CountSince reports submissions in a window, used by the health endpoint and
// by operational checks for a sudden spike.
func (r *ContactRepository) CountSince(ctx context.Context, since time.Time) (int, error) {
	const query = `SELECT COUNT(*) FROM contact_messages WHERE created_at >= ?`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var count int
	if err := r.db.QueryRowContext(ctx, query, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("repositories: count contact messages: %w", err)
	}
	return count, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
