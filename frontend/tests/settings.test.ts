import { describe, expect, it } from 'vitest'
import { resolveSite } from '~/utils/settings'
import { fallbackContent } from '~/content/fallback'
import { services as fallbackServices } from '~/content/services'

describe('settings resolver', () => {
  it('resolves the bundled fallback into a complete site object', () => {
    const site = resolveSite(fallbackContent.settings)

    expect(site.name).toBe('Kong Chansila')
    expect(site.role).toBe('Full-Stack Software Developer')
    expect(site.startYear).toBe(2020)
    expect(site.heroStack.length).toBeGreaterThan(0)
    expect(site.aboutParagraphs.length).toBeGreaterThan(0)
    expect(site.contactEnabled).toBe(true)
  })

  it('parses JSON list settings', () => {
    const site = resolveSite({
      hero_stack: '["Go","Nuxt"]',
      profiles: '[{"label":"GitHub","href":"https://github.com/x","icon":"simple-icons:github"}]',
    })

    expect(site.heroStack).toEqual(['Go', 'Nuxt'])
    expect(site.profiles[0]?.label).toBe('GitHub')
  })

  it('survives malformed JSON rather than throwing during render', () => {
    // A bad value in one settings row must not blank the whole site.
    const site = resolveSite({ hero_stack: '{not json', profiles: 'null', about_paragraphs: '"a string"' })

    expect(site.heroStack).toEqual([])
    expect(site.profiles).toEqual([])
    expect(site.aboutParagraphs).toEqual([])
  })

  it('treats an empty optional path as unset, not as a path of ""', () => {
    const site = resolveSite({ portrait: '', resume_file: '' })

    expect(site.portrait).toBeNull()
    expect(site.resumeFile).toBeNull()
  })

  it('returns a portrait path when one is set', () => {
    expect(resolveSite({ portrait: '/portrait.jpg' }).portrait).toBe('/portrait.jpg')
  })

  it('only disables the contact form on an explicit "false"', () => {
    expect(resolveSite({}).contactEnabled).toBe(true)
    expect(resolveSite({ contact_enabled: 'true' }).contactEnabled).toBe(true)
    expect(resolveSite({ contact_enabled: '' }).contactEnabled).toBe(true)
    expect(resolveSite({ contact_enabled: 'false' }).contactEnabled).toBe(false)
  })

  it('falls back to 2020 when the start year is missing or unparseable', () => {
    expect(resolveSite({}).startYear).toBe(2020)
    expect(resolveSite({ start_year: 'not a year' }).startYear).toBe(2020)
    expect(resolveSite({ start_year: '2019' }).startYear).toBe(2019)
  })
})

describe('bundled fallback', () => {
  it('carries every section the site renders', () => {
    expect(fallbackContent.projects.length).toBeGreaterThan(0)
    expect(fallbackContent.experience.length).toBeGreaterThan(0)
    expect(fallbackContent.capabilities.length).toBeGreaterThan(0)
    expect(fallbackContent.services.length).toBe(fallbackServices.length)
  })

  it('marks every fallback project as published', () => {
    // The fallback is what a visitor sees during an outage; a draft must not
    // become visible just because the API is down.
    for (const project of fallbackContent.projects) {
      expect(project.published, `${project.slug} is not published`).toBe(true)
    }
  })

  it('declares every setting key the resolver reads', () => {
    for (const key of [
      'name', 'role', 'tagline', 'positioning', 'description', 'start_year',
      'location', 'email', 'portrait_alt', 'hero_stack', 'profiles',
      'about_paragraphs', 'wordpress_seo_stack', 'contact_enabled',
    ]) {
      expect(fallbackContent.settings, `missing ${key}`).toHaveProperty(key)
    }
  })
})
