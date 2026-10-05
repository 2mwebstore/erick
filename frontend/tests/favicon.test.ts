// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

/**
 * The icon set lives in public/ and is referenced by path from two places: the
 * head tags in nuxt.config.ts and site.webmanifest. A path that points at a
 * file that is not there fails silently — the browser just shows no icon — so
 * every reference is checked against the files.
 */
const publicDir = resolve(process.cwd(), 'public')
const exists = (path: string) => existsSync(resolve(publicDir, path.replace(/^\//, '')))

describe('favicon', () => {
  it('links only icons that exist in public/', () => {
    const config = readFileSync(resolve(process.cwd(), 'nuxt.config.ts'), 'utf8')
    const hrefs = [...config.matchAll(/rel: '(?:icon|apple-touch-icon|manifest)'[^}]*href: '([^']+)'/g)].map((m) => m[1]!)

    expect(hrefs.length).toBeGreaterThanOrEqual(4)
    for (const href of hrefs) expect(exists(href), href).toBe(true)
  })

  it('answers /favicon.ico, which browsers and crawlers ask for without reading the page', () => {
    expect(exists('/favicon.ico')).toBe(true)
  })

  it('has a manifest whose icons exist', () => {
    const manifest = JSON.parse(readFileSync(resolve(publicDir, 'site.webmanifest'), 'utf8')) as {
      icons: { src: string; sizes: string }[]
    }
    expect(manifest.icons.length).toBeGreaterThan(0)
    for (const icon of manifest.icons) expect(exists(icon.src), icon.src).toBe(true)
  })

  // Modern browsers prefer an SVG icon when one is linked, so a leftover one
  // would keep showing instead of the PNG set.
  it('does not leave the old SVG icon behind', () => {
    const config = readFileSync(resolve(process.cwd(), 'nuxt.config.ts'), 'utf8')
    expect(config).not.toContain('favicon.svg')
  })
})
