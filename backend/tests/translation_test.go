package tests

import (
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/services"
)

func kmSet(entries ...[4]any) models.TranslationSet {
	set := models.TranslationSet{}
	for _, e := range entries {
		set.Put(e[0].(string), int64(e[1].(int)), e[2].(string), e[3].(string))
	}
	return set
}

func TestLocaliseSwapsTranslatedFieldsOnly(t *testing.T) {
	content := &models.SiteContent{
		Settings: map[string]string{"role": "Full-Stack Software Developer", "location": "Cambodia"},
		Projects: []models.Project{{
			ID: 7, Slug: "buffet-system", Title: "Buffet System", Category: "Restaurant / POS System",
			Description: "English description", Summary: "English summary",
			CaseStudy: &models.CaseStudy{Overview: "English overview", Role: "English role"},
		}},
	}

	services.Localise(content, kmSet(
		[4]any{services.EntitySetting, 0, "role", "អ្នកអភិវឌ្ឍន៍កម្មវិធី"},
		[4]any{services.EntityProject, 7, "title", "ប្រព័ន្ធប៊ូហ្វេ"},
		[4]any{services.EntityProject, 7, "caseStudy.overview", "ទិដ្ឋភាពទូទៅ"},
	))

	if got := content.Settings["role"]; got != "អ្នកអភិវឌ្ឍន៍កម្មវិធី" {
		t.Errorf("role not translated: %q", got)
	}
	// Untranslated fields must keep their English, not blank out.
	if got := content.Settings["location"]; got != "Cambodia" {
		t.Errorf("untranslated setting changed: %q", got)
	}
	if got := content.Projects[0].Title; got != "ប្រព័ន្ធប៊ូហ្វេ" {
		t.Errorf("title not translated: %q", got)
	}
	if got := content.Projects[0].Description; got != "English description" {
		t.Errorf("untranslated field changed: %q", got)
	}
	if got := content.Projects[0].CaseStudy.Overview; got != "ទិដ្ឋភាពទូទៅ" {
		t.Errorf("case study not translated: %q", got)
	}
	if got := content.Projects[0].CaseStudy.Role; got != "English role" {
		t.Errorf("untranslated case study field changed: %q", got)
	}
	// The slug is an address, not prose, and must never move with the language.
	if got := content.Projects[0].Slug; got != "buffet-system" {
		t.Errorf("slug changed: %q", got)
	}
}

func TestLocaliseIsANoOpWithoutTranslations(t *testing.T) {
	content := &models.SiteContent{
		Settings: map[string]string{"role": "Developer"},
		Services: []models.Service{{ID: 1, Title: "Web Applications"}},
	}
	services.Localise(content, models.TranslationSet{})

	if content.Settings["role"] != "Developer" || content.Services[0].Title != "Web Applications" {
		t.Error("empty translation set must leave the payload untouched")
	}
}

func TestLocaliseReplacesWholeListsOrNone(t *testing.T) {
	content := &models.SiteContent{
		Capabilities: []models.Capability{{ID: 3, Items: []string{"Go", "MySQL"}}},
	}

	// Malformed JSON keeps the English list rather than emptying the section.
	services.Localise(content, kmSet([4]any{services.EntityCapability, 3, "items", "not json"}))
	if len(content.Capabilities[0].Items) != 2 {
		t.Errorf("malformed list should have been ignored, got %v", content.Capabilities[0].Items)
	}

	services.Localise(content, kmSet([4]any{services.EntityCapability, 3, "items", `["ហ្គោ","មីស្គល"]`}))
	if len(content.Capabilities[0].Items) != 2 || content.Capabilities[0].Items[0] != "ហ្គោ" {
		t.Errorf("list not translated: %v", content.Capabilities[0].Items)
	}
}

func TestCleanTranslationsRejectsUnknownFieldsAndLocales(t *testing.T) {
	for name, incoming := range map[string]map[string]map[string]string{
		"unknown field":  {"km": {"slug": "something"}},
		"unknown locale": {"fr": {"title": "titre"}},
		"english":        {"en": {"title": "English belongs in the row itself"}},
	} {
		if _, err := services.CleanTranslations(services.EntityProject, incoming); err == nil {
			t.Errorf("%s: expected a validation error", name)
		}
	}
}

func TestCleanTranslationsDropsBlanksSoEnglishShowsThrough(t *testing.T) {
	cleaned, err := services.CleanTranslations(services.EntityProject, map[string]map[string]string{
		"km": {"title": "ប្រព័ន្ធ", "summary": "   ", "description": ""},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cleaned["km"]["summary"]; ok {
		t.Error("a blank translation must be dropped, not stored empty")
	}
	if cleaned["km"]["title"] != "ប្រព័ន្ធ" {
		t.Errorf("title lost: %v", cleaned["km"])
	}
}

func TestCleanTranslationsRequiresAListForListFields(t *testing.T) {
	if _, err := services.CleanTranslations(services.EntityCapability, map[string]map[string]string{
		"km": {"items": "just a sentence"},
	}); err == nil {
		t.Error("a list field given a sentence should be rejected")
	}
}

func TestTranslatableFieldsExcludeAddressesAndProperNouns(t *testing.T) {
	// A slug is part of a URL and an icon is not a word; translating either
	// would break links or render nothing.
	for entity, banned := range map[string][]string{
		services.EntityProject:    {"slug", "liveUrl", "githubUrl", "technologies", "year"},
		services.EntityCapability: {"slug", "icon"},
		services.EntityService:    {"slug", "icon"},
	} {
		for _, field := range banned {
			if services.TranslatableFields[entity][field] {
				t.Errorf("%s.%s must not be translatable", entity, field)
			}
		}
	}
}
