import type { Capability } from '~/types'

/**
 * §16 — the ONE section that carries technical depth.
 *
 * Deliberately no percentages, no proficiency bars, no logo wall. Each card is
 * a domain with concrete named technologies underneath it.
 */
export const capabilities: Capability[] = [
  {
    slug: 'architecture',
    title: 'Software Architecture',
    description:
      'System architecture, modular design, maintainable project structures, and scalable application design.',
    icon: 'lucide:layout-template',
    items: ['System design', 'Modular structure', 'Domain separation', 'Code review', 'Documentation'],
  },
  {
    slug: 'backend',
    title: 'Backend & APIs',
    description:
      'Business logic, REST APIs, authentication, authorization, validation, and third-party integrations.',
    icon: 'lucide:server',
    items: ['Go', 'Laravel / PHP', 'REST', 'Auth & sessions', 'Validation', 'Integrations'],
  },
  {
    slug: 'frontend',
    title: 'Frontend',
    description:
      'Responsive, API-driven interfaces with a typed component architecture and measured performance.',
    icon: 'lucide:monitor',
    items: ['Vue 3', 'Nuxt', 'TypeScript', 'Tailwind CSS', 'SSR / SSG', 'Accessibility'],
  },
  {
    slug: 'mobile',
    title: 'Mobile',
    description:
      'Cross-platform mobile applications sharing the same backend contracts as the web clients.',
    icon: 'lucide:smartphone',
    items: ['Flutter', 'Android', 'iOS', 'Firebase', 'Mobile API integration', 'Release builds'],
  },
  {
    slug: 'data',
    title: 'Data',
    description:
      'Relational design that holds up under real data volume, with integrity and recovery planned in.',
    icon: 'lucide:database',
    items: ['MySQL', 'Schema design', 'Migrations', 'Indexing', 'Query optimisation', 'Backup & recovery'],
  },
  {
    slug: 'infrastructure',
    title: 'Infrastructure',
    description:
      'Provisioning, serving, and routing production applications on Linux and managed platforms.',
    icon: 'lucide:cloud',
    items: ['Linux / Ubuntu', 'Docker', 'Nginx', 'VPS', 'Railway', 'DigitalOcean', 'Cloudflare', 'cPanel'],
  },
  {
    slug: 'security',
    title: 'Security',
    description:
      'Application and infrastructure hardening treated as part of the build, not a later pass.',
    icon: 'lucide:shield-check',
    items: [
      'Authentication',
      'Authorization',
      'Input validation',
      'HTTPS / TLS',
      'API protection',
      'Rate limiting',
      'Secure configuration',
    ],
  },
  {
    slug: 'production',
    title: 'Production',
    description:
      'Keeping systems observable and recoverable once real users depend on them.',
    icon: 'lucide:activity',
    items: ['Logging', 'Monitoring', 'Health checks', 'Error handling', 'Backups', 'Deployment', 'Maintenance'],
  },
]

/** §18 — WordPress / SEO capability tags, surfaced under Services and the resume. */
export const wordpressSeoStack: string[] = [
  'WordPress',
  'Elementor',
  'Elementor Pro',
  'PHP',
  'MySQL',
  'Technical SEO',
  'On-page SEO',
  'Performance',
  'Mobile SEO',
  'Sitemap',
  'Structured Data',
  'Cloudflare',
  'cPanel',
]
