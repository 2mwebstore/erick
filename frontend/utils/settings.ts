import type { Profile, ResolvedSite, SettingsMap } from '~/types'

/**
 * Turns the flat settings map from the database into a typed object.
 *
 * Settings are stored as strings because the table is key/value, which keeps new
 * settings migration-free. The cost is that lists arrive as JSON text, so every
 * read goes through here — and a malformed value falls back to a sane default
 * rather than throwing during render.
 */
function parseList<T>(raw: string | undefined, fallback: T[]): T[] {
  if (!raw) return fallback
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? (parsed as T[]) : fallback
  } catch {
    return fallback
  }
}

function text(settings: SettingsMap, key: string, fallback = ''): string {
  const value = settings[key]
  return value === undefined || value === '' ? fallback : value
}

/** Empty string means "unset" for an optional path, not a path of "". */
function optional(settings: SettingsMap, key: string): string | null {
  const value = settings[key]
  return value ? value : null
}

export function resolveSite(settings: SettingsMap): ResolvedSite {
  const startYear = Number.parseInt(text(settings, 'start_year', '2020'), 10)

  return {
    name: text(settings, 'name', 'Kong Chansila'),
    role: text(settings, 'role', 'Full-Stack Software Developer'),
    tagline: text(settings, 'tagline'),
    headline: text(settings, 'headline', 'Building production-ready software'),
    // Empty means "no quieter half", which is what the Settings hint promises;
    // running it through `text` gave the default back instead, so clearing the
    // field silently reinstated a line the editor had just deleted. Only an
    // absent key takes the default, which is what the bundled fallback needs.
    headlineTail: settings.headline_tail ?? 'from architecture to deployment.',
    positioning: text(settings, 'positioning'),
    description: text(settings, 'description'),
    startYear: Number.isFinite(startYear) ? startYear : 2020,
    location: text(settings, 'location'),
    email: text(settings, 'email'),
    phone: text(settings, 'phone'),
    availability: text(settings, 'availability'),
    portrait: optional(settings, 'portrait'),
    portraitAlt: text(settings, 'portrait_alt'),
    resumeFile: optional(settings, 'resume_file'),
    heroStack: parseList<string>(settings.hero_stack, []),
    profiles: parseList<Profile>(settings.profiles, []),
    aboutParagraphs: parseList<string>(settings.about_paragraphs, []),
    wordpressSeoStack: parseList<string>(settings.wordpress_seo_stack, []),
    // Anything other than an explicit "false" leaves the form enabled.
    contactEnabled: text(settings, 'contact_enabled', 'true') !== 'false',
  }
}
