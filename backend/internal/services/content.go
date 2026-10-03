package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/repositories"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Field limits chosen to match the database columns, so a value that validates
// can always be stored — a validation rule looser than the schema produces a
// 500 instead of a field error.
const (
	titleMax        = 120
	categoryMax     = 80
	slugMax         = 80
	urlMax          = 500
	shortTextMax    = 500
	longTextMax     = 8000
	listItemMax     = 500
	maxListEntries  = 40
	maxTechnologies = 30
)

type ContentService struct {
	repo         *repositories.ContentRepository
	translations *repositories.TranslationRepository
}

func NewContentService(repo *repositories.ContentRepository, translations *repositories.TranslationRepository) *ContentService {
	return &ContentService{repo: repo, translations: translations}
}

var ErrContentUnavailable = errors.New("content: no repository configured")

// Site returns the public payload in one language.
//
// English is the stored text, so it needs no overlay. Any other locale is the
// English payload with its translated fields swapped in; anything untranslated
// stays English rather than rendering blank.
func (s *ContentService) Site(ctx context.Context, includeUnpublished bool, locale string) (*models.SiteContent, error) {
	if s.repo == nil {
		return nil, ErrContentUnavailable
	}

	content, err := s.repo.SiteContent(ctx, includeUnpublished)
	if err != nil {
		return nil, err
	}
	content.Locale = DefaultLocale

	if locale == "" || locale == DefaultLocale || !SupportedLocales[locale] || s.translations == nil {
		return content, nil
	}

	set, err := s.translations.ForLocale(ctx, locale)
	if err != nil {
		// A translation table that cannot be read must not take the site down;
		// English is a complete, correct page on its own.
		return content, nil
	}
	Localise(content, set)
	content.Locale = locale
	return content, nil
}

// SiteForAdmin returns the English payload with every translation attached, so
// the editor can show both languages side by side.
func (s *ContentService) SiteForAdmin(ctx context.Context) (*models.SiteContent, error) {
	content, err := s.Site(ctx, true, DefaultLocale)
	if err != nil || s.translations == nil {
		return content, err
	}

	all, err := s.translations.AllLocales(ctx)
	if err != nil {
		return content, nil
	}

	attach := func(entity string, id int64) map[string]map[string]string {
		out := map[string]map[string]string{}
		for locale, set := range all {
			if fields := set.Fields(entity, id); len(fields) > 0 {
				out[locale] = fields
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}

	for i := range content.Projects {
		content.Projects[i].Translations = attach(EntityProject, content.Projects[i].ID)
	}
	for i := range content.Experience {
		content.Experience[i].Translations = attach(EntityExperience, content.Experience[i].ID)
	}
	for i := range content.Capabilities {
		content.Capabilities[i].Translations = attach(EntityCapability, content.Capabilities[i].ID)
	}
	for i := range content.Services {
		content.Services[i].Translations = attach(EntityService, content.Services[i].ID)
	}
	for i := range content.Principles {
		content.Principles[i].Translations = attach(EntityPrinciple, content.Principles[i].ID)
	}
	for i := range content.Pillars {
		content.Pillars[i].Translations = attach(EntityPillar, content.Pillars[i].ID)
	}
	content.SettingTranslations = attach(EntitySetting, 0)

	return content, nil
}

// saveTranslations persists the overlay for a row it has just written.
//
// A failure here is returned: the English save has already succeeded, but
// silently dropping the Khmer would leave the editor believing it saved.
func (s *ContentService) saveTranslations(ctx context.Context, entity string, id int64, incoming map[string]map[string]string) error {
	if s.translations == nil || id == 0 {
		return nil
	}
	cleaned, err := CleanTranslations(entity, incoming)
	if err != nil {
		return err
	}
	for locale, values := range cleaned {
		if err := s.translations.Replace(ctx, entity, id, locale, values); err != nil {
			return err
		}
	}
	return nil
}

// dropTranslations removes a deleted row's overlay. The table has no foreign
// key — entity_id points at five different tables — so this is the cleanup.
func (s *ContentService) dropTranslations(ctx context.Context, entity string, id int64) {
	if s.translations == nil {
		return
	}
	_ = s.translations.DeleteFor(ctx, entity, id)
}

// SaveProject validates and persists a project.
//
// Validation runs before the storage check so a malformed payload is reported as
// a field error whether or not a repository is configured — and so the rules can
// be tested without a database.
func (s *ContentService) SaveProject(ctx context.Context, p *models.Project) (int64, error) {
	fields := map[string]string{}

	p.Slug = strings.ToLower(strings.TrimSpace(p.Slug))
	p.Title = strings.TrimSpace(p.Title)
	p.Category = strings.TrimSpace(p.Category)

	switch {
	case p.Slug == "":
		fields["slug"] = "A slug is required — it becomes the page URL."
	case len(p.Slug) > slugMax:
		fields["slug"] = fmt.Sprintf("Keep the slug under %d characters.", slugMax)
	case !slugPattern.MatchString(p.Slug):
		fields["slug"] = "Use lowercase letters, numbers and single hyphens, e.g. buffet-system."
	}

	requireText(fields, "title", p.Title, 1, titleMax)
	requireText(fields, "category", p.Category, 1, categoryMax)
	limitText(fields, "description", p.Description, longTextMax)
	limitText(fields, "summary", p.Summary, longTextMax)
	limitText(fields, "image", p.Image, 255)
	limitText(fields, "imageAlt", p.ImageAlt, 255)

	// A half-written URL is worse than none, so reject anything that is not
	// plainly absolute (§14).
	checkURL(fields, "liveUrl", p.LiveURL)
	checkURL(fields, "githubUrl", p.GithubURL)

	if len(p.Technologies) > maxTechnologies {
		fields["technologies"] = fmt.Sprintf("Limit to %d technologies.", maxTechnologies)
	}
	cleaned := make([]string, 0, len(p.Technologies))
	seen := map[string]bool{}
	for _, tech := range p.Technologies {
		tech = strings.TrimSpace(tech)
		if tech == "" {
			continue
		}
		if len(tech) > 60 {
			fields["technologies"] = "Each technology must be under 60 characters."
			continue
		}
		// The table has a unique key on (project_id, name); de-duplicating here
		// turns a 500 into silently sensible behaviour.
		key := strings.ToLower(tech)
		if seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, tech)
	}
	p.Technologies = cleaned

	if p.CaseStudy != nil {
		validateCaseStudy(fields, p.CaseStudy)
	}

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}

	id, err := s.repo.SaveProject(ctx, p)
	if err != nil {
		return id, slugConflict(err)
	}
	// A project's case study is translated as part of the project, so both are
	// written under the one entity and can never disagree about which id they
	// belong to.
	return id, s.saveTranslations(ctx, EntityProject, id, p.Translations)
}

// slugConflict turns a unique-key violation into a field error, so the editor is
// told which value to change rather than being shown a server error.
func slugConflict(err error) error {
	if errors.Is(err, repositories.ErrDuplicate) {
		return &ValidationError{Fields: map[string]string{
			"slug": "Another item already uses that slug. Pick a different one.",
		}}
	}
	return err
}

func validateCaseStudy(fields map[string]string, cs *models.CaseStudy) {
	texts := map[string]*string{
		"overview": &cs.Overview, "context": &cs.Context, "problem": &cs.Problem,
		"role": &cs.Role, "approach": &cs.Approach, "architecture": &cs.Architecture,
		"database": &cs.Database, "api": &cs.API, "security": &cs.Security,
		"deployment": &cs.Deployment, "outcome": &cs.Outcome,
	}
	for name, value := range texts {
		*value = strings.TrimSpace(*value)
		limitText(fields, "caseStudy."+name, *value, longTextMax)
	}

	cs.Features = cleanList(fields, "caseStudy.features", cs.Features)
	cs.Challenges = cleanList(fields, "caseStudy.challenges", cs.Challenges)

	decisions := make([]models.TechnicalDecision, 0, len(cs.Decisions))
	for _, d := range cs.Decisions {
		d.Decision = strings.TrimSpace(d.Decision)
		d.Rationale = strings.TrimSpace(d.Rationale)
		if d.Decision == "" && d.Rationale == "" {
			continue
		}
		limitText(fields, "caseStudy.decisions", d.Decision, listItemMax)
		limitText(fields, "caseStudy.decisions", d.Rationale, longTextMax)
		decisions = append(decisions, d)
	}
	if len(decisions) > maxListEntries {
		fields["caseStudy.decisions"] = fmt.Sprintf("Limit to %d decisions.", maxListEntries)
	}
	cs.Decisions = decisions
}

// cleanList drops blank entries so an empty row in the editor does not become an
// empty bullet on the page.
func cleanList(fields map[string]string, name string, items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		limitText(fields, name, item, listItemMax)
		cleaned = append(cleaned, item)
	}
	if len(cleaned) > maxListEntries {
		fields[name] = fmt.Sprintf("Limit to %d entries.", maxListEntries)
	}
	return cleaned
}

func (s *ContentService) DeleteProject(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeleteProject(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityProject, id)
	return nil
}

// ── Experience ──────────────────────────────────────────────────────────────

func (s *ContentService) SaveExperience(ctx context.Context, e *models.ExperienceEntry) (int64, error) {
	fields := map[string]string{}
	e.Label = strings.TrimSpace(e.Label)
	e.Title = strings.TrimSpace(e.Title)
	e.Description = strings.TrimSpace(e.Description)

	e.Company = strings.TrimSpace(e.Company)
	e.Period = strings.TrimSpace(e.Period)
	e.Location = strings.TrimSpace(e.Location)

	requireText(fields, "label", e.Label, 1, 40)
	requireText(fields, "title", e.Title, 1, 160)
	limitText(fields, "company", e.Company, 160)
	limitText(fields, "period", e.Period, 40)
	limitText(fields, "location", e.Location, 120)
	limitText(fields, "description", e.Description, longTextMax)
	e.Technologies = cleanList(fields, "technologies", e.Technologies)

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}
	id, err := s.repo.SaveExperience(ctx, e)
	if err != nil {
		return id, err
	}
	return id, s.saveTranslations(ctx, EntityExperience, id, e.Translations)
}

func (s *ContentService) DeleteExperience(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeleteExperience(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityExperience, id)
	return nil
}

// ── Capabilities ────────────────────────────────────────────────────────────

func (s *ContentService) SaveCapability(ctx context.Context, c *models.Capability) (int64, error) {
	fields := map[string]string{}
	c.Slug = strings.ToLower(strings.TrimSpace(c.Slug))
	c.Title = strings.TrimSpace(c.Title)
	c.Description = strings.TrimSpace(c.Description)
	c.Icon = strings.TrimSpace(c.Icon)

	if !slugPattern.MatchString(c.Slug) {
		fields["slug"] = "Use lowercase letters, numbers and single hyphens."
	}
	requireText(fields, "title", c.Title, 1, titleMax)
	requireText(fields, "description", c.Description, 1, shortTextMax)
	requireText(fields, "icon", c.Icon, 1, 80)
	c.Items = cleanList(fields, "items", c.Items)

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}
	id, err := s.repo.SaveCapability(ctx, c)
	if err != nil {
		return id, slugConflict(err)
	}
	return id, s.saveTranslations(ctx, EntityCapability, id, c.Translations)
}

func (s *ContentService) DeleteCapability(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeleteCapability(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityCapability, id)
	return nil
}

// ── Services ────────────────────────────────────────────────────────────────

func (s *ContentService) SaveService(ctx context.Context, svc *models.Service) (int64, error) {
	fields := map[string]string{}
	svc.Slug = strings.ToLower(strings.TrimSpace(svc.Slug))
	svc.Title = strings.TrimSpace(svc.Title)
	svc.Description = strings.TrimSpace(svc.Description)
	svc.Icon = strings.TrimSpace(svc.Icon)

	if !slugPattern.MatchString(svc.Slug) {
		fields["slug"] = "Use lowercase letters, numbers and single hyphens."
	}
	requireText(fields, "title", svc.Title, 1, titleMax)
	requireText(fields, "description", svc.Description, 1, shortTextMax)
	requireText(fields, "icon", svc.Icon, 1, 80)

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}
	id, err := s.repo.SaveService(ctx, svc)
	if err != nil {
		return id, slugConflict(err)
	}
	return id, s.saveTranslations(ctx, EntityService, id, svc.Translations)
}

func (s *ContentService) DeleteService(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeleteService(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityService, id)
	return nil
}

// ── Pillars ─────────────────────────────────────────────────────────────────

func (s *ContentService) SavePillar(ctx context.Context, p *models.Pillar) (int64, error) {
	fields := map[string]string{}

	p.Slug = strings.ToLower(strings.TrimSpace(p.Slug))
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)

	if !slugPattern.MatchString(p.Slug) {
		fields["slug"] = "Use lowercase letters, numbers and single hyphens."
	}
	requireText(fields, "title", p.Title, 1, 160)
	requireText(fields, "description", p.Description, 1, longTextMax)

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}

	id, err := s.repo.SavePillar(ctx, p)
	if err != nil {
		return id, slugConflict(err)
	}
	return id, s.saveTranslations(ctx, EntityPillar, id, p.Translations)
}

func (s *ContentService) DeletePillar(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeletePillar(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityPillar, id)
	return nil
}

// ── Principles ──────────────────────────────────────────────────────────────

func (s *ContentService) SavePrinciple(ctx context.Context, p *models.Principle) (int64, error) {
	fields := map[string]string{}

	p.Slug = strings.ToLower(strings.TrimSpace(p.Slug))
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.Icon = strings.TrimSpace(p.Icon)

	if !slugPattern.MatchString(p.Slug) {
		fields["slug"] = "Use lowercase letters, numbers and single hyphens."
	}
	requireText(fields, "title", p.Title, 1, 160)
	requireText(fields, "description", p.Description, 1, longTextMax)
	requireText(fields, "icon", p.Icon, 1, 80)

	if len(fields) > 0 {
		return 0, &ValidationError{Fields: fields}
	}
	if s.repo == nil {
		return 0, ErrContentUnavailable
	}

	id, err := s.repo.SavePrinciple(ctx, p)
	if err != nil {
		return id, slugConflict(err)
	}
	return id, s.saveTranslations(ctx, EntityPrinciple, id, p.Translations)
}

func (s *ContentService) DeletePrinciple(ctx context.Context, id int64) error {
	if s.repo == nil {
		return ErrContentUnavailable
	}
	if err := s.repo.DeletePrinciple(ctx, id); err != nil {
		return err
	}
	s.dropTranslations(ctx, EntityPrinciple, id)
	return nil
}

// ── Settings ────────────────────────────────────────────────────────────────

// AllowedSettings is a closed set.
//
// The table is key/value so new settings need no migration, but accepting
// arbitrary keys from a request would let anyone with an editor account write
// unbounded rows. Adding a setting is a deliberate one-line change here.
var AllowedSettings = map[string]bool{
	"name": true, "role": true, "tagline": true, "positioning": true,
	"description": true, "start_year": true, "location": true, "email": true,
	"portrait": true, "portrait_alt": true, "resume_file": true,
	"hero_stack": true, "profiles": true, "about_paragraphs": true,
	"wordpress_seo_stack": true, "contact_enabled": true,
	"phone": true, "availability": true,
	// The hero headline, split so the second half can be set in the quieter
	// colour without markup in the value.
	"headline": true, "headline_tail": true,
}

func (s *ContentService) SaveSettings(ctx context.Context, values map[string]string) error {
	fields := map[string]string{}
	accepted := map[string]string{}

	for key, value := range values {
		if !AllowedSettings[key] {
			fields[key] = "Unknown setting."
			continue
		}
		if len(value) > longTextMax {
			fields[key] = fmt.Sprintf("Keep this under %d characters.", longTextMax)
			continue
		}
		accepted[key] = strings.TrimSpace(value)
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	if len(accepted) == 0 {
		return nil
	}
	if s.repo == nil {
		return ErrContentUnavailable
	}
	return s.repo.SaveSettings(ctx, accepted)
}

// SaveSettingTranslations stores the translated site settings.
//
// Settings are key/value with no row id, so they hang off entity id 0 with the
// setting key as the field name.
func (s *ContentService) SaveSettingTranslations(ctx context.Context, incoming map[string]map[string]string) error {
	if s.translations == nil || len(incoming) == 0 {
		return nil
	}
	cleaned, err := CleanTranslations(EntitySetting, incoming)
	if err != nil {
		return err
	}
	for locale, values := range cleaned {
		if err := s.translations.Replace(ctx, EntitySetting, 0, locale, values); err != nil {
			return err
		}
	}
	return nil
}

// ── Shared validation helpers ───────────────────────────────────────────────

func requireText(fields map[string]string, name, value string, min, max int) {
	runes := len([]rune(value))
	switch {
	case runes < min:
		fields[name] = "This field is required."
	case runes > max:
		fields[name] = fmt.Sprintf("Keep this under %d characters.", max)
	}
}

func limitText(fields map[string]string, name, value string, max int) {
	if len([]rune(value)) > max {
		fields[name] = fmt.Sprintf("Keep this under %d characters.", max)
	}
}

func checkURL(fields map[string]string, name, value string) {
	if value == "" {
		return
	}
	if len(value) > urlMax {
		fields[name] = fmt.Sprintf("Keep the URL under %d characters.", urlMax)
		return
	}
	if !strings.HasPrefix(value, "https://") && !strings.HasPrefix(value, "http://") {
		fields[name] = "Include the full address, starting with https://"
	}
}
