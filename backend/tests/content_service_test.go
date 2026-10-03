package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

// The content service validates before it touches storage, so these run without
// a database. ErrContentUnavailable is what a *valid* payload returns here.
func contentService() *services.ContentService {
	return services.NewContentService(nil, nil)
}

func saveProject(t *testing.T, p *models.Project) (map[string]string, error) {
	t.Helper()

	_, err := contentService().SaveProject(context.Background(), p)

	var invalid *services.ValidationError
	if errors.As(err, &invalid) {
		return invalid.Fields, err
	}
	return nil, err
}

func validProject() *models.Project {
	return &models.Project{
		Slug:         "buffet-system",
		Title:        "Buffet System",
		Category:     "Restaurant / POS System",
		Technologies: []string{"Flutter"},
	}
}

func TestSaveProjectAcceptsAValidPayload(t *testing.T) {
	fields, err := saveProject(t, validProject())
	if fields != nil {
		t.Fatalf("expected no field errors, got %v", fields)
	}
	if !errors.Is(err, services.ErrContentUnavailable) {
		t.Fatalf("expected ErrContentUnavailable once validation passed, got %v", err)
	}
}

func TestSaveProjectRejectsBadSlugs(t *testing.T) {
	for _, slug := range []string{"", "Bad Slug", "trailing-", "-leading", "double--hyphen", "sym!bol"} {
		t.Run(slug, func(t *testing.T) {
			p := validProject()
			p.Slug = slug

			fields, _ := saveProject(t, p)
			if fields == nil || fields["slug"] == "" {
				t.Fatalf("expected a slug error for %q, got %v", slug, fields)
			}
		})
	}
}

func TestSaveProjectNormalisesSlugCase(t *testing.T) {
	// An uppercase slug is lowercased rather than rejected — the URL is the
	// lowercase form either way, so refusing it would be pedantry.
	p := validProject()
	p.Slug = "Buffet-System"

	if fields, _ := saveProject(t, p); fields != nil {
		t.Fatalf("expected the slug to be normalised, got %v", fields)
	}
	if p.Slug != "buffet-system" {
		t.Fatalf("expected buffet-system, got %q", p.Slug)
	}
}

func TestSaveProjectAcceptsOrdinarySlugs(t *testing.T) {
	for _, slug := range []string{"portfolio", "buffet-system", "check-in-app", "a1", "v2-rewrite"} {
		p := validProject()
		p.Slug = slug

		if fields, _ := saveProject(t, p); fields != nil {
			t.Errorf("expected %q to be accepted, got %v", slug, fields)
		}
	}
}

func TestSaveProjectRequiresTitleAndCategory(t *testing.T) {
	p := validProject()
	p.Title = "  "
	p.Category = ""

	fields, _ := saveProject(t, p)
	if fields["title"] == "" || fields["category"] == "" {
		t.Fatalf("expected title and category errors, got %v", fields)
	}
}

func TestSaveProjectRejectsRelativeURLs(t *testing.T) {
	// A half-written URL is worse than none: it renders as a dead link (§14).
	for _, url := range []string{"example.com", "/work/thing", "www.example.com"} {
		p := validProject()
		p.LiveURL = url

		fields, _ := saveProject(t, p)
		if fields["liveUrl"] == "" {
			t.Errorf("expected %q to be rejected, got %v", url, fields)
		}
	}
}

func TestSaveProjectAcceptsAbsoluteURLs(t *testing.T) {
	p := validProject()
	p.LiveURL = "https://example.com"
	p.GithubURL = "https://github.com/example/repo"

	if fields, _ := saveProject(t, p); fields != nil {
		t.Fatalf("expected absolute URLs to pass, got %v", fields)
	}
}

func TestSaveProjectDeduplicatesAndTrimsTechnologies(t *testing.T) {
	p := validProject()
	p.Technologies = []string{" Go ", "go", "GO", "", "Nuxt", "Nuxt"}

	if fields, _ := saveProject(t, p); fields != nil {
		t.Fatalf("expected no errors, got %v", fields)
	}

	// The table has a unique key on (project_id, name); de-duplicating here turns
	// what would be a 500 into sensible behaviour.
	if len(p.Technologies) != 2 || p.Technologies[0] != "Go" || p.Technologies[1] != "Nuxt" {
		t.Fatalf("expected [Go Nuxt], got %v", p.Technologies)
	}
}

func TestSaveProjectAllowsAnEmptyTechnologyList(t *testing.T) {
	// Buffet System ships this way: an unknown stack is left empty rather than
	// guessed (§27).
	p := validProject()
	p.Technologies = nil

	if fields, _ := saveProject(t, p); fields != nil {
		t.Fatalf("expected an empty list to be allowed, got %v", fields)
	}
}

func TestSaveProjectCleansCaseStudyLists(t *testing.T) {
	p := validProject()
	p.CaseStudy = &models.CaseStudy{
		Features:   []string{"Real feature", "   ", ""},
		Challenges: []string{""},
		Decisions: []models.TechnicalDecision{
			{Decision: "Chose X", Rationale: "Because Y"},
			{Decision: "  ", Rationale: "  "},
		},
	}

	if fields, _ := saveProject(t, p); fields != nil {
		t.Fatalf("expected no errors, got %v", fields)
	}

	if len(p.CaseStudy.Features) != 1 {
		t.Errorf("blank features should be dropped, got %v", p.CaseStudy.Features)
	}
	if len(p.CaseStudy.Challenges) != 0 {
		t.Errorf("an all-blank list should end up empty, got %v", p.CaseStudy.Challenges)
	}
	if len(p.CaseStudy.Decisions) != 1 {
		t.Errorf("an empty decision row should be dropped, got %v", p.CaseStudy.Decisions)
	}
}

func TestSaveProjectEnforcesColumnLengths(t *testing.T) {
	// Validation looser than the schema produces a 500 instead of a field error.
	p := validProject()
	p.Title = strings.Repeat("a", 200)

	fields, _ := saveProject(t, p)
	if fields["title"] == "" {
		t.Fatalf("expected a length error, got %v", fields)
	}
}

func TestSaveSettingsRejectsUnknownKeys(t *testing.T) {
	// The table is key/value, so an open key set would let any editor write
	// unbounded rows.
	err := contentService().SaveSettings(context.Background(), map[string]string{
		"name":     "Kong Chansila",
		"evil_key": "anything",
	})

	var invalid *services.ValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected a validation error, got %v", err)
	}
	if invalid.Fields["evil_key"] == "" {
		t.Fatalf("expected an error naming the unknown key, got %v", invalid.Fields)
	}
}

func TestSaveSettingsAcceptsEveryDocumentedKey(t *testing.T) {
	values := map[string]string{}
	for key := range services.AllowedSettings {
		values[key] = "value"
	}

	err := contentService().SaveSettings(context.Background(), values)
	if !errors.Is(err, services.ErrContentUnavailable) {
		t.Fatalf("expected the allowed keys to validate, got %v", err)
	}
}

func TestSaveCapabilityAndServiceValidateSlugs(t *testing.T) {
	svc := contentService()

	_, err := svc.SaveCapability(context.Background(), &models.Capability{Slug: "Bad Slug", Title: "T", Description: "D", Icon: "i"})
	var invalid *services.ValidationError
	if !errors.As(err, &invalid) || invalid.Fields["slug"] == "" {
		t.Fatalf("expected a capability slug error, got %v", err)
	}

	_, err = svc.SaveService(context.Background(), &models.Service{Slug: "Bad Slug", Title: "T", Description: "D", Icon: "i"})
	if !errors.As(err, &invalid) || invalid.Fields["slug"] == "" {
		t.Fatalf("expected a service slug error, got %v", err)
	}
}

func TestSaveExperienceRequiresLabelAndTitle(t *testing.T) {
	_, err := contentService().SaveExperience(context.Background(), &models.ExperienceEntry{})

	var invalid *services.ValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected a validation error, got %v", err)
	}
	if invalid.Fields["label"] == "" || invalid.Fields["title"] == "" {
		t.Fatalf("expected label and title errors, got %v", invalid.Fields)
	}
}
