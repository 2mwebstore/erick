// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminRichTextEditor from '~/components/admin/RichTextEditor.vue'
import UiProse from '~/components/ui/Prose.vue'

const mount = (modelValue = '') =>
  mountSuspended(AdminRichTextEditor, { props: { id: 'f', label: 'Description', modelValue } })

describe('AdminRichTextEditor', () => {
  it('edits the raw value, so what is typed is what is stored', async () => {
    const wrapper = await mount('Plain text.')
    expect((wrapper.get('textarea').element as HTMLTextAreaElement).value).toBe('Plain text.')
  })

  it('offers the formatting it can actually render', async () => {
    const wrapper = await mount()
    const labels = wrapper.findAll('button').map((b) => b.attributes('aria-label') ?? b.text())
    for (const action of ['Bold', 'Italic', 'Code', 'Link', 'Bulleted list', 'Numbered list']) {
      expect(labels).toContain(action)
    }
  })

  it('wraps the selection rather than replacing it', async () => {
    const wrapper = await mount('make this bold')
    const el = wrapper.get('textarea').element as HTMLTextAreaElement
    el.setSelectionRange(10, 14)

    await wrapper.get('[aria-label="Bold"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('make this **bold**')
  })

  /** With nothing selected, a placeholder goes in so the author can type over it. */
  it('inserts a placeholder when nothing is selected', async () => {
    const wrapper = await mount('')
    await wrapper.get('[aria-label="Bold"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('**bold text**')
  })

  it('prefixes lines for a list instead of wrapping them', async () => {
    const wrapper = await mount('one\ntwo')
    const el = wrapper.get('textarea').element as HTMLTextAreaElement
    el.setSelectionRange(0, 7)

    await wrapper.get('[aria-label="Bulleted list"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('- one\n- two')
  })

  it('numbers a numbered list as it goes', async () => {
    const wrapper = await mount('one\ntwo')
    const el = wrapper.get('textarea').element as HTMLTextAreaElement
    el.setSelectionRange(0, 7)

    await wrapper.get('[aria-label="Numbered list"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('1. one\n2. two')
  })

  it('takes the keyboard shortcut for bold', async () => {
    const wrapper = await mount('word')
    const el = wrapper.get('textarea').element as HTMLTextAreaElement
    el.setSelectionRange(0, 4)

    await wrapper.get('textarea').trigger('keydown', { key: 'b', metaKey: true })
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('**word**')
  })

  it('shows a preview of what the site will render', async () => {
    const wrapper = await mount('**bold**')
    await wrapper.get('[aria-pressed]').trigger('click')
    expect(wrapper.html()).toContain('<strong>bold</strong>')
  })

  /** The preview goes through the same renderer, so it cannot become a hole. */
  it('does not execute markup in the preview', async () => {
    const wrapper = await mount('<img src=x onerror=alert(1)>')
    await wrapper.get('[aria-pressed]').trigger('click')
    expect(wrapper.html()).not.toContain('<img src=x')
  })
})

describe('UiProse with formatting', () => {
  it('renders the supported subset', async () => {
    const wrapper = await mountSuspended(UiProse, { props: { value: 'a **b** c' } })
    expect(wrapper.html()).toContain('<strong>b</strong>')
  })

  it('still shows a pending marker for a placeholder', async () => {
    const wrapper = await mountSuspended(UiProse, { props: { value: 'TODO: write this' } })
    expect(wrapper.find('[data-pending]').exists()).toBe(true)
  })

  it('never renders markup an author typed', async () => {
    const wrapper = await mountSuspended(UiProse, {
      props: { value: '<scr' + 'ipt>alert(1)</scr' + 'ipt>' },
    })
    expect(wrapper.html()).not.toContain('<script>')
  })
})
