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

type ContentRepository struct{ db *sql.DB }

func NewContentRepository(db *sql.DB) *ContentRepository { return &ContentRepository{db: db} }

// ── Read ────────────────────────────────────────────────────────────────────

// SiteContent loads everything the public site renders.
//
// Five queries rather than one join: the alternative multiplies every project
// row by its technology count and its case-study text, and the payload is small
// enough that the round trips cost less than the duplication would.
func (r *ContentRepository) SiteContent(ctx context.Context, includeUnpublished bool) (*models.SiteContent, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	settings, err := r.Settings(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := r.Projects(ctx, includeUnpublished)
	if err != nil {
		return nil, err
	}
	experience, err := r.Experience(ctx)
	if err != nil {
		return nil, err
	}
	capabilities, err := r.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	services, err := r.Services(ctx)
	if err != nil {
		return nil, err
	}
	principles, err := r.Principles(ctx)
	if err != nil {
		return nil, err
	}
	pillars, err := r.Pillars(ctx)
	if err != nil {
		return nil, err
	}

	return &models.SiteContent{
		Settings:     settings,
		Projects:     projects,
		Experience:   experience,
		Capabilities: capabilities,
		Services:     services,
		Principles:   principles,
		Pillars:      pillars,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

const projectColumns = `id, slug, title, category, COALESCE(description, ''), COALESCE(summary, ''),
	COALESCE(image, ''), COALESCE(image_alt, ''), COALESCE(year, ''),
	COALESCE(live_url, ''), COALESCE(github_url, ''), featured, published, sort_order, updated_at`

func scanProject(row interface{ Scan(...any) error }) (*models.Project, error) {
	var p models.Project
	if err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Category, &p.Description, &p.Summary,
		&p.Image, &p.ImageAlt, &p.Year, &p.LiveURL, &p.GithubURL,
		&p.Featured, &p.Published, &p.Order, &p.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	p.Technologies = []string{}
	return &p, nil
}

func (r *ContentRepository) Projects(ctx context.Context, includeUnpublished bool) ([]models.Project, error) {
	query := `SELECT ` + projectColumns + ` FROM projects`
	if !includeUnpublished {
		query += ` WHERE published = 1`
	}
	query += ` ORDER BY sort_order, id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repositories: list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	projects := []models.Project{}
	index := map[int64]int{}

	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan project: %w", err)
		}
		index[p.ID] = len(projects)
		projects = append(projects, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return projects, nil
	}

	if err := r.attachTechnologies(ctx, projects, index); err != nil {
		return nil, err
	}
	if err := r.attachCaseStudies(ctx, projects, index); err != nil {
		return nil, err
	}
	return projects, nil
}

// attachTechnologies fills every project's list in one query rather than one per
// project, so adding projects does not add round trips.
func (r *ContentRepository) attachTechnologies(ctx context.Context, projects []models.Project, index map[int64]int) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT project_id, name FROM project_technologies ORDER BY project_id, sort_order, id`)
	if err != nil {
		return fmt.Errorf("repositories: list technologies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var projectID int64
		var name string
		if err := rows.Scan(&projectID, &name); err != nil {
			return fmt.Errorf("repositories: scan technology: %w", err)
		}
		if i, ok := index[projectID]; ok {
			projects[i].Technologies = append(projects[i].Technologies, name)
		}
	}
	return rows.Err()
}

func (r *ContentRepository) attachCaseStudies(ctx context.Context, projects []models.Project, index map[int64]int) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT project_id,
		        COALESCE(overview,''), COALESCE(context,''), COALESCE(problem,''), COALESCE(role,''),
		        COALESCE(approach,''), COALESCE(architecture,''), COALESCE(database_notes,''),
		        COALESCE(api_notes,''), COALESCE(security,''), COALESCE(deployment,''), COALESCE(outcome,''),
		        features, challenges, decisions
		 FROM case_studies`)
	if err != nil {
		return fmt.Errorf("repositories: list case studies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var projectID int64
		var cs models.CaseStudy
		var features, challenges, decisions []byte

		if err := rows.Scan(
			&projectID, &cs.Overview, &cs.Context, &cs.Problem, &cs.Role, &cs.Approach,
			&cs.Architecture, &cs.Database, &cs.API, &cs.Security, &cs.Deployment, &cs.Outcome,
			&features, &challenges, &decisions,
		); err != nil {
			return fmt.Errorf("repositories: scan case study: %w", err)
		}

		// A malformed JSON column must not take the whole site down, so a decode
		// failure leaves that one list empty and the rest of the study intact.
		_ = decodeJSON(features, &cs.Features)
		_ = decodeJSON(challenges, &cs.Challenges)
		_ = decodeJSON(decisions, &cs.Decisions)

		if i, ok := index[projectID]; ok {
			study := cs
			projects[i].CaseStudy = &study
		}
	}
	return rows.Err()
}

func decodeJSON(raw []byte, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func (r *ContentRepository) ProjectBySlug(ctx context.Context, slug string) (*models.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	projects, err := r.Projects(ctx, true)
	if err != nil {
		return nil, err
	}
	for i := range projects {
		if projects[i].Slug == slug {
			return &projects[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *ContentRepository) Settings(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT setting_key, COALESCE(setting_value, '') FROM site_settings`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list settings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	settings := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("repositories: scan setting: %w", err)
		}
		settings[k] = v
	}
	return settings, rows.Err()
}

func (r *ContentRepository) Experience(ctx context.Context) ([]models.ExperienceEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, label, title, COALESCE(company, ''), COALESCE(period, ''),
		        COALESCE(location, ''), is_current, technologies,
		        COALESCE(description, ''), sort_order
		 FROM experience_entries ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list experience: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []models.ExperienceEntry{}
	for rows.Next() {
		var e models.ExperienceEntry
		var technologies []byte
		if err := rows.Scan(&e.ID, &e.Label, &e.Title, &e.Company, &e.Period,
			&e.Location, &e.Current, &technologies, &e.Description, &e.Order); err != nil {
			return nil, fmt.Errorf("repositories: scan experience: %w", err)
		}
		e.Technologies = []string{}
		if err := decodeJSON(technologies, &e.Technologies); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *ContentRepository) Capabilities(ctx context.Context) ([]models.Capability, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, title, description, icon, items, sort_order
		 FROM capabilities ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list capabilities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	list := []models.Capability{}
	for rows.Next() {
		var c models.Capability
		var items []byte
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Description, &c.Icon, &items, &c.Order); err != nil {
			return nil, fmt.Errorf("repositories: scan capability: %w", err)
		}
		c.Items = []string{}
		_ = decodeJSON(items, &c.Items)
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *ContentRepository) Services(ctx context.Context) ([]models.Service, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, title, description, icon, sort_order
		 FROM services ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	list := []models.Service{}
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.Description, &s.Icon, &s.Order); err != nil {
			return nil, fmt.Errorf("repositories: scan service: %w", err)
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *ContentRepository) Principles(ctx context.Context) ([]models.Principle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, title, description, icon, sort_order
		 FROM principles ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list principles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	principles := []models.Principle{}
	for rows.Next() {
		var p models.Principle
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description, &p.Icon, &p.Order); err != nil {
			return nil, fmt.Errorf("repositories: scan principle: %w", err)
		}
		principles = append(principles, p)
	}
	return principles, rows.Err()
}

func (r *ContentRepository) Pillars(ctx context.Context) ([]models.Pillar, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, title, description, sort_order
		 FROM pillars ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("repositories: list pillars: %w", err)
	}
	defer func() { _ = rows.Close() }()

	pillars := []models.Pillar{}
	for rows.Next() {
		var p models.Pillar
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description, &p.Order); err != nil {
			return nil, fmt.Errorf("repositories: scan pillar: %w", err)
		}
		pillars = append(pillars, p)
	}
	return pillars, rows.Err()
}
