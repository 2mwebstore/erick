import en from './locales/en.json'
import km from './locales/km.json'

/**
 * Messages are bundled rather than fetched per locale.
 *
 * Two languages of UI chrome is a few kilobytes of JSON — far less than the
 * round trip lazy loading costs on the first language switch. It also means
 * `t()` resolves synchronously everywhere, including in tests, where lazily
 * loaded files never arrive and every assertion about visible text was reading
 * back the translation key instead.
 */
export default defineI18nConfig(() => ({
  legacy: false,
  messages: { en, km },
}))
