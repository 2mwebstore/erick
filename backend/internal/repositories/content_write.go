package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

// SaveProject inserts or updates a project with its technologies and case study
// in one transaction.
//
// A project is only coherent with its child rows: replacing the technology list
// means deleting and re-inserting, and a failure halfway would otherwise leave a
// project with no technologies at all.
func (r *ContentRepository) SaveProject(ctx context.Context, p *models.Project) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("repositories: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once Commit has succeeded

	now := time.Now().UTC()

	if p.ID == 0 {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO projects
			   (slug, title, category, description, summary, image, image_alt, year,
			    live_url, github_url, featured, published, sort_order, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.Slug, p.Title, p.Category, nullable(p.Description), nullable(p.Summary),
			nullable(p.Image), nullable(p.ImageAlt), nullable(p.Year),
			nullable(p.LiveURL), nullable(p.GithubURL),
			p.Featured, p.Published, p.Order, now, now)
		if err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: insert project: %w", err)
		}
		if p.ID, err = result.LastInsertId(); err != nil {
			return 0, err
		}
	} else {
		if _, err := tx.ExecContext(ctx,
			`UPDATE projects SET
			   slug = ?, title = ?, category = ?, description = ?, summary = ?,
			   image = ?, image_alt = ?, year = ?, live_url = ?, github_url = ?,
			   featured = ?, published = ?, sort_order = ?, updated_at = ?
			 WHERE id = ?`,
			p.Slug, p.Title, p.Category, nullable(p.Description), nullable(p.Summary),
			nullable(p.Image), nullable(p.ImageAlt), nullable(p.Year),
			nullable(p.LiveURL), nullable(p.GithubURL),
			p.Featured, p.Published, p.Order, now, p.ID); err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: update project: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_technologies WHERE project_id = ?`, p.ID); err != nil {
		return 0, fmt.Errorf("repositories: clear technologies: %w", err)
	}
	for i, tech := range p.Technologies {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_technologies (project_id, name, sort_order) VALUES (?, ?, ?)`,
			p.ID, tech, i); err != nil {
			return 0, fmt.Errorf("repositories: insert technology %q: %w", tech, err)
		}
	}

	if p.CaseStudy == nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM case_studies WHERE project_id = ?`, p.ID); err != nil {
			return 0, fmt.Errorf("repositories: clear case study: %w", err)
		}
	} else if err := saveCaseStudy(ctx, tx, p.ID, p.CaseStudy, now); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("repositories: commit project: %w", err)
	}
	return p.ID, nil
}

func saveCaseStudy(ctx context.Context, tx *sql.Tx, projectID int64, cs *models.CaseStudy, now time.Time) error {
	features, err := encodeList(cs.Features)
	if err != nil {
		return err
	}
	challenges, err := encodeList(cs.Challenges)
	if err != nil {
		return err
	}
	decisions, err := encodeList(cs.Decisions)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO case_studies
		   (project_id, overview, context, problem, role, approach, architecture,
		    database_notes, api_notes, security, deployment, outcome,
		    features, challenges, decisions, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   overview = VALUES(overview), context = VALUES(context), problem = VALUES(problem),
		   role = VALUES(role), approach = VALUES(approach), architecture = VALUES(architecture),
		   database_notes = VALUES(database_notes), api_notes = VALUES(api_notes),
		   security = VALUES(security), deployment = VALUES(deployment), outcome = VALUES(outcome),
		   features = VALUES(features), challenges = VALUES(challenges),
		   decisions = VALUES(decisions), updated_at = VALUES(updated_at)`,
		projectID, nullable(cs.Overview), nullable(cs.Context), nullable(cs.Problem), nullable(cs.Role),
		nullable(cs.Approach), nullable(cs.Architecture), nullable(cs.Database), nullable(cs.API),
		nullable(cs.Security), nullable(cs.Deployment), nullable(cs.Outcome),
		features, challenges, decisions, now)
	if err != nil {
		return fmt.Errorf("repositories: save case study: %w", err)
	}
	return nil
}

// encodeList stores an empty list as NULL rather than "[]", so "no features" and
// "features not filled in yet" read the same way in the database as they do in
// the UI.
func encodeList(value any) (any, error) {
	switch v := value.(type) {
	case []string:
		if len(v) == 0 {
			return nil, nil
		}
	case []models.TechnicalDecision:
		if len(v) == 0 {
			return nil, nil
		}
	case nil:
		return nil, nil
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("repositories: encode list: %w", err)
	}
	return string(encoded), nil
}

func (r *ContentRepository) DeleteProject(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Technologies and the case study cascade from the foreign keys.
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("repositories: delete project: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Experience ──────────────────────────────────────────────────────────────

func (r *ContentRepository) SaveExperience(ctx context.Context, e *models.ExperienceEntry) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	technologies, err := json.Marshal(e.Technologies)
	if err != nil {
		return 0, fmt.Errorf("repositories: encode experience technologies: %w", err)
	}

	if e.ID == 0 {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO experience_entries
			   (label, title, company, period, location, is_current, technologies, description, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.Label, e.Title, nullable(e.Company), nullable(e.Period), nullable(e.Location),
			e.Current, technologies, nullable(e.Description), e.Order)
		if err != nil {
			return 0, fmt.Errorf("repositories: insert experience: %w", err)
		}
		return result.LastInsertId()
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE experience_entries
		 SET label = ?, title = ?, company = ?, period = ?, location = ?,
		     is_current = ?, technologies = ?, description = ?, sort_order = ?
		 WHERE id = ?`,
		e.Label, e.Title, nullable(e.Company), nullable(e.Period), nullable(e.Location),
		e.Current, technologies, nullable(e.Description), e.Order, e.ID); err != nil {
		return 0, fmt.Errorf("repositories: update experience: %w", err)
	}
	return e.ID, nil
}

func (r *ContentRepository) DeleteExperience(ctx context.Context, id int64) error {
	return r.deleteByID(ctx, "experience_entries", id)
}

// ── Capabilities ────────────────────────────────────────────────────────────

func (r *ContentRepository) SaveCapability(ctx context.Context, c *models.Capability) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, err := json.Marshal(c.Items)
	if err != nil {
		return 0, fmt.Errorf("repositories: encode capability items: %w", err)
	}

	if c.ID == 0 {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO capabilities (slug, title, description, icon, items, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			c.Slug, c.Title, c.Description, c.Icon, string(items), c.Order)
		if err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: insert capability: %w", err)
		}
		return result.LastInsertId()
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE capabilities SET slug = ?, title = ?, description = ?, icon = ?, items = ?, sort_order = ?
		 WHERE id = ?`,
		c.Slug, c.Title, c.Description, c.Icon, string(items), c.Order, c.ID); err != nil {
		if errors.Is(asDuplicate(err), ErrDuplicate) {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("repositories: update capability: %w", err)
	}
	return c.ID, nil
}

func (r *ContentRepository) DeleteCapability(ctx context.Context, id int64) error {
	return r.deleteByID(ctx, "capabilities", id)
}

// ── Services ────────────────────────────────────────────────────────────────

func (r *ContentRepository) SaveService(ctx context.Context, s *models.Service) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if s.ID == 0 {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO services (slug, title, description, icon, sort_order) VALUES (?, ?, ?, ?, ?)`,
			s.Slug, s.Title, s.Description, s.Icon, s.Order)
		if err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: insert service: %w", err)
		}
		return result.LastInsertId()
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE services SET slug = ?, title = ?, description = ?, icon = ?, sort_order = ? WHERE id = ?`,
		s.Slug, s.Title, s.Description, s.Icon, s.Order, s.ID); err != nil {
		if errors.Is(asDuplicate(err), ErrDuplicate) {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("repositories: update service: %w", err)
	}
	return s.ID, nil
}

func (r *ContentRepository) DeleteService(ctx context.Context, id int64) error {
	return r.deleteByID(ctx, "services", id)
}

// ── Settings ────────────────────────────────────────────────────────────────

// SaveSettings upserts a batch in one transaction, so a partially applied
// settings save cannot happen.
func (r *ContentRepository) SaveSettings(ctx context.Context, values map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repositories: begin settings: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	for key, value := range values {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO site_settings (setting_key, setting_value, updated_at) VALUES (?, ?, ?)
			 ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value), updated_at = VALUES(updated_at)`,
			key, value, now); err != nil {
			return fmt.Errorf("repositories: save setting %q: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repositories: commit settings: %w", err)
	}
	return nil
}

// deleteByID is safe from injection because the table name is a compile-time
// constant supplied by this package, never by a request.
func (r *ContentRepository) deleteByID(ctx context.Context, table string, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM `+table+` WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("repositories: delete from %s: %w", table, err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Principles ──────────────────────────────────────────────────────────────

func (r *ContentRepository) SavePrinciple(ctx context.Context, p *models.Principle) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if p.ID == 0 {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO principles (slug, title, description, icon, sort_order) VALUES (?, ?, ?, ?, ?)`,
			p.Slug, p.Title, p.Description, p.Icon, p.Order)
		if err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: insert principle: %w", err)
		}
		return result.LastInsertId()
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE principles SET slug = ?, title = ?, description = ?, icon = ?, sort_order = ? WHERE id = ?`,
		p.Slug, p.Title, p.Description, p.Icon, p.Order, p.ID); err != nil {
		if errors.Is(asDuplicate(err), ErrDuplicate) {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("repositories: update principle: %w", err)
	}
	return p.ID, nil
}

func (r *ContentRepository) DeletePrinciple(ctx context.Context, id int64) error {
	return r.deleteByID(ctx, "principles", id)
}

// ── Pillars ─────────────────────────────────────────────────────────────────

func (r *ContentRepository) SavePillar(ctx context.Context, p *models.Pillar) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if p.ID == 0 {
		result, err := r.db.ExecContext(ctx,
			`INSERT INTO pillars (slug, title, description, sort_order) VALUES (?, ?, ?, ?)`,
			p.Slug, p.Title, p.Description, p.Order)
		if err != nil {
			if errors.Is(asDuplicate(err), ErrDuplicate) {
				return 0, ErrDuplicate
			}
			return 0, fmt.Errorf("repositories: insert pillar: %w", err)
		}
		return result.LastInsertId()
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE pillars SET slug = ?, title = ?, description = ?, sort_order = ? WHERE id = ?`,
		p.Slug, p.Title, p.Description, p.Order, p.ID); err != nil {
		if errors.Is(asDuplicate(err), ErrDuplicate) {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("repositories: update pillar: %w", err)
	}
	return p.ID, nil
}

func (r *ContentRepository) DeletePillar(ctx context.Context, id int64) error {
	return r.deleteByID(ctx, "pillars", id)
}
