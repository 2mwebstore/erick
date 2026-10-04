// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { resolve } from 'node:path'
import { nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import LayoutHeader from '~/components/layout/Header.vue'
import NavigationMobile from '~/components/navigation/Mobile.vue'
import NavigationThemeToggle from '~/components/navigation/ThemeToggle.vue'
import { navItems } from '~/content/navigation'
import { siteConfig } from '~/content/site'

describe('primary navigation (§9)', () => {
  // The mobile panel teleports to <body> and mounts are not torn down between
  // tests, so a previous test's panel would otherwise be the one found here.
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('lists the homepage sections in the order they appear', () => {
    expect(navItems.map((i) => i.key)).toEqual([
      'experience',
      'work',
      'capabilities',
      'services',
      'contact',
    ])
  })

  /**
   * The scroll spy observes exactly `navItems`, so a homepage section carrying
   * an id but missing from the nav is a stretch of page it cannot see. The old
   * spy only ever assigned a non-empty value, so the highlight stuck on whatever
   * was lit when that stretch was entered — which is what adding the Experience
   * section did. The current spy clears it instead, so a section can be left out
   * of the nav on purpose; it just has to be named here, so that a new section
   * is never left out by accident.
   */
  const LEFT_OUT_OF_NAV = ['about']

  it('has a nav item for every homepage section that anchors one, bar those left out on purpose', () => {
    const dir = resolve(process.cwd(), 'components/sections')
    const anchored = readdirSync(dir)
      .filter((file) => file.endsWith('.vue'))
      .flatMap((file) => {
        const source = readFileSync(resolve(dir, file), 'utf8')
        // Only the section element's own id — form fields carry ids too.
        return [...source.matchAll(/<(?:UiSection|section)\b[^>]*>/g)]
          .map((tag) => /\sid="([a-z-]+)"/.exec(tag[0])?.[1])
          .filter((id): id is string => Boolean(id))
      })
      .sort()

    expect(anchored).toEqual([...navItems.map((i) => i.section), ...LEFT_OUT_OF_NAV].sort())
  })

  it('points every item at an anchor that a homepage section actually renders', () => {
    for (const item of navItems) {
      expect(item.hash).toBe(item.section)
    }
  })

  /**
   * Public links go through localePath, or useSectionLink for a homepage
   * section. Written by hand they pointed at English from every Khmer page —
   * project cards, "All work", the error page — so a Khmer reader who clicked
   * was switched language, and the Khmer project pages were linked from nothing
   * but the language switcher. Section links built by hand came out as
   * /km/#work, which costs a 301 on a full page load.
   *
   * The admin panel is English-only and is left out.
   */
  it('builds every public internal link in the reader\'s language', () => {
    const handWritten = /\bto="\/|:to="[`']\/|\}\/#\$\{/
    const offenders: string[] = []
    const scan = (path: string) => {
      for (const entry of readdirSync(path, { withFileTypes: true })) {
        const child = resolve(path, entry.name)
        // pages/admin/, components/admin/ and layouts/admin.vue
        if (entry.name === 'admin' || entry.name === 'admin.vue') continue
        if (entry.isDirectory()) {
          scan(child)
        } else if (entry.name.endsWith('.vue') && handWritten.test(readFileSync(child, 'utf8'))) {
          offenders.push(child.replace(process.cwd(), ''))
        }
      }
    }
    for (const dir of ['components', 'pages', 'layouts']) scan(resolve(process.cwd(), dir))
    for (const file of ['app.vue', 'error.vue']) {
      if (handWritten.test(readFileSync(resolve(process.cwd(), file), 'utf8'))) offenders.push(`/${file}`)
    }

    expect(offenders).toEqual([])
  })

  it('carries no labels of its own, so every language reads the same structure', () => {
    for (const item of navItems) {
      expect(item).not.toHaveProperty('label')
    }
  })

  it('renders the name, the nav, and both calls to action', async () => {
    const wrapper = await mountSuspended(LayoutHeader)

    expect(wrapper.text()).toContain(siteConfig.name)
    expect(wrapper.find('nav[aria-label="Primary"]').exists()).toBe(true)

    const hrefs = wrapper.findAll('a').map((a) => a.attributes('href'))
    expect(hrefs).toContain('/resume')
    expect(hrefs).toContain('/#contact')
  })

  it('exposes the mobile menu button state to assistive technology', async () => {
    const wrapper = await mountSuspended(LayoutHeader)

    const toggle = wrapper.find('button[aria-controls="mobile-menu"]')
    expect(toggle.exists()).toBe(true)
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(toggle.attributes('aria-label')).toBe('Open menu')

    await toggle.trigger('click')

    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(toggle.attributes('aria-label')).toBe('Close menu')
  })

  it('opens the mobile panel with the same items and closes on request', async () => {
    const wrapper = await mountSuspended(NavigationMobile, {
      props: { open: true, offset: '4rem' },
    })

    const panel = document.querySelector('#mobile-menu')
    expect(panel).not.toBeNull()
    // Labels come from the locale files now, so the invariant here is the
    // structure: one link per nav item, each with a visible label.
    const links = [...panel!.querySelectorAll('nav a')]
    expect(links).toHaveLength(navItems.length)
    for (const link of links) {
      expect(link.textContent?.trim()).toBeTruthy()
    }

    // Teleported nodes are outside wrapper.element, so click the real one.
    panel!.querySelector<HTMLElement>('a')!.click()
    await nextTick()
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('renders nothing when the mobile panel is closed', async () => {
    await mountSuspended(NavigationMobile, { props: { open: false, offset: '4rem' } })
    expect(document.querySelector('#mobile-menu')).toBeNull()
  })

  /**
   * Regression: the panel used to live inside <header>, which gets a
   * backdrop-filter once the page is scrolled. That makes the header the
   * containing block for `position: fixed` children, so the panel was laid out
   * inside a 56px header and collapsed to nothing — the menu opened to an empty
   * screen, but only after scrolling. It must stay out of the header.
   */
  it('teleports the panel out of the header so no filtered ancestor can contain it', async () => {
    const wrapper = await mountSuspended(NavigationMobile, {
      props: { open: true, offset: '4rem' },
    })

    expect(wrapper.element.querySelector?.('#mobile-menu') ?? null).toBeNull()
    expect(document.body.querySelector('#mobile-menu')).not.toBeNull()
  })

  it('sits flush under the header at whichever height it currently has', async () => {
    await mountSuspended(NavigationMobile, { props: { open: true, offset: '3.5rem' } })
    const panel = document.querySelector('#mobile-menu') as HTMLElement
    expect(panel.style.top).toBe('3.5rem')
  })
})

describe('theme toggle (§4)', () => {
  it('describes the current theme and is operable as a button', async () => {
    const wrapper = await mountSuspended(NavigationThemeToggle)

    const button = wrapper.find('button')
    expect(button.exists()).toBe(true)
    expect(button.attributes('aria-label')).toContain('Theme:')
  })

  it('cycles light, dark, then system', async () => {
    const wrapper = await mountSuspended(NavigationThemeToggle)
    const button = wrapper.find('button')

    const labels: string[] = []
    for (let i = 0; i < 3; i++) {
      await button.trigger('click')
      labels.push(button.attributes('aria-label') ?? '')
    }

    expect(labels[0]).toContain('light')
    expect(labels[1]).toContain('dark')
    expect(labels[2]).toContain('system')
  })

  it('persists the chosen theme so it survives a reload', async () => {
    const wrapper = await mountSuspended(NavigationThemeToggle)
    const button = wrapper.find('button')

    await button.trigger('click') // light
    expect(localStorage.getItem('theme')).toBe('light')

    await button.trigger('click') // dark
    expect(localStorage.getItem('theme')).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    await button.trigger('click') // system — the stored override is removed
    expect(localStorage.getItem('theme')).toBeNull()
  })
})

/**
 * The header's own widths.
 *
 * `UiButton` always emits `inline-flex`, and a `hidden` passed in from the
 * header is the same kind of utility at the same specificity — which of the two
 * wins is decided by the order Tailwind emits them in, not by the order they
 * are written. It emitted `inline-flex` last, so Resume and Let's Talk were on
 * screen at 320px, and the name beside them was truncated to "K…". Hiding is
 * done by a wrapper the button cannot overrule.
 */
describe('header controls', () => {
  const header = readFileSync(resolve(process.cwd(), 'components/layout/Header.vue'), 'utf8')

  it('never puts a display utility on a UiButton', () => {
    const buttons = header.match(/<UiButton[^>]*>/g) ?? []
    expect(buttons.length).toBeGreaterThan(0)
    for (const tag of buttons) {
      expect(tag, tag).not.toMatch(/\b(hidden|inline-flex|block|flex)\b/)
    }
  })

  it('hides the calls to action with a wrapper instead', () => {
    expect(header).toMatch(/<div class="hidden items-center gap-2 lg:flex">/)
  })

  /**
   * One breakpoint, used by all four: six navigation items, the language
   * switcher, the theme button and two calls to action overflowed 768px, and
   * what fell off the edge was the Resume link — which is in no other menu at
   * that width.
   */
  it('switches between the bar and the drawer at one width', () => {
    expect(header).toContain("matchMedia('(min-width: 1024px)')")
    expect(header).toMatch(/<nav [^>]*class="hidden lg:block"/)
    expect(header).toContain('lg:hidden')
    expect(header).not.toMatch(/\bmd:(hidden|block|flex|inline-flex)\b/)

    const drawer = readFileSync(resolve(process.cwd(), 'components/navigation/Mobile.vue'), 'utf8')
    expect(drawer).toContain('lg:hidden')
    expect(drawer).not.toContain('md:hidden')
  })
})
