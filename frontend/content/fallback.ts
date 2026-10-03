/**
 * Bundled fallback content.
 *
 * Content now lives in MySQL and is edited in /admin. These modules keep two
 * jobs:
 *
 *  1. They seed the database — `node scripts/export-seed.mjs` generates the
 *     seed migration from them.
 *  2. They are the fallback when the API is unreachable. That matters because
 *     moving content into the database put it on the read path for every page:
 *     without this, a database outage would blank the site instead of serving a
 *     slightly stale copy of it.
 *
 * Edits made in /admin do not appear here. After a significant content change,
 * regenerate the seed so the fallback stops drifting from reality.
 */
import type { SiteContentPayload } from '~/types'
import { capabilities, wordpressSeoStack } from './capabilities'
import { aboutParagraphs, timeline } from './experience'
import { projects } from './projects'
import { principles, services } from './services'
import { positioningPillars, siteConfig } from './site'

export const fallbackContent: SiteContentPayload = {
  settings: {
    name: siteConfig.name,
    role: siteConfig.role,
    tagline: siteConfig.tagline,
    headline: siteConfig.headline,
    headline_tail: siteConfig.headlineTail,
    positioning: siteConfig.positioning,
    description: siteConfig.description,
    start_year: String(siteConfig.startYear),
    location: siteConfig.location,
    email: siteConfig.email,
    phone: siteConfig.phone,
    availability: siteConfig.availability,
    portrait: siteConfig.portrait ?? '',
    portrait_alt: siteConfig.portraitAlt,
    resume_file: siteConfig.resumeFile ?? '',
    hero_stack: JSON.stringify(siteConfig.heroStack),
    profiles: JSON.stringify(siteConfig.profiles),
    about_paragraphs: JSON.stringify(aboutParagraphs),
    wordpress_seo_stack: JSON.stringify(wordpressSeoStack),
    contact_enabled: 'true',
  },
  projects: projects.map((p) => ({ ...p, published: true })),
  experience: timeline,
  capabilities,
  services,
  principles,
  pillars: positioningPillars,
}
