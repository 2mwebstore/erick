import type { TimelineEntry } from '~/types'
import { siteConfig } from './site'

/**
 * §12 — a capability progression, not an employment history.
 *
 * Deliberately contains no employers, job titles, or dates beyond the year
 * range the developer confirmed. Add roles here only when supplied.
 */
export const timeline: TimelineEntry[] = [
  {
    label: String(siteConfig.startYear),
    title: 'Started full-time software development',
  },
  {
    label: 'Backend',
    title: 'Backend & API development',
    description: 'Business logic, REST APIs, authentication, validation, and third-party integrations.',
  },
  {
    label: 'Frontend',
    title: 'Frontend & web applications',
    description: 'API-driven interfaces, admin dashboards, and customer-facing web applications.',
  },
  {
    label: 'Mobile',
    title: 'Mobile applications',
    description: 'Cross-platform Flutter applications consuming the same backend services.',
  },
  {
    label: 'Data',
    title: 'Database & infrastructure',
    description: 'Relational schema design, migrations, query performance, Linux servers, and Nginx.',
  },
  {
    label: 'Production',
    title: 'Cloud, DevOps & production operations',
    description: 'Docker, Cloudflare, deployment pipelines, monitoring, backups, and ongoing maintenance.',
  },
  {
    label: '2026',
    title: siteConfig.role,
    description: 'Owning systems end to end, from requirements through production maintenance.',
  },
]

/**
 * Short About prose (§12). Two paragraphs maximum — the evidence lives in
 * Selected Work, not here.
 */
export const aboutParagraphs: string[] = [
  `I am a full-stack software developer with professional development experience since ${siteConfig.startYear}. I work across backend systems, frontend applications, mobile development, databases, APIs, infrastructure, deployment, and production operations.`,
  'Most of my work has been building and running complete systems rather than isolated features: designing the data model, writing the API, building the interface that consumes it, then deploying and maintaining the result. That shapes how I make decisions — I favour structures that stay readable a year later, and I plan for the failure modes before they reach production.',
]
