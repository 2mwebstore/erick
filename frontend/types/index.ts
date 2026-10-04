/**
 * Content model (§26).
 *
 * Every optional field is optional on purpose: a section renders only when it
 * holds verified information. Nothing in the UI invents a fallback.
 */

export interface CaseStudy {
  overview?: string
  context?: string
  problem?: string
  role?: string
  approach?: string
  architecture?: string
  features?: string[]
  database?: string
  api?: string
  security?: string
  deployment?: string
  challenges?: string[]
  decisions?: TechnicalDecision[]
  outcome?: string
}

export interface TechnicalDecision {
  decision: string
  rationale: string
}

/**
 * Translations carried by the admin payload and sent back on save: locale →
 * field → text. The public payload never includes it — the API has already
 * applied the right language — so it is optional everywhere.
 */
export type Translations = Record<string, Record<string, string>>

export interface Project {
  /** Database id; absent only for the bundled fallback content. */
  id?: number
  slug: string
  title: string
  category: string
  /** One sentence, used on the card and as the case-study meta description. */
  description: string
  /** Longer lede shown at the top of the case study. */
  summary?: string
  technologies: string[]
  image?: string
  imageAlt?: string
  featured: boolean
  /** Drafts stay editable in /admin but are hidden from the public API. */
  published?: boolean
  /** Ordering within Selected Work; lower comes first. */
  order?: number
  year?: string
  translations?: Translations
  liveUrl?: string
  githubUrl?: string
  caseStudy?: CaseStudy
  /** ISO timestamp of the last edit; absent for the bundled fallback content. */
  updatedAt?: string
}

export interface Capability {
  /** Numeric database id; absent for the bundled fallback content. */
  id?: number
  /** Stable key used for anchors and as the CMS slug. */
  slug: string
  title: string
  description: string
  icon: string
  items: string[]
  order?: number
  translations?: Translations
}

export interface Service {
  id?: number
  slug: string
  title: string
  description: string
  icon: string
  order?: number
  translations?: Translations
}

export interface TimelineEntry {
  /** Present once the entry comes from the database. */
  id?: number
  label: string
  title: string
  /** Employer, period and place. Optional: an entry may be a stage, not a job. */
  company?: string
  period?: string
  location?: string
  /** Drives the "Current" badge, rather than parsing the period text. */
  current?: boolean
  technologies?: string[]
  description?: string
  order?: number
  translations?: Translations
}

/**
 * A social or professional profile link.
 *
 * `image` is a URL to a logo. It replaced an Iconify name: the icon sets do not
 * carry every platform, and a URL can point at whatever the profile actually
 * is. `icon` is kept as a fallback so links saved before the change still
 * render something rather than a blank square.
 */
export interface Profile {
  label: string
  href: string
  image?: string
  icon?: string
}

/**
 * One of the numbered positioning statements under the hero (§1). The 01/02/03
 * label is its position on the page, not a stored field.
 */
export interface Pillar {
  id?: number
  slug: string
  title: string
  description: string
  order?: number
  translations?: Translations
}

export interface Principle {
  id?: number
  /** Stable key, used as the CMS slug. */
  slug: string
  title: string
  description: string
  icon: string
  order?: number
  translations?: Translations
}

export interface NavItem {
  /** Key under `nav.` in the locale files; the label is never stored here. */
  key: string
  /** The section's id, appended to whichever locale's home page is current. */
  hash: string
  section: string
}

export type ProjectType =
  | 'Web Application'
  | 'Mobile Application'
  | 'E-commerce'
  | 'Backend / API'
  | 'WordPress / SEO'
  | 'DevOps'
  | 'Other'

export interface ContactPayload {
  name: string
  email: string
  /** Optional: a reply goes to the email address, not the phone. */
  phone?: string
  subject?: string
  projectType: ProjectType | ''
  message: string
  /** Honeypot — must stay empty. */
  company?: string
}

export interface ContactResponse {
  ok: boolean
  message: string
  errors?: Record<string, string>
}

/* ── CMS payload (§28) ─────────────────────────────────────────────────────── */

/** Raw settings as stored: a flat key/value map, values sometimes JSON. */
export type SettingsMap = Record<string, string>

export interface SiteContentPayload {
  settings: SettingsMap
  projects: Project[]
  experience: TimelineEntry[]
  capabilities: Capability[]
  services: Service[]
  principles: Principle[]
  pillars: Pillar[]
  generatedAt?: string
  /** Language this payload was rendered in. */
  locale?: string
  /** Settings have no row id, so their translations arrive separately. */
  settingTranslations?: Translations
}

/** Resolved, typed view of the settings map that components consume. */
export interface ResolvedSite {
  name: string
  role: string
  tagline: string
  /** Hero headline, split so the tail can carry the quieter colour. */
  headline: string
  headlineTail: string
  positioning: string
  description: string
  startYear: number
  location: string
  email: string
  phone: string
  /** One line about availability, shown under the contact details. */
  availability: string
  portrait: string | null
  portraitAlt: string
  resumeFile: string | null
  heroStack: string[]
  profiles: Profile[]
  aboutParagraphs: string[]
  wordpressSeoStack: string[]
  contactEnabled: boolean
}

export interface AdminUser {
  id: number
  email: string
  name: string
  role: 'admin' | 'editor'
  is_active: boolean
  last_login_at?: string
  created_at: string
}

export interface AdminMessage {
  id: number
  name: string
  email: string
  phone?: string
  subject?: string
  project_type: string
  message: string
  created_at: string
  read_at?: string
  archived_at?: string
  ip_address?: string
  user_agent?: string
}

export interface AuditEntry {
  id: number
  actor_email: string
  action: string
  entity: string
  entity_id?: string
  detail?: unknown
  ip_address?: string
  created_at: string
}

/** Whether image fields can upload to R2, and what they accept. */
export interface UploadSettings {
  enabled: boolean
  maxBytes?: number
  types?: string[]
}

/** One finding of the SEO check (backend/internal/seoaudit). */
export interface SeoCheck {
  id: string
  group: 'sitelinks' | 'technical' | 'onpage'
  title: string
  status: 'pass' | 'warn' | 'fail'
  detail: string
  urls?: string[]
}

/** What the SEO check's crawl learned about one URL. */
export interface SeoPage {
  url: string
  status: number
  error?: string
  redirectTo?: string
  /** Clicks from the home page; -1 when no link leads here. */
  depth: number
  inSitemap: boolean
  indexable: boolean
  title: string
  description: string
  canonical: string
  h1: number
  imagesWithoutAlt: number
  structuredData: string[]
  hreflang?: Record<string, string>
  incoming: number
  outgoing: number
}

export interface SeoReport {
  origin: string
  startedAt: string
  durationMs: number
  /** This site's own diagnostic, 0–100 — not anything Google reports. */
  score: number
  sitelinksScore: number
  truncated: boolean
  checks: SeoCheck[]
  pages: SeoPage[]
}

export interface SeoAuditStatus {
  configured: boolean
  origin?: string
  running: boolean
  error?: string
  report?: SeoReport
}
