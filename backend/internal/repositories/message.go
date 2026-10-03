package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

// MessageFilter narrows the admin inbox listing.
type MessageFilter struct {
	UnreadOnly bool
	Archived   bool
	Limit      int
	Offset     int
}

type MessageListing struct {
	Messages []models.ContactMessage `json:"messages"`
	Total    int                     `json:"total"`
	Unread   int                     `json:"unread"`
}

// List returns a page of messages plus the counts the inbox header needs.
func (r *ContactRepository) List(ctx context.Context, f MessageFilter) (*MessageListing, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := `WHERE archived_at IS NULL`
	if f.Archived {
		where = `WHERE archived_at IS NOT NULL`
	}
	if f.UnreadOnly {
		where += ` AND read_at IS NULL`
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, COALESCE(phone, ''), COALESCE(subject, ''),
		        project_type, message,
		        COALESCE(ip_address, ''), COALESCE(user_agent, ''),
		        created_at, read_at, archived_at
		 FROM contact_messages `+where+`
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("repositories: list messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	listing := &MessageListing{Messages: []models.ContactMessage{}}

	for rows.Next() {
		var m models.ContactMessage
		var readAt, archivedAt sql.NullTime

		if err := rows.Scan(
			&m.ID, &m.Name, &m.Email, &m.Phone, &m.Subject, &m.ProjectType, &m.Message,
			&m.IPAddress, &m.UserAgent, &m.CreatedAt, &readAt, &archivedAt,
		); err != nil {
			return nil, fmt.Errorf("repositories: scan message: %w", err)
		}
		if readAt.Valid {
			m.ReadAt = &readAt.Time
		}
		if archivedAt.Valid {
			m.ArchivedAt = &archivedAt.Time
		}
		listing.Messages = append(listing.Messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM contact_messages `+where).Scan(&listing.Total); err != nil {
		return nil, fmt.Errorf("repositories: count messages: %w", err)
	}
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM contact_messages WHERE read_at IS NULL AND archived_at IS NULL`,
	).Scan(&listing.Unread); err != nil {
		return nil, fmt.Errorf("repositories: count unread: %w", err)
	}

	return listing, nil
}

// SetRead marks a message read or unread.
func (r *ContactRepository) SetRead(ctx context.Context, id int64, read bool) error {
	return r.setTimestamp(ctx, id, "read_at", read)
}

// SetArchived moves a message out of the inbox, or back into it.
func (r *ContactRepository) SetArchived(ctx context.Context, id int64, archived bool) error {
	return r.setTimestamp(ctx, id, "archived_at", archived)
}

// setTimestamp sets a nullable timestamp column to now, or clears it. The column
// name is a constant from this package, never from a request.
func (r *ContactRepository) setTimestamp(ctx context.Context, id int64, column string, set bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var value any
	if set {
		value = time.Now().UTC()
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE contact_messages SET `+column+` = ? WHERE id = ?`, value, id)
	if err != nil {
		return fmt.Errorf("repositories: update %s: %w", column, err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContactRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM contact_messages WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("repositories: delete message: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// All returns every non-archived message for CSV export.
func (r *ContactRepository) All(ctx context.Context) ([]models.ContactMessage, error) {
	listing, err := r.List(ctx, MessageFilter{Limit: 200})
	if err != nil {
		return nil, err
	}
	return listing.Messages, nil
}
