import { fallbackContent } from '~/content/fallback'
import { DEFAULT_LOCALE } from '~/utils/locales'
import type { SiteContentPayload } from '~/types'

/**
 * Loads site content once per request, before any page renders.
 *
 * A plugin rather than a composable call in each section: content must be present
 * during server rendering so crawlers and no-JavaScript visitors see the real
 * page, and resolving it here means components read it synchronously instead of
 * each opening its own async boundary.
 *
 * The state is serialised into the payload, so client-side navigation reuses it
 * without a second request.
 */
export default defineNuxtPlugin({
  name: 'site-content',

  // Which language to ask for is decided by i18n's route detection, and that
  // happens in a plugin of its own. Without this dependency the content is
  // fetched before the locale is known, and /km renders with English text while
  // the chrome around it is Khmer.
  dependsOn: ['i18n:plugin:route-locale-detect'],

  async setup(nuxtApp) {
    const content = useState<SiteContentPayload>('site-content', () => fallbackContent)
    const degraded = useState<boolean>('site-content-degraded', () => false)
    /**
     * Which language the payload in `content` is actually in.
     *
     * Serialised alongside it, so the client knows what it was handed rather
     * than assuming it matches the URL it is on.
     */
    const loaded = useState<string>('site-content-locale', () => DEFAULT_LOCALE)

    // Read from the route, never from a header: /km/work must serve Khmer
    // whatever the browser's Accept-Language happens to say.
    const i18n = nuxtApp.$i18n as { locale?: { value?: string } } | undefined
    const wanted = () => i18n?.locale?.value ?? DEFAULT_LOCALE

    async function load(locale: string) {
      try {
        content.value = await $fetch<SiteContentPayload>('/api/content', { query: { locale } })
        loaded.value = locale
        degraded.value = false
      } catch (error) {
        // The site stays readable on whatever it already has — bundled content
        // on a first load — rather than failing to render. This is the cost of
        // moving content into the database, and the reason the fallback exists.
        degraded.value = true
        console.warn('[content] API unavailable, serving bundled fallback:', (error as Error)?.message)
      }
    }

    if (import.meta.server) {
      await load(wanted())
      return
    }

    /**
     * On the client, keep the content in step with the language.
     *
     * Switching language is a client-side route change and plugins run once, so
     * without this the chrome turned Khmer while every word that comes from the
     * database — headline, tagline, About, experience — stayed in the language
     * the page was first rendered in. Watching the locale rather than the route
     * also covers a change made by anything other than the switcher.
     *
     * It fires immediately when what arrived does not match what the URL asks
     * for, which is also how a server-side fetch that failed gets repaired: the
     * visitor gets one more attempt instead of a page of bundled English.
     */
    watch(wanted, (locale) => {
      if (locale !== loaded.value || degraded.value) load(locale)
    }, { immediate: true })
  },
})
