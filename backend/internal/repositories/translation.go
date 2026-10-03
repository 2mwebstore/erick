package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

type TranslationRepository struct{ db *sql.DB }

func NewTranslationRepository(db *sql.DB) *TranslationRepository {
	return &TranslationRepository{db: db}
}

// ForLocale loads every translation for one language in a single query.
//
// The public read needs all of them at once to overlay a whole payload, and the
// table is small — one row per translated field — so paging or filtering per
// entity would cost more round trips than it saves.
func (r *TranslationRepository) ForLocale(ctx context.Context, locale string) (models.TranslationSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx,
		`SELECT entity, entity_id, field, value FROM content_translations WHERE locale = ?`, locale)
	if err != nil {
		return nil, fmt.Errorf("repositories: load translations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	set := models.TranslationSet{}
	for rows.Next() {
		var entity, field, value string
		var id int64
		if err := rows.Scan(&entity, &id, &field, &value); err != nil {
			return nil, fmt.Errorf("repositories: scan translation: %w", err)
		}
		set.Put(entity, id, field, value)
	}
	return set, rows.Err()
}

// AllLocales returns every translation, grouped by locale, for the admin panel.
func (r *TranslationRepository) AllLocales(ctx context.Context) (map[string]models.TranslationSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx,
		`SELECT locale, entity, entity_id, field, value FROM content_translations`)
	if err != nil {
		return nil, fmt.Errorf("repositories: load all translations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	all := map[string]models.TranslationSet{}
	for rows.Next() {
		var locale, entity, field, value string
		var id int64
		if err := rows.Scan(&locale, &entity, &id, &field, &value); err != nil {
			return nil, fmt.Errorf("repositories: scan translation: %w", err)
		}
		set, ok := all[locale]
		if !ok {
			set = models.TranslationSet{}
			all[locale] = set
		}
		set.Put(entity, id, field, value)
	}
	return all, rows.Err()
}

// Replace writes the translations for one row in one language.
//
// `values` is the complete set for that row: a field that is absent or blank is
// deleted rather than stored empty, so clearing a Khmer box in the admin panel
// puts the English back instead of rendering nothing. The whole row is written
// in one transaction so a half-saved translation is never visible.
func (r *TranslationRepository) Replace(ctx context.Context, entity string, id int64, locale string, values map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repositories: begin translation write: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	keep := make([]any, 0, len(values)+3)
	placeholders := make([]string, 0, len(values))

	now := time.Now().UTC()
	for field, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO content_translations (entity, entity_id, field, locale, value, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = VALUES(updated_at)`,
			entity, id, field, locale, value, now); err != nil {
			return fmt.Errorf("repositories: save translation: %w", err)
		}
		placeholders = append(placeholders, "?")
		keep = append(keep, field)
	}

	// Remove anything for this row that was not sent, or was sent blank.
	query := `DELETE FROM content_translations WHERE entity = ? AND entity_id = ? AND locale = ?`
	args := []any{entity, id, locale}
	if len(keep) > 0 {
		query += ` AND field NOT IN (` + strings.Join(placeholders, ", ") + `)`
		args = append(args, keep...)
	}
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("repositories: prune translations: %w", err)
	}

	return tx.Commit()
}

// DeleteFor removes every translation of a row, in every language. Called when
// the row itself is deleted: the table has no foreign key, because `entity_id`
// points at six different tables.
func (r *TranslationRepository) DeleteFor(ctx context.Context, entity string, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM content_translations WHERE entity = ? AND entity_id = ?`, entity, id); err != nil {
		return fmt.Errorf("repositories: delete translations: %w", err)
	}
	return nil
}
