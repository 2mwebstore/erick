// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { computed, ref } from 'vue'
import UiCodeBackdrop from '~/components/ui/CodeBackdrop.vue'
import { techIconMeta } from '~/utils/tech-icons'
import { fallbackContent } from '~/content/fallback'
import type { SiteContentPayload } from '~/types'

const settings = ref<Record<string, string>>({})
const projectTech = ref<string[][]>([])
const capabilityItems = ref<string[][]>([])

mockNuxtImport('useSiteContent', () => () => {
  const payload = computed<SiteContentPayload>(() => ({
    ...fallbackContent,
    settings: { ...fallbackContent.settings, ...settings.value },
  }))
  return {
    payload,
    site: computed(() => resolveSite(payload.value.settings)),
    projects: computed(() =>
      projectTech.value.map((technologies, i) => ({ slug: `p${i}`, technologies })),
    ),
    featuredProjects: computed(() => []),
    experience: computed(() => []),
    capabilities: computed(() => capabilityItems.value.map((items, i) => ({ id: i, items }))),
    services: computed(() => []),
    principles: computed(() => []),
    pillars: computed(() => []),
    usingFallback: computed(() => false),
  }
})

const mount = () => mountSuspended(UiCodeBackdrop)

const chipNames = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findAll('.tech-chip').map((chip) => chip.text())

beforeEach(() => {
  settings.value = {}
  projectTech.value = []
  capabilityItems.value = []
})

describe('code backdrop', () => {
  /**
   * It is scenery. Anything here that reached the accessibility tree would be
   * read out as if it were content, and anything that took a click would steal
   * one from the buttons underneath.
   */
  it('is hidden from assistive technology and takes no clicks', async () => {
    const wrapper = await mount()
    const root = wrapper.get('div')

    expect(root.attributes('aria-hidden')).toBe('true')
    expect(root.classes()).toContain('pointer-events-none')
    expect(root.classes()).toContain('-z-10')
  })

  /**
   * The pane shows the real settings rather than invented code (§27). If it
   * ever stops doing that, it becomes a claim nobody is maintaining.
   */
  it('shows the name, where and since when, from the settings', async () => {
    settings.value = {
      name: 'A REAL NAME',
      location: 'Somewhere',
      hero_stack: '["Go","Laravel","Vue","Nuxt"]',
      start_year: '2020',
    }
    const wrapper = await mount()
    const pane = wrapper.get('.code-pane').text()

    expect(pane).toContain('A REAL NAME')
    expect(pane).toContain('Somewhere')
    expect(pane).toContain('2020')
  })

  /**
   * The hero prints the stack in full in the row under the buttons. The pane
   * used to print three of the same words a few hundred pixels away.
   */
  it('does not reprint the stack the hero already shows', async () => {
    settings.value = { hero_stack: '["Go","Laravel","Vue"]', location: 'Somewhere' }
    const wrapper = await mount()
    const pane = wrapper.get('.code-pane').text()

    expect(pane).not.toContain('stack')
    for (const tech of ['Go', 'Laravel', 'Vue']) expect(pane).not.toContain(tech)
  })

  /** A field with nothing in it is left out, not printed as an empty string. */
  it('drops a line it has no value for', async () => {
    settings.value = { name: 'A REAL NAME', location: '', start_year: '2020' }
    const wrapper = await mount()
    const pane = wrapper.get('.code-pane').text()

    expect(pane).not.toContain('from')
    expect(pane).not.toContain('""')
    expect(pane).toContain('A REAL NAME')
  })

  /**
   * The chips are the stack, not a list of technologies someone thought looked
   * good. Anything without a brand mark is dropped rather than drawn as a bare
   * word, and there are only so many places for one to sit.
   */
  it('floats the real stack, in the real brand colours', async () => {
    settings.value = {
      hero_stack: '["Go","Laravel","Vue","Nuxt","Flutter","MySQL","Docker","Cloudflare"]',
    }
    const wrapper = await mount()
    const chips = wrapper.findAll('.tech-chip')

    expect(chips.length).toBeGreaterThan(0)
    expect(chips.map((c) => c.text())).toEqual(
      expect.arrayContaining(['Go', 'Laravel', 'Vue']),
    )

    // The official brand hex from utils/tech-icons.ts, and its dark variant
    // where the brand colour is unreadable on a near-black ground.
    const go = chips.find((c) => c.text() === 'Go')!
    expect(go.attributes('style')).toContain(techIconMeta('Go')!.color)
    const vue = chips.find((c) => c.text() === 'Vue')!
    expect(vue.attributes('style')).toContain(techIconMeta('Vue')!.dark)
  })

  it('drops a technology it has no mark for, rather than drawing a bare word', async () => {
    settings.value = { hero_stack: '["Go","Something Invented","Vue"]' }
    const wrapper = await mount()

    expect(wrapper.findAll('.tech-chip').map((c) => c.text())).toEqual(['Go', 'Vue'])
  })

  it('never floats more chips than it has places to put them', async () => {
    settings.value = {
      hero_stack: JSON.stringify(Array.from({ length: 30 }, () => 'Go')),
    }
    const wrapper = await mount()

    expect(wrapper.findAll('.tech-chip').length).toBeLessThanOrEqual(8)
  })

  it('follows the settings when they change', async () => {
    settings.value = { name: 'SOMEONE ELSE', hero_stack: '["Go"]', start_year: '2019' }
    const wrapper = await mount()

    expect(wrapper.text()).toContain('SOMEONE ELSE')
    expect(wrapper.text()).toContain('2019')
  })

  /** No output and no timings: an invented build time is an invented fact. */
  it('shows commands only, never results', async () => {
    const wrapper = await mount()
    const text = wrapper.text()

    expect(text).toContain('go build ./cmd/api')
    expect(text).not.toMatch(/\b\d+(\.\d+)?\s*(ms|s)\b/)
    expect(text).not.toMatch(/compiled|passed|ok\s/i)
  })
})

describe('code backdrop styles', () => {
  const css = readFileSync(
    resolve(process.cwd(), 'components/ui/CodeBackdrop.vue'),
    'utf8',
  ).split('<style scoped>')[1] ?? ''

  it('stops every moving thing under reduced motion (§25)', () => {
    const block = /prefers-reduced-motion: reduce\)\s*\{([\s\S]*?)\n\}/.exec(css)?.[1] ?? ''
    expect(block).toContain('animation: none')

    /*
       Every class that animates has to be named in that block. Listing the
       three by hand would pass for a fourth that nobody remembered to add, so
       the rules are read out of the stylesheet instead.
    */
    const animated = new Set<string>()
    for (const [, selector, body] of css.matchAll(/(\.[\w-]+(?:,\s*\.[\w-]+)*)\s*\{([^}]*)\}/g)) {
      if (!/animation:\s*(?!none)\S/.test(body!)) continue
      for (const one of selector!.split(',')) animated.add(one.trim())
    }

    expect(animated.size).toBeGreaterThan(0)
    const stopped = block.slice(0, block.indexOf('{'))
    for (const selector of animated) {
      expect(stopped, `${selector} keeps animating`).toContain(selector)
    }
  })

  const float = /@keyframes float3d\s*\{([\s\S]*?)\n\}/.exec(css)?.[1] ?? ''
  const shadow = /@keyframes float3d-shadow\s*\{([\s\S]*?)\n\}/.exec(css)?.[1] ?? ''

  /**
   * Height on its own reads as a sliding rectangle. The tilt on both axes and
   * the change of scale are what make it read as an object with a near edge and
   * a far one.
   */
  it('floats in three dimensions, not just up and down', () => {
    for (const fn of ['translateY(', 'rotateX(', 'rotateY(', 'scale(']) {
      expect(float, `float3d is missing ${fn})`).toContain(fn)
    }
  })

  /** Transform only, so the float composites and never triggers layout. */
  it('floats with transforms rather than offsets', () => {
    expect(float).not.toMatch(/\b(top|left|right|bottom|margin|width|height):/)
  })

  /** A rotation without perspective is a squashed rectangle, not depth. */
  it('gives the cards a room to float in', () => {
    const container = /\.code-backdrop\s*\{([\s\S]*?)\n\}/.exec(css)?.[1] ?? ''
    expect(container).toMatch(/perspective:\s*1000px/)
    expect(container).toContain('transform-style: preserve-3d')
    // An overflow here would force the container back to `flat` and undo it.
    expect(container).not.toMatch(/overflow:\s*(hidden|clip|auto|scroll)/)
  })

  /**
   * The shadow is what makes the height readable: widest and faintest when the
   * card is highest, tight and dark when it is down.
   */
  it('casts a shadow that spreads and fades as the card rises', () => {
    const at = (percent: string) => {
      const frame = new RegExp(`(^|\\n)\\s*${percent}\\s*\\{([\\s\\S]*?)\\}`).exec(shadow)?.[2] ?? ''
      return {
        scale: Number(/scale\(([\d.]+)\)/.exec(frame)?.[1]),
        opacity: Number(/opacity:\s*([\d.]+)/.exec(frame)?.[1]),
      }
    }
    const low = at('0%')
    const high = at('50%')

    expect(low).toEqual({ scale: 0.8, opacity: 0.6 })
    expect(high).toEqual({ scale: 1.2, opacity: 0.2 })

    // The card is at its highest at the same instant.
    const cardHigh = /(^|\n)\s*50%\s*\{([\s\S]*?)\}/.exec(float)?.[2] ?? ''
    expect(cardHigh).toMatch(/translateY\(-/)
    expect(at('100%')).toEqual(low)
  })

  /**
   * A shadow that lags its object stops looking like a shadow, so both read
   * their timing from the one property the stage sets.
   */
  it('keeps the shadow on the same clock as the card', () => {
    for (const selector of ['.float3d', '.ground-shadow']) {
      const rule = new RegExp(`\\${selector}\\s*\\{([^}]*)\\}`).exec(css)?.[1] ?? ''
      expect(rule).toContain('var(--float-duration')
      expect(rule).toContain('var(--float-delay')
    }
  })

  /** Glass: the page behind it is visible through it, blurred. */
  it('is glass rather than a grey box', () => {
    const pane = /\.code-pane\s*\{([\s\S]*?)\n\}/.exec(css)?.[1] ?? ''
    expect(pane).toMatch(/backdrop-filter:\s*blur\(/)
    expect(pane).toContain('-webkit-backdrop-filter')
    expect(pane).toMatch(/background-image:\s*linear-gradient/)
    // The lit edge along the top, which is what reads as a pane of glass.
    expect(pane).toMatch(/inset 0 1px 0 rgb\(255 255 255/)
  })

  /**
   * Which decorations exist at a width is decided by the utilities in each
   * slot — `hidden lg:inline-flex` and the like. A scoped rule carries the
   * component's data attribute, so a `display` here outranks Tailwind's
   * `hidden` and the element appears at every width: that is exactly how every
   * chip ended up across the eyebrow and the portrait on a 390px screen.
   */
  it('never sets display on an element whose breakpoint decides it', () => {
    for (const selector of ['.tech-chip', '.code-pane', '.code-token']) {
      const body = new RegExp(`\\${selector}\\s*\\{([^}]*)\\}`, 'g')
      for (const [, rule] of css.matchAll(body)) {
        expect(rule, `${selector} sets display`).not.toMatch(/(^|[^-])display:/)
      }
    }
  })

  it('leaves itself off printed pages', () => {
    expect(/@media print\s*\{\s*\.code-backdrop\s*\{\s*display: none/.test(css)).toBe(true)
  })
})

/**
 * The same words, twice on one screen.
 *
 * The hero prints the stack in full in the row under the buttons. Floating the
 * same eight names around it meant every technology on the page appeared twice
 * — three of them three times, counting the editor pane. The chips come from
 * the technologies listed on the real projects instead: the same kind of data,
 * one layer further in, and nothing the visitor can already read.
 */
describe('code backdrop, deduplicated', () => {
  it('floats technologies the hero row does not already show', async () => {
    settings.value = { hero_stack: '["Go","Laravel","Vue"]' }
    projectTech.value = [['Go', 'TypeScript', 'Vue'], ['WordPress', 'PHP']]

    const names = chipNames(await mount())

    expect(names).toEqual(['TypeScript', 'WordPress', 'PHP'])
  })

  /** "Nuxt", "Nuxt 4" and "nuxt" are one technology, not three. */
  it('treats a version and a different case as the same technology', async () => {
    settings.value = { hero_stack: '["Nuxt","Vue"]' }
    projectTech.value = [['Nuxt 4', 'Vue 3', 'nuxt', 'TypeScript'], ['TYPESCRIPT']]

    expect(chipNames(await mount())).toEqual(['TypeScript'])
  })

  /** A compound names something in the row, so it is a duplicate too. */
  it('excludes a compound that names a technology from the row', async () => {
    settings.value = { hero_stack: '["Laravel"]' }
    projectTech.value = [['Laravel / PHP', 'Nginx']]

    expect(chipNames(await mount())).toEqual(['Nginx'])
  })

  it('reaches into the capability lists when the projects run out', async () => {
    settings.value = { hero_stack: '["Go"]' }
    projectTech.value = [['Go']]
    capabilityItems.value = [['Firebase', 'Go']]

    expect(chipNames(await mount())).toEqual(['Firebase'])
  })

  /**
   * A new site has no projects, and nothing to duplicate either — better the
   * stack than an empty background.
   */
  it('falls back to the stack when there is nothing else to show', async () => {
    settings.value = { hero_stack: '["Go","Laravel"]' }

    expect(chipNames(await mount())).toEqual(['Go', 'Laravel'])
  })

  it('never shows the same technology twice', async () => {
    settings.value = { hero_stack: '[]' }
    projectTech.value = [['TypeScript', 'Nginx'], ['TypeScript', 'nginx'], ['Nginx']]

    const names = chipNames(await mount())
    expect(names).toEqual([...new Set(names)])
    expect(names).toEqual(['TypeScript', 'Nginx'])
  })
})
