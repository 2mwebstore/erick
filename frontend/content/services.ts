import type { Principle, Service } from '~/types'

/** §17 — outcomes, not technology lists. */
export const services: Service[] = [
  {
    slug: 'web-apps',
    title: 'Custom Web Applications',
    description:
      'Business systems, dashboards, portals, customer applications, and internal tools built around your actual workflow.',
    icon: 'lucide:app-window',
  },
  {
    slug: 'backend',
    title: 'Backend & API Systems',
    description:
      'Secure APIs, business logic, integrations, authentication, and the database design underneath them.',
    icon: 'lucide:server-cog',
  },
  {
    slug: 'mobile',
    title: 'Mobile Applications',
    description:
      'Cross-platform mobile applications using Flutter, sharing one backend with your web product.',
    icon: 'lucide:smartphone',
  },
  {
    slug: 'ecommerce',
    title: 'E-commerce',
    description:
      'Product catalogues, customer accounts, orders, payments, promotions, and the administration system that runs them.',
    icon: 'lucide:shopping-bag',
  },
  {
    slug: 'wordpress-seo',
    title: 'WordPress & SEO',
    description:
      'Professional websites with Elementor, plus technical SEO, performance optimisation, and hosting setup.',
    icon: 'lucide:search',
  },
  {
    slug: 'deployment',
    title: 'Production Deployment',
    description:
      'Docker, Linux, VPS, Nginx, Cloudflare, backups, monitoring, and the operational work after launch.',
    icon: 'lucide:rocket',
  },
]

/** §19 — four principles, kept small on purpose. */
export const principles: Principle[] = [
  {
    slug: 'architecture',
    title: 'Architecture',
    description: 'Designed before implementation, so the structure survives the second and third feature.',
    icon: 'lucide:compass',
  },
  {
    slug: 'security',
    title: 'Security',
    description: 'Considered from the application layer down to the infrastructure it runs on.',
    icon: 'lucide:lock',
  },
  {
    slug: 'reliability',
    title: 'Reliability',
    description: 'Errors, backups, monitoring, and recovery are part of the system, not an afterthought.',
    icon: 'lucide:heart-pulse',
  },
  {
    slug: 'maintainability',
    title: 'Maintainability',
    description: 'Clean structure and documentation so future development does not start with archaeology.',
    icon: 'lucide:wrench',
  },
]

/** §20 — contact form project types. */
export const projectTypes = [
  'Web Application',
  'Mobile Application',
  'E-commerce',
  'Backend / API',
  'WordPress / SEO',
  'DevOps',
  'Other',
] as const
