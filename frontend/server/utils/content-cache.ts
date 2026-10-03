/**
 * Cache invalidation for site content.
 *
 * Content is cached in three places, and an edit has to clear all of them or it
 * will not be visible for minutes:
 *
 *   nitro:handlers:site-content — the /api/content response (60s fresh, then
 *   served stale for up to 10 minutes while it refreshes).
 *
 *   nitro:routes — the rendered HTML and the _payload.json responses held by the
 *   `swr` route rules (5 minutes for the pages, 10 for the résumé and sitemap).
 *
 *   nuxt payload cache — Nuxt's own copy of each route's payload, which is what
 *   client-side navigation reads. Stale entries here mean the page is right on a
 *   hard refresh but wrong when reached by clicking a link.
 *
 * They compound. Without this, saving in the admin panel left the old text on
 * the site for several minutes and several refreshes, because stale-while-
 * revalidate hands the *previous* copy to the request that triggers the refresh
 * — so the first reload after an edit still showed the old page even once the
 * entry had expired.
 *
 * Everything under `nitro:routes` goes, not just one path: a project title
 * appears on the home page, the work index, its own case study and the sitemap,
 * so there is no useful way to narrow it. The cost is that the next request for
 * each page renders instead of being served from cache, which is the point.
 */

/**
 * Keys are removed one by one rather than with `storage.clear(base)`, which
 * looks like it should work and silently does nothing: the cache is served by a
 * filesystem driver mounted above these keys, and its `clear` ignores the base
 * it is given, so every entry survived.
 */
const STALE_PREFIXES = [
  'nitro:routes',
  'nitro:handlers:site-content',
  'nuxt:payload',
  // The same payload cache seen through the parent mount. In dev the payload
  // cache is remounted (see `devStorage` in nuxt.config) to keep Nuxt from
  // writing a file and a directory to one path, so it is listed under both
  // names and both have to go.
  'nuxt-root:payload',
]

/** Exported so the prefixes can be tested without a live cache. */
export function isStaleContentKey(key: string): boolean {
  return STALE_PREFIXES.some((prefix) => key.startsWith(prefix))
}

export async function invalidateContentCache(): Promise<void> {
  const cache = useStorage('cache')

  try {
    const keys = await cache.getKeys()
    const stale = keys.filter(isStaleContentKey)
    await Promise.all(stale.map((key) => cache.removeItem(key, { removeMeta: true })))
  } catch (error) {
    // A cache that cannot be cleared must not fail a save that is already
    // committed to the database. The edit is safe; it is only slow to appear.
    console.error('[content] could not clear the content cache', (error as Error)?.message)
  }
}
