import { resolveLocale } from '~/utils/locales'
import type { SiteContentPayload } from '~/types'

/**
 * Public content proxy.
 *
 * The Go API is not reachable from the browser, so the page fetches content from
 * this origin and the server forwards it. The upstream response is cached here as
 * well as in Nitro's route rules: two cheap layers mean a content read almost
 * never reaches MySQL.
 */
export default defineCachedEventHandler(
  async (event): Promise<SiteContentPayload> => {
    const config = useRuntimeConfig(event)

    return await $fetch<SiteContentPayload>('/v1/content', {
      baseURL: config.apiBaseUrl,
      query: { locale: readLocale(event) },
      timeout: Number(config.apiTimeoutMs) || 8000,
    })
  },
  {
    name: 'site-content',
    maxAge: 60,
    // Serve the previous copy while refreshing, so a slow or failing API shows
    // stale content rather than an error.
    staleMaxAge: 600,
    swr: true,
    // The locale is part of the key. Without it the first language requested
    // after a cache miss would be served to everyone until it expired.
    getKey: (event) => `site-content-${readLocale(event)}`,
  },
)

/**
 * An unknown locale is answered in English rather than rejected: a bad value in
 * a query string should not be able to turn the site into an error page.
 */
function readLocale(event: Parameters<typeof getQuery>[0]): string {
  return resolveLocale(getQuery(event).locale)
}
