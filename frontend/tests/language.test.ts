// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import NavigationLanguageSwitcher from '~/components/navigation/LanguageSwitcher.vue'
import { REQUIRED_API_CONTRACT } from '~/composables/useAdmin'

describe('language switcher', () => {
  /**
   * The panel is hidden, not removed. Both addresses have to be in the markup
   * a crawler reads — and readable without JavaScript, which is exactly the
   * case where the dropdown never opens at all.
   */
  it('keeps both languages in the markup while the panel is closed', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)

    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    expect(wrapper.findAll('a')).toHaveLength(2)
    expect(wrapper.find('[style*="display: none"]').exists()).toBe(true)
  })

  it('opens on click and closes again', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)
    const trigger = wrapper.get('button')

    await trigger.trigger('click')
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(wrapper.find('[style*="display: none"]').exists()).toBe(false)

    await trigger.trigger('click')
    expect(trigger.attributes('aria-expanded')).toBe('false')
  })

  /** Escape is how a keyboard leaves a menu; the trigger gets focus back. */
  it('closes on Escape', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher, { attachTo: document.body })
    const trigger = wrapper.get('button')

    await trigger.trigger('click')
    expect(trigger.attributes('aria-expanded')).toBe('true')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()

    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)
    wrapper.unmount()
  })

  /**
   * The trigger is a globe, not a flag: a flag names a country, and the control
   * changes a language. Its accessible name says which one is current, because
   * the flag and the two-letter code do not say it to a screen reader.
   */
  it('names the current language on the trigger', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)
    const trigger = wrapper.get('button')

    expect(trigger.attributes('aria-label')).toContain('English')
    expect(trigger.attributes('aria-controls')).toBeTruthy()
    // Globe and chevron, both decorative: the name above is what is announced.
    expect(trigger.findAll('svg').length).toBeGreaterThanOrEqual(2)
    expect(trigger.findAll('[aria-hidden="true"]').length).toBeGreaterThanOrEqual(2)
  })

  it('offers both languages as real links', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)

    const links = wrapper.findAll('a')
    expect(links).toHaveLength(2)
    for (const link of links) {
      expect(link.attributes('href')).toBeTruthy()
    }
  })

  /**
   * A flag is a country, not a language, so it decorates the name rather than
   * replacing it. Losing the text would leave the control unreadable to anyone
   * who does not recognise the flag.
   */
  it('shows a flag beside each language without replacing its name', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)

    expect(wrapper.findAll('svg').length).toBeGreaterThanOrEqual(2)
    expect(wrapper.text()).toContain('EN')
    expect(wrapper.text()).toContain('ភាសាខ្មែរ')
  })

  it('marks the current language for assistive technology', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)
    expect(wrapper.findAll('[aria-current="true"]')).toHaveLength(1)
  })

  it('tags each option with its own language so the right font is used', async () => {
    const wrapper = await mountSuspended(NavigationLanguageSwitcher)
    const langs = wrapper.findAll('a').map((a) => a.attributes('lang'))
    expect(langs).toEqual(['en', 'km'])
  })
})

describe('API contract', () => {
  /**
   * The panel refuses to trust a server older than the contract it was built
   * against, because such a server answers current request bodies with
   * "Malformed request." and nothing else. backend/tests/contract_test.go
   * asserts the Go constant matches this one.
   */
  it('is a positive number, so a server that omits it reads as stale', () => {
    expect(REQUIRED_API_CONTRACT).toBeGreaterThan(0)
    expect(Number.isInteger(REQUIRED_API_CONTRACT)).toBe(true)
  })
})
