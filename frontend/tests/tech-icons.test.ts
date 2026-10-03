import { describe, expect, it } from 'vitest'
import { techIcon, techIconMap, techIconMeta } from '~/utils/tech-icons'
import { projects } from '~/content/projects'
import { capabilities } from '~/content/capabilities'
import { siteConfig } from '~/content/site'
import simpleIcons from '@iconify-json/simple-icons/icons.json'
import lucide from '@iconify-json/lucide/icons.json'

describe('technology icons', () => {
  it('maps a known technology to an icon', () => {
    expect(techIcon('Go')).toBe('simple-icons:go')
    expect(techIcon('MySQL')).toBe('simple-icons:mysql')
  })

  it('carries a brand colour for brand marks', () => {
    expect(techIconMeta('Go')?.color).toBe('#00ADD8')
    expect(techIconMeta('Docker')?.color).toBe('#2496ED')
  })

  it('uses only well-formed six-digit hex colours', () => {
    for (const [tech, meta] of Object.entries(techIconMap)) {
      for (const key of ['color', 'dark'] as const) {
        const value = meta[key]
        if (value === undefined) continue
        expect(value, `${tech}.${key} is not a #rrggbb hex`).toMatch(/^#[0-9A-Fa-f]{6}$/)
      }
    }
  })

  it('gives a dark override to every colour too dark to read on near-black', () => {
    // Relative luminance of the light-mode colour; anything this dark would
    // vanish against #0a0a0b, so it must declare a lighter dark-mode variant.
    const luminance = (hex: string) => {
      const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255)
      const lin = (c: number) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4)
      return 0.2126 * lin(r!) + 0.7152 * lin(g!) + 0.0722 * lin(b!)
    }

    for (const [tech, meta] of Object.entries(techIconMap)) {
      if (!meta.color) continue
      if (luminance(meta.color) < 0.12) {
        expect(meta.dark, `${tech} (${meta.color}) needs a dark-mode colour`).toBeDefined()
      }
    }
  })

  it('gives a light override to every colour too pale to read on white', () => {
    const luminance = (hex: string) => {
      const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255)
      const lin = (c: number) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4)
      return 0.2126 * lin(r!) + 0.7152 * lin(g!) + 0.0722 * lin(b!)
    }

    // The light-mode value itself must not be near-white.
    for (const [tech, meta] of Object.entries(techIconMap)) {
      if (!meta.color) continue
      expect(luminance(meta.color), `${tech} (${meta.color}) is too pale on white`).toBeLessThan(0.62)
    }
  })

  it('ignores case and a trailing version number', () => {
    expect(techIcon('nuxt')).toBe(techIcon('Nuxt'))
    expect(techIcon('Nuxt 4')).toBe(techIcon('Nuxt'))
    expect(techIcon('Vue 3')).toBe(techIcon('Vue'))
    expect(techIcon('  Docker  ')).toBe('simple-icons:docker')
  })

  it('returns undefined rather than a wrong icon for an unmapped term', () => {
    expect(techIcon('Schema design')).toBeUndefined()
    expect(techIcon('Query optimisation')).toBeUndefined()
    expect(techIcon('')).toBeUndefined()
  })

  it('references only icons that exist in the installed collections', () => {
    // Guards against a typo silently shipping an empty box. Iconify renders
    // nothing for an unknown name, so this would otherwise be invisible.
    const collections: Record<string, { icons: Record<string, unknown>; aliases?: Record<string, unknown> }> = {
      'simple-icons': simpleIcons as never,
      'lucide': lucide as never,
    }

    for (const [tech, meta] of Object.entries(techIconMap)) {
      const [prefix, name] = meta.icon.split(':')
      const collection = collections[prefix!]
      expect(collection, `${tech} uses unknown collection "${prefix}"`).toBeDefined()
      const known = Boolean(collection!.icons[name!] || collection!.aliases?.[name!])
      expect(known, `${tech} → ${meta.icon} does not exist`).toBe(true)
    }
  })

  it('covers every technology used in the hero stack', () => {
    for (const tech of siteConfig.heroStack) {
      expect(techIcon(tech), `no icon for hero technology "${tech}"`).toBeDefined()
    }
  })

  it('covers the brand-name technologies listed on projects', () => {
    // Concept terms legitimately have no mark; brand names should have one.
    const brands = new Set(
      projects
        .flatMap((p) => p.technologies)
        .filter((t) => /^(Go|Vue|Nuxt|Flutter|MySQL|Docker|Nginx|Cloudflare|TypeScript|Tailwind CSS|WordPress|PHP|Laravel)( \d+)?$/.test(t)),
    )

    for (const tech of brands) {
      expect(techIcon(tech), `no icon for "${tech}"`).toBeDefined()
    }
  })

  it('does not crash on any capability item', () => {
    for (const capability of capabilities) {
      for (const item of capability.items) {
        expect(() => techIcon(item)).not.toThrow()
      }
    }
  })
})
