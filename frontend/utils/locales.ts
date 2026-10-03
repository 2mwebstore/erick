/**
 * The languages the site serves, in one place.
 *
 * nuxt.config.ts declares them to the i18n module, but the content pipeline
 * needs them too — the plugin asks the API for one, the proxy validates the
 * query string, and the two have to agree on which language is the source text.
 * Kept as a plain module rather than read from the module's runtime config so
 * server code can import it without pulling in the app.
 */

/** The language content is authored in. Everything else is an overlay of it. */
export const DEFAULT_LOCALE = 'en'

/** Closed set: an unrecognised code is answered in English, never stored. */
export const SUPPORTED_LOCALES = [DEFAULT_LOCALE, 'km'] as const

export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]

export function isSupportedLocale(value: unknown): value is SupportedLocale {
  return typeof value === 'string' && (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

/** Narrows anything to a language the API will answer. */
export function resolveLocale(value: unknown): SupportedLocale {
  return isSupportedLocale(value) ? value : DEFAULT_LOCALE
}
