// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { navItems } from '~/content/navigation'

/**
 * Read from disk rather than imported.
 *
 * The i18n module compiles anything under i18n/locales into message functions
 * at build time, so `import en from '.../en.json'` hands back a compiled object
 * with source-map metadata, not the file. These assertions are about the file.
 */
const load = (code: string) =>
  JSON.parse(readFileSync(resolve(process.cwd(), `i18n/locales/${code}.json`), 'utf8'))

const en = load('en')
const km = load('km')

type Tree = { [key: string]: string | Tree }

function flatten(tree: Tree, prefix = ''): Record<string, string> {
  const out: Record<string, string> = {}
  for (const [key, value] of Object.entries(tree)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'string') out[path] = value
    else Object.assign(out, flatten(value, path))
  }
  return out
}

/**
 * Keys that are deliberately identical in both languages: initialisms and
 * product names that stay Latin in Khmer technical writing, and an example
 * email address. Anything else matching its English is a key that was copied
 * across and never translated.
 */
const SAME_IN_BOTH = new Set([
  'work.api',
  'contact.placeholder.email',
  'contact.type.Backend / API',
  'contact.type.WordPress / SEO',
  'contact.type.DevOps',
])

const flatEn = flatten(en as Tree)
const flatKm = flatten(km as Tree)

describe('locale files', () => {
  /**
   * A key added to English but not Khmer does not fail the build — vue-i18n
   * falls back — so the Khmer page would quietly show an English word. This is
   * the only thing that catches it.
   */
  it('has exactly the same keys in both languages', () => {
    expect(Object.keys(flatKm).sort()).toEqual(Object.keys(flatEn).sort())
  })

  it('has no empty strings', () => {
    for (const [key, value] of Object.entries({ ...flatEn, ...flatKm })) {
      expect(value.trim(), key).not.toBe('')
    }
  })

  /**
   * Khmer has its own script, so an identical string means a key was copied
   * across and never translated. `API` is the exception: it is an initialism
   * that stays Latin in Khmer technical writing.
   */
  it('actually translates every Khmer string', () => {
    const untranslated = Object.keys(flatEn).filter(
      (key) => flatEn[key] === flatKm[key] && !SAME_IN_BOTH.has(key),
    )
    expect(untranslated).toEqual([])
  })

  it('writes Khmer in Khmer script', () => {
    const khmer = /[ក-៿]/
    const latinOnly = Object.entries(flatKm)
      .filter(([key]) => !SAME_IN_BOTH.has(key) && key !== 'work.wordpressSeo')
      .filter(([, value]) => !khmer.test(value))
      .map(([key]) => key)
    expect(latinOnly).toEqual([])
  })

  it('covers every navigation item', () => {
    for (const item of navItems) {
      expect(flatEn[`nav.${item.key}`], item.key).toBeTruthy()
      expect(flatKm[`nav.${item.key}`], item.key).toBeTruthy()
    }
  })
})
