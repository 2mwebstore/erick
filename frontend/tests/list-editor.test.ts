// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminListEditor from '~/components/admin/ListEditor.vue'

describe('AdminListEditor', () => {
  const mount = (modelValue: string[]) =>
    mountSuspended(AdminListEditor, { props: { label: 'Paragraphs', modelValue } })

  it('renders one input per item', async () => {
    const wrapper = await mount(['one', 'two'])
    expect(wrapper.findAll('input')).toHaveLength(2)
  })

  /**
   * Add emits a list with a blank row on the end. Any binding that filters
   * blanks on the way in therefore throws the row away before it can be typed
   * into, which is exactly how the Khmer About paragraphs field came to do
   * nothing at all. The contract is pinned here so a future binding cannot
   * quietly reintroduce it.
   */
  it('adds a row by emitting a trailing blank', async () => {
    const wrapper = await mount(['one'])

    const add = wrapper.findAll('button').find((b) => b.text().includes('Add'))
    await add!.trigger('click')

    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toEqual(['one', ''])
  })

  it('emits a new array rather than mutating the one it was given', async () => {
    const original = ['one', 'two']
    const wrapper = await mount(original)

    await wrapper.findAll('input')[0]!.setValue('edited')

    const emitted = wrapper.emitted('update:modelValue')?.[0]?.[0] as string[]
    expect(emitted).toEqual(['edited', 'two'])
    expect(original).toEqual(['one', 'two'])
  })

  it('removes the row asked for', async () => {
    const wrapper = await mount(['one', 'two', 'three'])
    await wrapper.get('[aria-label="Remove Paragraphs item 2"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toEqual(['one', 'three'])
  })

  it('reorders without losing anything', async () => {
    const wrapper = await mount(['one', 'two'])
    await wrapper.get('[aria-label="Move Paragraphs item 2 up"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toEqual(['two', 'one'])
  })

  it('treats an absent list as empty rather than failing', async () => {
    const wrapper = await mountSuspended(AdminListEditor, { props: { label: 'Tools' } })
    expect(wrapper.findAll('input')).toHaveLength(0)
  })
})
