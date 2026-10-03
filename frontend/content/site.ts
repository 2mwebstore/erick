import type { Pillar, Profile } from '~/types'
/**
 * Site-wide identity and positioning.
 *
 * ── Editing rules (§27) ──────────────────────────────────────────────────────
 * Nothing here may be invented. A string beginning with `TODO:` is treated as
 * an unfilled placeholder: the UI renders it as a visible "content pending"
 * note instead of prose, and `npm run content:check` lists it. Delete the
 * `TODO:` prefix once you have replaced the text with something true.
 * Leave a field out entirely (or as an empty string) to hide its UI.
 */

export const siteConfig = {
  name: 'Kong Chansila',
  role: 'Full-Stack Software Developer',

  /** §1 supporting statement. */
  tagline:
    'Designing, building, and operating production-ready web, mobile, and business systems since 2020.',

  /**
   * Hero headline, in two halves: the first is set in the foreground colour,
   * the second in the quieter one. Split as data rather than markup so both
   * halves can be edited and translated from the admin panel.
   */
  headline: 'Building production-ready software',
  headlineTail: 'from architecture to deployment.',

  /** §1 professional positioning, used in About and the resume profile. */
  positioning:
    'I build complete software systems across backend, frontend, mobile, databases, APIs, infrastructure, security, deployment, and production operations.',

  /** Used for meta description on the homepage. */
  description:
    'Full-Stack Software Developer building production-ready web, mobile, and business systems — backend, frontend, mobile, databases, APIs, infrastructure, and deployment. Full-time development since 2020.',

  startYear: 2020,

  /** Location shown in the footer and resume. Remove if you would rather not publish it. */
  location: 'Cambodia',

  /** Shown on the contact card. Empty hides the row rather than showing a gap. */
  phone: '',

  /** One line under the contact details, e.g. current availability. */
  availability: '',

  /** Contact address the form and footer use. Change this if you want a dedicated address. */
  email: 'membertsd@gmail.com',

  /**
   * Only add a profile here once the URL is real — a dead link costs more
   * credibility than an absent one. `label` is used as the accessible name.
   */
  /**
   * Profile links. `image` is a logo URL — either a file in public/ or a remote
   * one. Add a profile only once its URL resolves (§27).
   */
  profiles: [
    // { label: 'GitHub', href: 'https://github.com/…', image: '/logos/github.svg' },
    // { label: 'LinkedIn', href: 'https://linkedin.com/in/…', image: '/logos/linkedin.svg' },
    // { label: 'Telegram', href: 'https://t.me/…', image: '/logos/telegram.svg' },
  ] as Profile[],

  /**
   * Hero portrait. Put the file in frontend/public/ and set the path here, e.g.
   * '/portrait.jpg'. While this is null the hero shows an editable frame rather
   * than a stock photo. Use a square or 4:5 crop, at least 800px wide.
   */
  portrait: null as string | null,

  /** Alt text for the portrait. Describe the person, not the file. */
  portraitAlt: 'Kong Chansila, Full-Stack Software Developer',

  /** Compact hero technology line (§10). Keep it short — this is not the full list. */
  heroStack: ['Go', 'Laravel', 'Vue', 'Nuxt', 'Flutter', 'MySQL', 'Docker', 'Cloudflare'],

  /** Resume PDF placed in frontend/public/. Set to null until the file exists. */
  resumeFile: null as string | null,
} as const

/** Three things the homepage must communicate immediately (§1). */
export const positioningPillars: Pillar[] = [
  {
    slug: 'experience',
    title: 'Experience',
    description: `Professional full-time software development since ${siteConfig.startYear}.`,
  },
  {
    slug: 'technical-breadth',
    title: 'Technical breadth',
    description:
      'Backend, frontend, mobile, databases, APIs, infrastructure, WordPress, SEO, and cloud deployment.',
  },
  {
    slug: 'production-ownership',
    title: 'Production ownership',
    description:
      'Architecture, development, database, security, deployment, monitoring, backup, and maintenance — not just writing code.',
  },
]

/** The lifecycle chain shown in the hero visual and About (§1). */
export const lifecycle = [
  'Architecture',
  'Development',
  'Database',
  'Security',
  'Deployment',
  'Monitoring',
  'Backup',
  'Maintenance',
]
