import type { Project } from '~/types'

/**
 * §13 / §15 / §26 — Selected Work.
 *
 * ── How to fill this in ──────────────────────────────────────────────────────
 * Any string starting with `TODO:` is an unfilled placeholder. The UI renders
 * it as a visible "content pending" note rather than prose, and
 * `npm run content:check` reports every one. Replace the whole string with real
 * text and drop the prefix — or delete the field, and its section disappears
 * from the case study.
 *
 * Do not add a `liveUrl` or `githubUrl` until the URL resolves. Do not list a
 * feature that is not implemented. An absent section is always better than an
 * invented one.
 */
export const projects: Project[] = [
  {
    slug: 'bubble-white',
    title: 'Bubble White',
    category: 'E-commerce / Web Application',
    description:
      'A Cambodia-focused clothing e-commerce platform covering product discovery, shopping workflows, customer accounts, orders, promotions, image management, and payment integration.',
    summary:
      'A Cambodia-focused clothing e-commerce platform. The system covers the full commercial path — browsing and product discovery, cart and checkout workflows, customer accounts, order handling, promotions, product image management, and payment integration — together with the administration needed to operate it day to day.',
    technologies: ['Nuxt', 'Vue', 'Tailwind CSS', 'MySQL', 'Cloudflare', 'Payment APIs'],
    featured: true,
    order: 1,
    // liveUrl: 'TODO: add only once the production URL is live',
    caseStudy: {
      overview:
        'An online clothing store for the Cambodian market, built as a complete commerce system rather than a catalogue site: customers browse and buy, and the business operates orders, stock imagery, and promotions from the same application.',
      context: 'TODO: who the store serves, how they sold before this existed, and what scale it operates at.',
      problem: 'TODO: the specific problem the platform was built to solve.',
      role: 'TODO: which parts you owned — data model, API, storefront, admin, deployment.',
      approach: 'TODO: how you sequenced the build and why.',
      architecture: 'TODO: request path, rendering strategy, and where each responsibility lives.',
      // Features below are confirmed scope; add to this list only what is shipped.
      features: [
        'Product discovery and browsing',
        'Shopping and checkout workflow',
        'Customer accounts',
        'Order management',
        'Promotions',
        'Product image management',
        'Payment integration',
      ],
      database: 'TODO: core tables, key relationships, and the indexes that matter.',
      api: 'TODO: endpoint surface, auth model, and validation strategy.',
      security: 'TODO: authentication, authorization, payment handling, and hardening.',
      deployment: 'TODO: hosting, TLS, Cloudflare configuration, and release process.',
      challenges: [
        'TODO: a real constraint you hit, and how you resolved it.',
      ],
      decisions: [
        { decision: 'TODO: a technical decision you made', rationale: 'TODO: why, and what you traded away' },
      ],
      outcome: 'TODO: what the platform does for the business now. No invented metrics.',
    },
  },

  {
    slug: 'coffee-app',
    title: 'Coffee App',
    category: 'Mobile / Business Application',
    description: 'TODO: one sentence describing what this application does, in business terms.',
    technologies: ['Flutter'],
    featured: true,
    order: 2,
    caseStudy: {
      overview: 'TODO: what the app is and who uses it.',
      role: 'TODO: which parts you built.',
      // features: ['TODO: only confirmed, shipped functionality'],
      // architecture: 'TODO: app structure, state management, backend contract.',
      // api: 'TODO: the services it talks to.',
      // deployment: 'TODO: distribution — store release, internal build, or both.',
    },
  },

  {
    slug: 'check-in-app',
    title: 'Check-in App',
    category: 'Application',
    description: 'TODO: one sentence describing what gets checked in, by whom, and why.',
    technologies: ['Flutter'],
    featured: true,
    order: 3,
    caseStudy: {
      overview: 'TODO: what the app is and who uses it.',
      role: 'TODO: which parts you built.',
      // features: ['TODO: only confirmed, shipped functionality'],
      // database: 'TODO: what is stored and how.',
      // security: 'TODO: how identity and records are protected.',
    },
  },

  {
    slug: 'buffet-system',
    title: 'Buffet System',
    category: 'Restaurant / POS System',
    description: 'TODO: one sentence on what the system does for a buffet restaurant — ordering, tables, billing, or all three.',
    // Left empty deliberately: I do not know what this was built with, and
    // guessing a stack would be a fabricated claim (§27). Add the real ones and
    // the tag row and Technology section appear automatically.
    technologies: [],
    featured: true,
    order: 4,
    caseStudy: {
      overview: 'TODO: what the system is, who operates it, and in what kind of venue.',
      problem: 'TODO: what the restaurant was doing before, and what broke at their scale.',
      role: 'TODO: which parts you built.',
      // approach: 'TODO: how you sequenced it.',
      // architecture: 'TODO: devices, server, and how orders reach the kitchen.',
      // features: [
      //   'TODO: e.g. table and session management',
      //   'TODO: e.g. order taking on a handheld device',
      //   'TODO: e.g. kitchen ticket printing',
      //   'TODO: e.g. per-head buffet pricing and time limits',
      //   'TODO: e.g. billing and daily sales reporting',
      // ],
      // database: 'TODO: what is stored — sessions, orders, menu, pricing.',
      // api: 'TODO: the service the devices talk to.',
      // security: 'TODO: staff roles and what each can do.',
      // deployment: 'TODO: on-premise server, cloud, or both.',
      // challenges: ['TODO: e.g. keeping ordering working when the venue wifi drops'],
      // outcome: 'TODO: what it changed for the restaurant. No invented numbers.',
    },
  },

  {
    slug: 'portfolio',
    title: 'Portfolio',
    category: 'Web Development',
    description:
      'This site — a server-rendered Nuxt 4 frontend with a Go API and MySQL behind it, containerised and deployed behind Nginx and Cloudflare.',
    summary:
      'This site is the smallest honest example of the stack I work in: a server-rendered Nuxt 4 frontend, a Go API handling the one route that genuinely needs a server, MySQL for the data that must persist, and a container setup that runs the same way locally and in production.',
    technologies: ['Nuxt 4', 'Vue 3', 'TypeScript', 'Tailwind CSS', 'Go', 'MySQL', 'Docker', 'Nginx', 'Cloudflare'],
    featured: true,
    order: 5,
    caseStudy: {
      overview:
        'A portfolio built as a production system rather than a static page, so the engineering choices on display are the ones actually running the site.',
      problem:
        'A portfolio has almost no inherent backend requirement. The interesting constraint is deciding where complexity is justified: adding a database and an API service for their own sake would contradict the judgement the site is meant to demonstrate.',
      role: 'Sole developer — design system, frontend, API, schema, containers, deployment documentation.',
      approach:
        'Content lives in typed TypeScript modules and is prerendered, so pages ship as static HTML with no runtime data dependency. Only the contact form needs server behaviour, so that is the only thing with an API and a table behind it.',
      architecture:
        'Cloudflare terminates TLS and fronts the origin. Nginx reverse-proxies to the Nuxt server, which prerenders every content route. The browser posts the contact form to a Nuxt server route, which forwards it server-side to the Go API — the API is never reachable from the public internet, so it needs no CORS surface. The Go service validates, rate-limits, and persists to MySQL.',
      features: [
        'Prerendered content routes with typed content modules',
        'Dark, light, and system theme with persisted preference',
        'Per-project case-study pages that render only verified sections',
        'Contact API with validation, per-IP rate limiting, and honeypot spam protection',
        'Person, WebSite, and CreativeWork structured data',
        'Generated sitemap and robots.txt',
        'Multi-stage Docker builds for both services with health checks',
      ],
      database:
        'One table, contact_messages, holding the submission plus request metadata for abuse investigation. Migrations are plain, numbered, forward-and-back SQL files — no ORM, because there is no schema complexity to justify one.',
      api: 'Go net/http with a small middleware chain: request ID, structured logging, panic recovery, body-size limit, security headers, and a per-IP token-bucket limiter. Two routes: POST /v1/contact and GET /health.',
      security:
        'The API binds to the container network only. Input is validated and length-bounded server-side before it reaches the database, statements are parameterised, error responses never echo internals, and secrets come from the environment rather than the image.',
      deployment:
        'docker compose builds both services against a MySQL container with a persistent volume and health-gated startup ordering. The same compose file plus an env file is what runs in production behind Nginx.',
      decisions: [
        {
          decision: 'Prerender content routes instead of serving them from the database',
          rationale:
            'The content changes when I change it, not per request. Prerendering removes the database from the critical path for every page a visitor actually reads, so the site stays up and fast even if the API is down.',
        },
        {
          decision: 'Proxy the contact form through the Nuxt server rather than calling Go from the browser',
          rationale:
            'It keeps the Go service unreachable from the internet and removes the need for a public CORS policy. The cost is one extra internal hop on a route that runs a few times a day.',
        },
        {
          decision: 'No CMS or admin panel',
          rationale:
            'One author editing typed files gets type safety and version history for free. An admin panel would add authentication, sessions, and an attack surface to solve a problem that does not exist yet.',
        },
      ],
      outcome:
        'The site is fully prerendered, keyboard accessible, and works with JavaScript-driven animation disabled. Build, typecheck, lint, and the API test suite all run clean.',
    },
  },

  {
    slug: 'wordpress-seo',
    title: 'WordPress & SEO',
    category: 'Web / SEO',
    description: 'TODO: one sentence on the kind of WordPress and SEO work you deliver, without naming clients you cannot cite.',
    technologies: [
      'WordPress',
      'Elementor Pro',
      'PHP',
      'MySQL',
      'Technical SEO',
      'Structured Data',
      'Cloudflare',
      'cPanel',
    ],
    featured: true,
    order: 6,
    caseStudy: {
      overview: 'TODO: the scope of this work — build, migration, optimisation, or ongoing maintenance.',
      role: 'TODO: what you were responsible for.',
      approach: 'TODO: your technical SEO process — audit, fixes, measurement.',
      // features: [
      //   'TODO: e.g. Core Web Vitals remediation',
      //   'TODO: e.g. structured data implementation',
      //   'TODO: e.g. sitemap and indexation cleanup',
      // ],
      deployment: 'TODO: hosting, caching, and CDN configuration you set up.',
      // outcome: 'TODO: real, citable results only — no invented traffic numbers.',
    },
  },
]

/** Featured projects for the homepage, in author-defined order. */
export const featuredProjects = projects
  .filter((p) => p.featured)
  .sort((a, b) => (a.order ?? 99) - (b.order ?? 99))

export const allProjects = [...projects].sort((a, b) => (a.order ?? 99) - (b.order ?? 99))

export function getProject(slug: string): Project | undefined {
  return projects.find((p) => p.slug === slug)
}
