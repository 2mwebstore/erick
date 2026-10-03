import type { Project } from '~/types'

function origin(): string {
  return useRuntimeConfig().public.siteUrl.replace(/\/+$/, '')
}

/**
 * Absolute canonical URL for the current route.
 *
 * Resolved eagerly during setup rather than inside a computed: unhead evaluates
 * head values after the component has been rendered, at which point the Nuxt
 * instance — and therefore useRuntimeConfig — is no longer available. This bit
 * us during prerender, so the value stays a plain string.
 */
export function useCanonical(path?: string): string {
  return absoluteUrl(path ?? useRoute().path, origin())
}

/**
 * Per-page metadata: title, description, canonical, Open Graph, Twitter (§22).
 * Called once per page; the title template lives in app.vue.
 */
export function usePageSeo(input: {
  title: string
  description: string
  path?: string
  /** Absolute or root-relative image path. */
  image?: string
  type?: 'website' | 'article' | 'profile'
  noindex?: boolean
}) {
  const { site } = useSiteContent()
  const base = origin()
  const canonical = absoluteUrl(input.path ?? useRoute().path, base)
  const image = absoluteUrl(input.image ?? '/og-default.png', base)

  useHead({
    link: [{ rel: 'canonical', href: canonical }],
    meta: input.noindex ? [{ name: 'robots', content: 'noindex, follow' }] : [],
  })

  useSeoMeta({
    title: input.title,
    description: toPlainText(input.description),
    ogTitle: `${input.title} — ${site.value.name}`,
    ogDescription: input.description,
    ogUrl: canonical,
    ogImage: image,
    ogType: (input.type ?? 'website') as 'website',
    twitterTitle: `${input.title} — ${site.value.name}`,
    twitterDescription: toPlainText(input.description),
    twitterImage: image,
    ogImageWidth: 1200,
    ogImageHeight: 630,
    ogImageType: 'image/png',
    ogImageAlt: `${site.value.name} — ${site.value.role}`,
  })

  return { canonical }
}

function jsonLd(id: string, data: Record<string, unknown>) {
  useHead({
    script: [
      {
        key: id,
        type: 'application/ld+json',
        innerHTML: JSON.stringify(data),
      },
    ],
  })
}

/** Person + WebSite graph, emitted once site-wide (§22). */
export function usePersonSchema() {
  const { site } = useSiteContent()
  const base = origin()

  const person: Record<string, unknown> = {
    '@type': 'Person',
    '@id': `${base}/#person`,
    name: site.value.name,
    jobTitle: site.value.role,
    description: site.value.positioning,
    url: base,
    email: `mailto:${site.value.email}`,
    knowsAbout: [
      'Software architecture',
      'Backend development',
      'Frontend development',
      'Mobile development',
      'Databases',
      'API design',
      'DevOps',
      'Technical SEO',
    ],
  }

  if (site.value.location) {
    person.address = { '@type': 'PostalAddress', addressCountry: site.value.location }
  }
  if (site.value.profiles.length) person.sameAs = site.value.profiles.map((p) => p.href)

  jsonLd('ld-identity', {
    '@context': 'https://schema.org',
    '@graph': [
      person,
      {
        '@type': 'WebSite',
        '@id': `${base}/#website`,
        url: base,
        name: `${site.value.name} — ${site.value.role}`,
        description: toPlainText(site.value.description),
        inLanguage: 'en',
        publisher: { '@id': `${base}/#person` },
      },
    ],
  })
}

/** WebPage node for a specific route (§22). */
export function useWebPageSchema(input: { name: string; description: string; path?: string }) {
  const base = origin()
  const url = absoluteUrl(input.path ?? useRoute().path, base)

  jsonLd('ld-webpage', {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    '@id': `${url}#webpage`,
    url,
    name: input.name,
    description: toPlainText(input.description),
    isPartOf: { '@id': `${base}/#website` },
    about: { '@id': `${base}/#person` },
    inLanguage: 'en',
  })
}

/** CreativeWork node for a project case study (§22). */
export function useProjectSchema(project: Project) {
  const base = origin()
  const url = absoluteUrl(`/work/${project.slug}`, base)

  const work: Record<string, unknown> = {
    '@type': 'CreativeWork',
    '@id': `${url}#project`,
    name: project.title,
    url,
    genre: project.category,
    creator: { '@id': `${base}/#person` },
    inLanguage: 'en',
  }

  if (hasContent(project.description)) work.description = toPlainText(project.description)
  if (project.technologies.length) work.keywords = project.technologies.join(', ')
  if (project.liveUrl) work.sameAs = project.liveUrl

  jsonLd('ld-project', {
    '@context': 'https://schema.org',
    '@graph': [
      work,
      {
        '@type': 'BreadcrumbList',
        itemListElement: [
          { '@type': 'ListItem', position: 1, name: 'Home', item: base },
          { '@type': 'ListItem', position: 2, name: 'Work', item: `${base}/work` },
          { '@type': 'ListItem', position: 3, name: project.title, item: url },
        ],
      },
    ],
  })
}
