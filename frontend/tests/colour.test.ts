// @vitest-environment node
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

/**
 * The palette, read from the stylesheet rather than restated here.
 *
 * This one is a single hue ramp, so its middle is the problem: #4A7FA7 is
 * 4.10:1 on the pale end and 4.08:1 on the dark end, clearing 4.5:1 against
 * neither. It is the wash, never text, and the accents are shades either side
 * of it. This file is what stops a later tidy-up quietly promoting it back
 * into a text colour.
 */
const css = readFileSync(resolve(process.cwd(), 'assets/css/main.css'), 'utf8')

function tokens(selector: string): Record<string, string> {
  const block = new RegExp(`${selector}\\s*\\{([\\s\\S]*?)\\n\\}`).exec(css)
  if (!block) throw new Error(`no ${selector} block in main.css`)

  const out: Record<string, string> = {}
  for (const [, name, value] of block[1]!.matchAll(/--([\w-]+):\s*(#[0-9a-f]{6})\s*;/gi)) {
    out[name!] = value!
  }
  return out
}

const channel = (c: number) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4)

function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => channel(parseInt(hex.slice(i, i + 2), 16) / 255))
  return 0.2126 * r! + 0.7152 * g! + 0.0722 * b!
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x)
  return (hi! + 0.05) / (lo! + 0.05)
}

const THEMES = { light: tokens(':root'), dark: tokens(':root\\.dark') }

describe.each(Object.entries(THEMES))('%s theme', (_name, t) => {
  /** Every combination the components actually put together. */
  const pairs: [string, string, string][] = [
    ['body text on the page', 'fg', 'bg'],
    ['body text on a card', 'fg', 'surface'],
    ['secondary text on the page', 'fg-muted', 'bg'],
    ['secondary text on a card', 'fg-muted', 'surface'],
    ['meta text on the page', 'fg-subtle', 'bg'],
    ['meta text on a card', 'fg-subtle', 'surface'],
    ['links on the page', 'accent', 'bg'],
    ['links on a card', 'accent', 'surface'],
    ['links on a raised panel', 'accent', 'elevated'],
    ['button label on the gold fill', 'accent-fg', 'accent-solid'],
  ]

  it.each(pairs)('%s clears 4.5:1', (_label, fg, bg) => {
    expect(contrast(t[fg]!, t[bg]!)).toBeGreaterThanOrEqual(4.5)
  })

  it('has a fill dark or light enough to carry a label', () => {
    expect(contrast(t['accent-fg']!, t['accent-solid']!)).toBeGreaterThanOrEqual(4.5)
  })

  it('has a hover state at least as readable as the resting one', () => {
    expect(contrast(t['accent-hover']!, t['bg']!)).toBeGreaterThanOrEqual(
      contrast(t['accent']!, t['bg']!) - 0.01,
    )
  })
})

describe('palette', () => {
  it('uses the five given colours where each one belongs', () => {
    expect(THEMES.light.bg).toBe('#f6fafd')
    expect(THEMES.light.fg).toBe('#0a1931')
    expect(THEMES.light['fg-muted']).toBe('#1a3d63')
    expect(THEMES.light['border-strong']).toBe('#b3cfe5')
    expect(THEMES.dark.bg).toBe('#0a1931')
    expect(THEMES.dark.elevated).toBe('#1a3d63')
    expect(THEMES.dark['accent-solid']).toBe('#b3cfe5')
  })

  /** The whole reason the mid blue is decoration and not a text colour. */
  it('keeps the mid blue out of every text slot', () => {
    expect(contrast('#4a7fa7', '#f6fafd')).toBeLessThan(4.5)
    expect(contrast('#4a7fa7', '#0a1931')).toBeLessThan(4.5)

    for (const theme of Object.values(THEMES)) {
      for (const slot of ['fg', 'fg-muted', 'fg-subtle', 'accent']) {
        expect(theme[slot]).not.toBe('#4a7fa7')
      }
    }
  })

  it('gives the mid blue the one job with no contrast requirement', () => {
    expect(THEMES.light['wash-a']).toBe('#4a7fa7')
    expect(THEMES.dark['wash-a']).toBe('#4a7fa7')
  })
})

/**
 * The gradient-painted name.
 *
 * It is text painted with a gradient, so it is not one colour on one background
 * but every stop in that gradient. The weakest one is what has to clear 4.5:1 —
 * and it has to in both themes, since the stops are tokens each theme
 * redefines.
 */
describe('hero headline gradient', () => {
  const rule = /\.headline-gradient\s*\{([\s\S]*?)\n {2}\}/.exec(css)?.[1] ?? ''

  /** The `var(--x)` stops of the gradient, in the order they are painted. */
  const stops = [...rule.matchAll(/var\(--([\w-]+)\)/g)].map((m) => m[1]!)

  it('paints the glyphs with palette tokens rather than invented colours', () => {
    expect(stops.length).toBeGreaterThan(0)
    for (const stop of stops) {
      expect(THEMES.light[stop], `--${stop} is not a palette token`).toBeTruthy()
    }
  })

  it.each(Object.entries(THEMES))('is readable at every frame in the %s theme', (_name, t) => {
    for (const stop of stops) {
      expect(contrast(t[stop]!, t.bg!), `--${stop}`).toBeGreaterThanOrEqual(4.5)
    }
  })

  /**
   * Clipping a gradient to text needs transparent text, and transparent text in
   * a browser that does not clip is an invisible headline. The colour may only
   * be dropped where the clip is known to work.
   */
  it('only makes the text transparent inside a support query', () => {
    const guarded = /@supports[^{]*background-clip:\s*text[^{]*\{[\s\S]*?color:\s*transparent/.exec(
      css,
    )
    expect(guarded).not.toBeNull()

    const unguarded = rule.includes('color: transparent')
    expect(unguarded).toBe(false)
  })

  /**
   * Colour in motion, and the loop has to be seamless.
   *
   * The ramp travels by exactly one tile of a repeating background, and its
   * first and last stop are the same token — so the frame the animation ends on
   * is pixel-identical to the one it started on and nothing jumps. A tile of a
   * different width, or an end stop that did not match the start, would show a
   * seam crossing the name once per cycle.
   */
  it('loops seamlessly rather than restarting', () => {
    expect(rule).toContain('background-repeat: repeat')

    const size = /background-size:\s*(\d+)%/.exec(rule)?.[1]
    const from = /@keyframes headline-gradient-flow\s*\{[\s\S]*?from\s*\{\s*background-position:\s*(\d+)%/
      .exec(css)?.[1]
    const to = /@keyframes headline-gradient-flow\s*\{[\s\S]*?to\s*\{\s*background-position:\s*(\d+)%/
      .exec(css)?.[1]

    expect(from).toBe('0')
    expect(to).toBe(size)

    const first = stops.at(0)
    const last = stops.at(-1)
    expect(last, 'the ramp must end on the colour it starts with').toBe(first)
  })

  it('travels at a constant speed, so there is no point it pauses at', () => {
    const animation = /animation:\s*headline-gradient-flow[^;]*/.exec(css)?.[0] ?? ''
    expect(animation).toContain('linear')
    expect(animation).toContain('infinite')
  })

  it('stops moving under reduced motion, without losing the colour (§25)', () => {
    const reduced = /prefers-reduced-motion: reduce\)\s*\{\s*\.headline-gradient\s*\{([\s\S]*?)\n {4}\}/
      .exec(css)?.[1]
    expect(reduced).toBeTruthy()
    expect(reduced).toContain('animation: none')
    expect(reduced).not.toContain('background-image: none')
    // Parked somewhere along the ramp rather than at 0%, where the name would
    // be a single flat colour and the gradient pointless.
    expect(reduced).toMatch(/background-position:\s*[1-9]/)
  })

  /** Forced colours and print both drop the paint; neither may drop the words. */
  it('falls back to plain ink where a gradient cannot be painted', () => {
    const fallback = /@media \(forced-colors: active\), print\s*\{\s*\.headline-gradient\s*\{([\s\S]*?)\n {4}\}/
      .exec(css)?.[1]
    expect(fallback).toBeTruthy()
    expect(fallback).toContain('background-image: none')
    expect(fallback).toContain('color: var(--fg)')
  })
})
