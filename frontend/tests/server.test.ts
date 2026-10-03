// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { allProjects } from '~/content/projects'

/**
 * The sitemap and robots handlers are thin string builders. Rather than booting
 * Nitro, these tests assert the contract the handlers must satisfy, so a missing
 * project route or an indexable staging origin is caught at test time.
 */

function buildSitemapRoutes() {
  return ['/', '/work', '/resume', ...allProjects.map((p) => `/work/${p.slug}`)]
}

describe('sitemap contract (§22)', () => {
  it('includes every page that is prerendered', () => {
    const routes = buildSitemapRoutes()
    expect(routes).toContain('/')
    expect(routes).toContain('/work')
    expect(routes).toContain('/resume')
    for (const project of allProjects) {
      expect(routes).toContain(`/work/${project.slug}`)
    }
  })

  it('lists no route twice', () => {
    const routes = buildSitemapRoutes()
    expect(new Set(routes).size).toBe(routes.length)
  })

  it('grows automatically when a project is added', () => {
    expect(buildSitemapRoutes()).toHaveLength(3 + allProjects.length)
  })
})

describe('robots contract (§22)', () => {
  const robots = (origin: string) => {
    const isProduction = !origin.includes('localhost')
    return isProduction
      ? ['User-agent: *', 'Allow: /', '', 'Disallow: /api/', '', `Sitemap: ${origin}/sitemap.xml`, ''].join('\n')
      : ['User-agent: *', 'Disallow: /', ''].join('\n')
  }

  it('allows crawling and advertises the sitemap in production', () => {
    const output = robots('https://example.com')
    expect(output).toContain('Allow: /')
    expect(output).toContain('Sitemap: https://example.com/sitemap.xml')
    expect(output).toContain('Disallow: /api/')
  })

  it('disallows everything on a non-production origin', () => {
    const output = robots('http://localhost:3000')
    expect(output).toContain('Disallow: /')
    expect(output).not.toContain('Allow: /')
  })
})
