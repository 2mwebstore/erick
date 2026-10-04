import { fallbackContent } from '~/content/fallback'
import type { SiteContentPayload } from '~/types'
import { absoluteUrl } from '~/utils/format'
import { DEFAULT_LOCALE, LOCALE_LANGUAGE, localizedPath, SUPPORTED_LOCALES } from '~/utils/locales'

/**
 * Sitemap generated from the CMS (§22).
 *
 * Project slugs now live in the database, so this asks the content API for them.
 * On failure it falls back to the bundled content rather than emitting a sitemap
 * with no project pages — an empty sitemap actively tells crawlers to forget
 * pages that still exist.
 *
 * Every page is listed once per language, and each entry carries xhtml:link
 * alternates naming the other. Listing the Khmer URLs without those alternates
 * would look like duplicate content rather than a translation.
 *
 * lastmod is only given where the content says when it last changed: a project
 * page carries its own edit date, and /work the newest of them. The home and
 * résumé pages have no single edit date, so they carry none. Stamping every
 * entry with the day the sitemap was generated, as this once did, teaches a
 * search engine to ignore lastmod altogether.
 */

/** YYYY-MM-DD from an ISO timestamp, or undefined when there is no usable date. */
function day(timestamp?: string): string | undefined {
  if (!timestamp) return undefined
  const date = new Date(timestamp)
  // Go's zero time (year 1) is what an unset timestamp serialises as.
  return Number.isNaN(date.getTime()) || date.getUTCFullYear() < 2000
    ? undefined
    : date.toISOString().slice(0, 10)
}

export default defineEventHandler(async (event) => {
  const { public: pub, apiBaseUrl, apiTimeoutMs } = useRuntimeConfig(event)
  const origin = pub.siteUrl.replace(/\/+$/, '')

  let content: SiteContentPayload = fallbackContent
  try {
    content = await $fetch<SiteContentPayload>('/v1/content', {
      baseURL: apiBaseUrl,
      timeout: Number(apiTimeoutMs) || 8000,
    })
  } catch (error) {
    console.warn('[sitemap] content API unavailable, using bundled slugs:', (error as Error)?.message)
  }

  const projectDays = content.projects.map((p) => day(p.updatedAt)).filter((d): d is string => !!d)
  const newestProject = projectDays.length ? projectDays.sort().at(-1) : undefined

  const routes: { path: string; priority: string; changefreq: string; lastmod?: string }[] = [
    { path: '/', priority: '1.0', changefreq: 'monthly' },
    { path: '/work', priority: '0.9', changefreq: 'monthly', lastmod: newestProject },
    { path: '/resume', priority: '0.7', changefreq: 'yearly' },
    ...content.projects.map((p) => ({
      path: `/work/${p.slug}`,
      priority: '0.8',
      changefreq: 'yearly',
      lastmod: day(p.updatedAt),
    })),
  ]

  // Same paths as the canonical tags, which build them through localizedPath
  // too. absoluteUrl keeps the home page's trailing slash, as the canonical does.
  const href = (locale: string, path: string) => absoluteUrl(localizedPath(path, locale), origin)

  const urls = routes
    .flatMap((route) =>
      SUPPORTED_LOCALES.map((locale) => {
        const alternates = SUPPORTED_LOCALES.map(
          (alt) =>
            `    <xhtml:link rel="alternate" hreflang="${LOCALE_LANGUAGE[alt]}" href="${href(alt, route.path)}"/>`,
        )
          .concat(
            `    <xhtml:link rel="alternate" hreflang="x-default" href="${href(DEFAULT_LOCALE, route.path)}"/>`,
          )
          .join('\n')

        return `  <url>
    <loc>${href(locale, route.path)}</loc>
${alternates}
${route.lastmod ? `    <lastmod>${route.lastmod}</lastmod>\n` : ''}    <changefreq>${route.changefreq}</changefreq>
    <priority>${route.priority}</priority>
  </url>`
      }),
    )
    .join('\n')

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  setHeader(event, 'cache-control', 'public, max-age=3600')

  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls}
</urlset>
`
})
