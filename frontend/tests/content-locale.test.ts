// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { registerEndpoint } from '@nuxt/test-utils/runtime'
import { nextTick, ref } from 'vue'
import plugin from '~/plugins/00.site-content'
import { fallbackContent } from '~/content/fallback'
import { DEFAULT_LOCALE, resolveLocale, SUPPORTED_LOCALES } from '~/utils/locales'
import type { SiteContentPayload } from '~/types'

/**
 * Content has to follow the language.
 *
 * The site is rendered on the server, so a visit straight to /km has always
 * been Khmer. Switching language from a page, though, is a client-side route
 * change, and plugins run once per app: the chrome turned Khmer while the
 * headline, tagline, About paragraphs and experience entries — every word that
 * comes from the database — stayed in whichever language the page was first
 * rendered in. Nothing in the suite noticed, because nothing here changed the
 * locale after the app had started. These tests do.
 */

/** Which languages were asked of the content proxy, in order. */
const asked: string[] = []
let fail = false

registerEndpoint('/api/content', (event) => {
  const locale = new URL(event.node.req.url ?? '/', 'http://test').searchParams.get('locale') ?? ''
  asked.push(locale)
  if (fail) throw createError({ statusCode: 502, statusMessage: 'upstream down' })
  return { ...fallbackContent, settings: { ...fallbackContent.settings, name: `name-${locale}` } }
})

/** Runs the plugin's client half against a locale we control. */
async function start(locale: string) {
  const current = ref(locale)
  await (plugin as { setup: (app: unknown) => Promise<void> | void }).setup({
    $i18n: { locale: current },
  })
  await nextTick()
  return current
}

/** Waits for the watcher's fetch to settle. */
const settle = () => new Promise((resolve) => setTimeout(resolve, 20))

function seed(locale: string, degraded = false) {
  useState<SiteContentPayload>('site-content').value = {
    ...fallbackContent,
    settings: { ...fallbackContent.settings, name: `name-${locale}` },
  }
  useState<string>('site-content-locale').value = locale
  useState<boolean>('site-content-degraded').value = degraded
}

beforeEach(() => {
  asked.length = 0
  fail = false
  seed(DEFAULT_LOCALE)
})

describe('content follows the language', () => {
  it('asks for nothing when the payload already matches the page', async () => {
    await start(DEFAULT_LOCALE)
    await settle()

    // The server already fetched this language and serialised it. Fetching it
    // again on every hydration would double the work for no new content.
    expect(asked).toEqual([])
  })

  it('refetches in the new language when the locale changes', async () => {
    const locale = await start(DEFAULT_LOCALE)

    locale.value = 'km'
    await nextTick()
    await settle()

    expect(asked).toEqual(['km'])
    expect(useState<SiteContentPayload>('site-content').value.settings.name).toBe('name-km')
    expect(useState<string>('site-content-locale').value).toBe('km')
  })

  it('follows a switch back, rather than caching the first change', async () => {
    const locale = await start(DEFAULT_LOCALE)

    locale.value = 'km'
    await nextTick()
    await settle()
    locale.value = 'en'
    await nextTick()
    await settle()

    expect(asked).toEqual(['km', 'en'])
    expect(useState<SiteContentPayload>('site-content').value.settings.name).toBe('name-en')
  })

  /**
   * The state says which language it holds, not which language was wanted. If
   * server rendering fell back, the URL and the payload disagree, and the
   * client is the last chance to put that right.
   */
  it('repairs a page whose server-side fetch fell back to bundled content', async () => {
    seed(DEFAULT_LOCALE, true)

    await start('km')
    await settle()

    expect(asked).toEqual(['km'])
    expect(useState<boolean>('site-content-degraded').value).toBe(false)
  })

  it('retries even when the language matches, if the content is bundled', async () => {
    seed('km', true)

    await start('km')
    await settle()

    expect(asked).toEqual(['km'])
  })

  /**
   * A language switch that cannot reach the API must not empty the page: the
   * words already on screen are in the wrong language, which is a far smaller
   * failure than no words at all.
   */
  it('keeps the content it is showing when a refetch fails', async () => {
    const locale = await start(DEFAULT_LOCALE)
    fail = true

    locale.value = 'km'
    await nextTick()
    await settle()

    expect(useState<SiteContentPayload>('site-content').value.settings.name).toBe('name-en')
    expect(useState<boolean>('site-content-degraded').value).toBe(true)
    // Still recorded as English, so the next attempt knows it has work to do.
    expect(useState<string>('site-content-locale').value).toBe(DEFAULT_LOCALE)
  })
})

describe('locale narrowing', () => {
  it('answers an unknown or missing language in the source language', () => {
    expect(resolveLocale('km')).toBe('km')
    expect(resolveLocale('en')).toBe('en')
    expect(resolveLocale('fr')).toBe(DEFAULT_LOCALE)
    expect(resolveLocale(undefined)).toBe(DEFAULT_LOCALE)
    expect(resolveLocale(['km'])).toBe(DEFAULT_LOCALE)
    // A query string arrives as a string; an object must not slip through.
    expect(resolveLocale({ toString: () => 'km' })).toBe(DEFAULT_LOCALE)
  })

  it('treats English as the source text, not a translation', () => {
    expect(SUPPORTED_LOCALES[0]).toBe(DEFAULT_LOCALE)
  })
})
