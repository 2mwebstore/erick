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

  it('has a Windows tile config whose image exists', () => {
    const config = readFileSync(resolve(publicDir, 'browserconfig.xml'), 'utf8')
    const images = [...config.matchAll(/src="([^"]+)"/g)].map((m) => m[1]!)
    expect(images.length).toBeGreaterThan(0)
    for (const src of images) expect(exists(src), src).toBe(true)
  })

  // Modern browsers prefer an SVG icon when one is linked, so a leftover link
  // would keep showing the old monogram instead of the PNG set.
  it('links no SVG icon', () => {
    const config = readFileSync(resolve(process.cwd(), 'nuxt.config.ts'), 'utf8')
    const iconLinks = [...config.matchAll(/\{ rel: '(?:icon|apple-touch-icon)'[^}]*\}/g)].map((m) => m[0])
    expect(iconLinks.some((link) => link.includes('.svg'))).toBe(false)
    expect(exists('/favicon.svg')).toBe(false)
  })

  // The old address still gets asked for; it must lead to an icon that exists.
  it('redirects the old /favicon.svg to the current icon', () => {
    const config = readFileSync(resolve(process.cwd(), 'nuxt.config.ts'), 'utf8')
    const target = /'\/favicon\.svg': \{ redirect: \{ to: '([^']+)', statusCode: 301 \} \}/.exec(config)?.[1]
    expect(target, 'a 301 rule for /favicon.svg').toBeDefined()
    expect(exists(target!), target).toBe(true)
  })
})
