package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

type AuditRepository struct{ db *sql.DB }

func NewAuditRepository(db *sql.DB) *AuditRepository { return &AuditRepository{db: db} }

// Record writes an audit entry.
//
// actor_email is stored alongside the user id because the log must still say who
// did what after that account has been deleted.
func (r *AuditRepository) Record(ctx context.Context, e *models.AuditEntry) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var detail any
	if e.Detail != nil {
		encoded, err := json.Marshal(e.Detail)
		if err != nil {
			return fmt.Errorf("repositories: encode audit detail: %w", err)
		}
		detail = string(encoded)
	}

	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_log (user_id, actor_email, action, entity, entity_id, detail, ip_address, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.UserID, e.ActorEmail, e.Action, e.Entity, nullable(e.EntityID), detail,
		nullable(e.IPAddress), time.Now().UTC()); err != nil {
		return fmt.Errorf("repositories: record audit: %w", err)
	}
	return nil
}

// AuditListing is one page of the log plus the count the pager needs to know
// how many pages there are.
type AuditListing struct {
	Entries []models.AuditEntry `json:"entries"`
	Total   int                 `json:"total"`
}

// Recent returns one page of the audit log, newest first.
//
// The log only grows, so it is paged at the database rather than in the browser.
func (r *AuditRepository) Recent(ctx context.Context, limit, offset int) (*AuditListing, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, actor_email, action, entity, COALESCE(entity_id, ''),
		        detail, COALESCE(ip_address, ''), created_at
		 FROM audit_log ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repositories: recent audit: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []models.AuditEntry{}
	for rows.Next() {
		var e models.AuditEntry
		var userID sql.NullInt64
		var detail []byte

		if err := rows.Scan(&e.ID, &userID, &e.ActorEmail, &e.Action, &e.Entity,
			&e.EntityID, &detail, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("repositories: scan audit: %w", err)
		}
		if userID.Valid {
			id := userID.Int64
			e.UserID = &id
		}
		if len(detail) > 0 {
			var parsed any
			if json.Unmarshal(detail, &parsed) == nil {
				e.Detail = parsed
			}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	listing := &AuditListing{Entries: entries}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&listing.Total); err != nil {
		return nil, fmt.Errorf("repositories: count audit: %w", err)
	}
	return listing, nil
}
