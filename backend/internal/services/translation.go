package services

import (
	"encoding/json"
	"strings"

	"github.com/kongchansila/portfolio/backend/internal/models"
)

// Entities that can carry translations. A case study is not its own entity:
// its fields hang off the project as `caseStudy.<field>`, so a project and its
// study are always saved, loaded and deleted together.
const (
	EntityProject    = "project"
	EntityExperience = "experience"
	EntityCapability = "capability"
	EntityService    = "service"
	EntityPillar     = "pillar"
	EntityPrinciple  = "principle"
	EntitySetting    = "setting"
)

// DefaultLocale holds the source text. It lives in the content tables
// themselves and is never stored as a translation.
const DefaultLocale = "en"

// SupportedLocales is a closed set. An unknown locale is rejected rather than
// stored, so a typo cannot quietly create a language nothing will ever read.
var SupportedLocales = map[string]bool{
	DefaultLocale: true,
	"km":          true,
}

// TranslatableFields is the closed set of what may be translated, per entity.
//
// Deliberately excludes slugs, URLs, icons, technology names and dates. A slug
// is part of an address and must not change with language; the rest are either
// proper nouns or not words at all.
var TranslatableFields = map[string]map[string]bool{
	EntityProject: {
		"title": true, "category": true, "description": true, "summary": true,
		"imageAlt":           true,
		"caseStudy.overview": true, "caseStudy.context": true, "caseStudy.problem": true,
		"caseStudy.role": true, "caseStudy.approach": true, "caseStudy.architecture": true,
		"caseStudy.database": true, "caseStudy.api": true, "caseStudy.security": true,
		"caseStudy.deployment": true, "caseStudy.outcome": true,
		"caseStudy.features": true, "caseStudy.challenges": true,
	},
	EntityExperience: {
		"label": true, "title": true, "description": true,
		"company": true, "period": true, "location": true,
	},
	EntityCapability: {"title": true, "description": true, "items": true},
	EntityService:    {"title": true, "description": true},
	EntityPrinciple:  {"title": true, "description": true},
	EntityPillar:     {"title": true, "description": true},
	EntitySetting: {
		"name": true, "role": true, "tagline": true, "positioning": true,
		"description": true, "location": true, "about_paragraphs": true,
		"portrait_alt": true, "availability": true,
		"headline": true, "headline_tail": true,
	},
}

// listFields hold a JSON array rather than a sentence.
var listFields = map[string]bool{
	"caseStudy.features": true, "caseStudy.challenges": true, "items": true,
}

// CleanTranslations filters a submitted `locale → field → value` map down to
// what is allowed, and reports anything it had to drop.
//
// Unknown fields and locales are errors rather than silent no-ops: a typo in a
// field name would otherwise look like a successful save that never appears on
// the site.
func CleanTranslations(entity string, incoming map[string]map[string]string) (map[string]map[string]string, error) {
	allowed, ok := TranslatableFields[entity]
	if !ok || len(incoming) == 0 {
		return nil, nil
	}

	fields := map[string]string{}
	cleaned := map[string]map[string]string{}

	for locale, values := range incoming {
		if locale == DefaultLocale {
			// English is the content itself, not an overlay of it.
			fields["translations"] = "English is edited in the fields above, not as a translation."
			continue
		}
		if !SupportedLocales[locale] {
			fields["translations"] = "Unsupported language: " + locale
			continue
		}

		out := map[string]string{}
		for field, value := range values {
			if !allowed[field] {
				fields["translations"] = "Cannot translate field: " + field
				continue
			}
			value = strings.TrimSpace(value)
			if value == "" {
				continue // absent means "fall back to English"
			}
			if listFields[field] && !validJSONList(value) {
				fields["translations."+field] = "Expected a list."
				continue
			}
			out[field] = value
		}
		cleaned[locale] = out
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}
	return cleaned, nil
}

func validJSONList(value string) bool {
	var items []string
	return json.Unmarshal([]byte(value), &items) == nil
}

// Localise overlays a locale's translations onto an English payload, in place.
//
// Field by field: anything without a translation keeps its English text, so a
// partly translated site reads completely rather than showing gaps. This is the
// only sane default when one person is writing both languages by hand.
func Localise(content *models.SiteContent, set models.TranslationSet) {
	if content == nil || len(set) == 0 {
		return
	}

	for key := range content.Settings {
		if value, ok := set.Text(EntitySetting, 0, key); ok {
			content.Settings[key] = value
		}
	}

	for i := range content.Projects {
		p := &content.Projects[i]
		text(set, EntityProject, p.ID, "title", &p.Title)
		text(set, EntityProject, p.ID, "category", &p.Category)
		text(set, EntityProject, p.ID, "description", &p.Description)
		text(set, EntityProject, p.ID, "summary", &p.Summary)
		text(set, EntityProject, p.ID, "imageAlt", &p.ImageAlt)

		if cs := p.CaseStudy; cs != nil {
			text(set, EntityProject, p.ID, "caseStudy.overview", &cs.Overview)
			text(set, EntityProject, p.ID, "caseStudy.context", &cs.Context)
			text(set, EntityProject, p.ID, "caseStudy.problem", &cs.Problem)
			text(set, EntityProject, p.ID, "caseStudy.role", &cs.Role)
			text(set, EntityProject, p.ID, "caseStudy.approach", &cs.Approach)
			text(set, EntityProject, p.ID, "caseStudy.architecture", &cs.Architecture)
			text(set, EntityProject, p.ID, "caseStudy.database", &cs.Database)
			text(set, EntityProject, p.ID, "caseStudy.api", &cs.API)
			text(set, EntityProject, p.ID, "caseStudy.security", &cs.Security)
			text(set, EntityProject, p.ID, "caseStudy.deployment", &cs.Deployment)
			text(set, EntityProject, p.ID, "caseStudy.outcome", &cs.Outcome)
			list(set, EntityProject, p.ID, "caseStudy.features", &cs.Features)
			list(set, EntityProject, p.ID, "caseStudy.challenges", &cs.Challenges)
		}
	}

	for i := range content.Experience {
		e := &content.Experience[i]
		text(set, EntityExperience, e.ID, "label", &e.Label)
		text(set, EntityExperience, e.ID, "title", &e.Title)
		text(set, EntityExperience, e.ID, "description", &e.Description)
		text(set, EntityExperience, e.ID, "company", &e.Company)
		text(set, EntityExperience, e.ID, "period", &e.Period)
		text(set, EntityExperience, e.ID, "location", &e.Location)
	}

	for i := range content.Capabilities {
		c := &content.Capabilities[i]
		text(set, EntityCapability, c.ID, "title", &c.Title)
		text(set, EntityCapability, c.ID, "description", &c.Description)
		list(set, EntityCapability, c.ID, "items", &c.Items)
	}

	for i := range content.Services {
		s := &content.Services[i]
		text(set, EntityService, s.ID, "title", &s.Title)
		text(set, EntityService, s.ID, "description", &s.Description)
	}

	for i := range content.Principles {
		p := &content.Principles[i]
		text(set, EntityPrinciple, p.ID, "title", &p.Title)
		text(set, EntityPrinciple, p.ID, "description", &p.Description)
	}

	for i := range content.Pillars {
		p := &content.Pillars[i]
		text(set, EntityPillar, p.ID, "title", &p.Title)
		text(set, EntityPillar, p.ID, "description", &p.Description)
	}
}

func text(set models.TranslationSet, entity string, id int64, field string, target *string) {
	if value, ok := set.Text(entity, id, field); ok && strings.TrimSpace(value) != "" {
		*target = value
	}
}

// list replaces a whole list or none of it. A malformed value keeps the English
// rather than rendering an empty section.
func list(set models.TranslationSet, entity string, id int64, field string, target *[]string) {
	value, ok := set.Text(entity, id, field)
	if !ok {
		return
	}
	var items []string
	if json.Unmarshal([]byte(value), &items) != nil || len(items) == 0 {
		return
	}
	*target = items
}
