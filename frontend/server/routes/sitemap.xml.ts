import { fallbackContent } from '~/content/fallback'
import type { SiteContentPayload } from '~/types'

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
 */
const LOCALES = [
  { code: 'en', hreflang: 'en-US', prefix: '' },
  { code: 'km', hreflang: 'km-KH', prefix: '/km' },
]

export default defineEventHandler(async (event) => {
  const { public: pub, apiBaseUrl, apiTimeoutMs } = useRuntimeConfig(event)
  const origin = pub.siteUrl.replace(/\/+$/, '')
  const lastmod = new Date().toISOString().slice(0, 10)

  let content: SiteContentPayload = fallbackContent
  try {
    content = await $fetch<SiteContentPayload>('/v1/content', {
      baseURL: apiBaseUrl,
      timeout: Number(apiTimeoutMs) || 8000,
    })
  } catch (error) {
    console.warn('[sitemap] content API unavailable, using bundled slugs:', (error as Error)?.message)
  }

  const routes = [
    { path: '/', priority: '1.0', changefreq: 'monthly' },
    { path: '/work', priority: '0.9', changefreq: 'monthly' },
    { path: '/resume', priority: '0.7', changefreq: 'yearly' },
    ...content.projects.map((p) => ({
      path: `/work/${p.slug}`,
      priority: '0.8',
      changefreq: 'yearly',
    })),
  ]

  // The home page keeps its trailing slash so the sitemap and the canonical tag
  // are byte-identical; nothing else has one.
  const href = (prefix: string, path: string) =>
    `${origin}${prefix}${path === '/' && prefix ? '' : path}`

  const urls = routes
    .flatMap((route) =>
      LOCALES.map((locale) => {
        const alternates = LOCALES.map(
          (alt) =>
            `    <xhtml:link rel="alternate" hreflang="${alt.hreflang}" href="${href(alt.prefix, route.path)}"/>`,
        )
          .concat(
            `    <xhtml:link rel="alternate" hreflang="x-default" href="${href('', route.path)}"/>`,
          )
          .join('\n')

        return `  <url>
    <loc>${href(locale.prefix, route.path)}</loc>
${alternates}
    <lastmod>${lastmod}</lastmod>
    <changefreq>${route.changefreq}</changefreq>
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
