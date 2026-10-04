import type { Project } from '~/types'
import { LOCALE_LANGUAGE, localizedPath, resolveLocale, SUPPORTED_LOCALES } from '~/utils/locales'

function origin(): string {
  return useRuntimeConfig().public.siteUrl.replace(/\/+$/, '')
}

function currentLocale() {
  return resolveLocale(useI18n().locale.value)
}

/**
 * Pages pass their English path; this returns it in the language being
 * rendered. Without it every /km page declared its English twin as canonical,
 * which tells a search engine the Khmer page is a duplicate to drop — the
 * opposite of what its hreflang alternates say.
 */
function pagePath(path?: string): string {
  return path === undefined ? useRoute().path : localizedPath(path, currentLocale())
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
  const canonical = absoluteUrl(pagePath(input.path), base)
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
  const sameAs = sameAsUrls(site.value.profiles.map((p) => p.href), base)
  if (sameAs.length) person.sameAs = sameAs

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
        inLanguage: SUPPORTED_LOCALES.map((code) => LOCALE_LANGUAGE[code]),
        publisher: { '@id': `${base}/#person` },
      },
    ],
  })
}

/** WebPage node for a specific route (§22). */
export function useWebPageSchema(input: { name: string; description: string; path?: string }) {
  const base = origin()
  const url = absoluteUrl(pagePath(input.path), base)

  jsonLd('ld-webpage', {
    '@context': 'https://schema.org',
    '@type': 'WebPage',
    '@id': `${url}#webpage`,
    url,
    name: input.name,
    description: toPlainText(input.description),
    isPartOf: { '@id': `${base}/#website` },
    about: { '@id': `${base}/#person` },
    inLanguage: LOCALE_LANGUAGE[currentLocale()],
  })
}

/** CreativeWork node for a project case study (§22). */
export function useProjectSchema(project: Project) {
  const { t } = useI18n()
  const base = origin()
  const locale = currentLocale()
  const link = (path: string) => absoluteUrl(localizedPath(path, locale), base)
  const url = link(`/work/${project.slug}`)
  const home = t('nav.home')

  const work: Record<string, unknown> = {
    '@type': 'CreativeWork',
    '@id': `${url}#project`,
    name: project.title,
    url,
    genre: project.category,
    creator: { '@id': `${base}/#person` },
    inLanguage: LOCALE_LANGUAGE[locale],
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
          // The nav label is lower-case ("home") in English; a breadcrumb is a title.
          { '@type': 'ListItem', position: 1, name: home.charAt(0).toUpperCase() + home.slice(1), item: link('/') },
          { '@type': 'ListItem', position: 2, name: t('nav.work'), item: link('/work') },
          { '@type': 'ListItem', position: 3, name: project.title, item: url },
        ],
      },
    ],
  })
}
