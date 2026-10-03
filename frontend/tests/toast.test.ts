// @vitest-environment nuxt
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useToast } from '~/composables/useToast'
import UiToasts from '~/components/ui/Toasts.vue'

describe('useToast', () => {
  beforeEach(() => {
    useToast().toasts.value = []
    vi.useFakeTimers()
  })
  afterEach(() => vi.useRealTimers())

  it('adds a message and hands back its id', () => {
    const toast = useToast()
    const id = toast.success('Saved.')

    expect(toast.toasts.value).toHaveLength(1)
    expect(toast.toasts.value[0]).toMatchObject({ id, kind: 'success', message: 'Saved.' })
  })

  it('puts the newest on top, where the eye goes', () => {
    const toast = useToast()
    toast.success('first')
    toast.success('second')
    expect(toast.toasts.value.map((t) => t.message)).toEqual(['second', 'first'])
  })

  it('removes a message once its time is up', () => {
    const toast = useToast()
    toast.success('Saved.')

    vi.advanceTimersByTime(3999)
    expect(toast.toasts.value).toHaveLength(1)

    vi.advanceTimersByTime(1)
    expect(toast.toasts.value).toHaveLength(0)
  })

  /** An error is read, not just noticed, so it has to outlast a success. */
  it('keeps an error on screen longer than a success', () => {
    const toast = useToast()
    toast.success('ok')
    toast.error('broken')

    vi.advanceTimersByTime(4000)
    expect(toast.toasts.value.map((t) => t.kind)).toEqual(['error'])
  })

  it('never keeps a message a caller asked to pin', () => {
    const toast = useToast()
    toast.notify('info', 'stays', 0)

    vi.advanceTimersByTime(60_000)
    expect(toast.toasts.value).toHaveLength(1)
  })

  /**
   * A failed bulk save fires one per row. Twenty identical lines down the
   * corner of the screen is noise, so the stack is capped.
   */
  it('caps the stack rather than covering the page', () => {
    const toast = useToast()
    for (let i = 0; i < 10; i++) toast.error(`failure ${i}`)

    expect(toast.toasts.value).toHaveLength(4)
    expect(toast.toasts.value[0]!.message).toBe('failure 9')
  })

  it('dismisses exactly the one asked for', () => {
    const toast = useToast()
    const keep = toast.success('keep')
    const drop = toast.success('drop')

    toast.dismiss(drop)
    expect(toast.toasts.value.map((t) => t.id)).toEqual([keep])
  })
})

describe('UiToasts', () => {
  beforeEach(() => {
    useToast().toasts.value = []
    document.body.innerHTML = ''
  })

  it('renders nothing visible until something is reported', async () => {
    await mountSuspended(UiToasts)
    expect(document.body.textContent?.trim()).toBe('')
  })

  it('shows the message', async () => {
    const toast = useToast()
    await mountSuspended(UiToasts)
    toast.success('Deleted “Bubble White”.')
    await nextTick()

    expect(document.body.textContent).toContain('Deleted “Bubble White”.')
  })

  /**
   * Two regions, not one: a success waits its turn, an error interrupts. A
   * single region would mean announcing failures late or everything rudely.
   */
  it('announces errors assertively and everything else politely', async () => {
    const toast = useToast()
    await mountSuspended(UiToasts)

    const assertive = document.querySelector('[aria-live="assertive"]')
    const polite = document.querySelector('[aria-live="polite"]')
    expect(assertive).not.toBeNull()
    expect(polite).not.toBeNull()

    toast.error('It broke.')
    toast.success('It worked.')
    await nextTick()

    expect(assertive!.textContent).toContain('It broke.')
    expect(polite!.textContent).toContain('It worked.')
  })

  it('can be dismissed by hand', async () => {
    const toast = useToast()
    await mountSuspended(UiToasts)
    toast.success('Saved.')
    await nextTick()

    const button = document.querySelector<HTMLButtonElement>('[aria-label="Dismiss"]')
    expect(button).not.toBeNull()
    button!.click()
    await nextTick()

    expect(toast.toasts.value).toHaveLength(0)
  })
})
