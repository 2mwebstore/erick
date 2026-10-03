// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { nextTick, ref } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { DEFAULT_PAGE_SIZE, usePagedList, usePager } from '~/composables/usePagedList'
import AdminPagination from '~/components/admin/Pagination.vue'
import AdminConfirmDelete from '~/components/admin/ConfirmDelete.vue'

const rows = (n: number) => Array.from({ length: n }, (_, i) => ({ id: i + 1 }))

describe('usePager', () => {
  it('defaults to ten per page', () => {
    expect(DEFAULT_PAGE_SIZE).toBe(10)
    expect(usePager().pageSize.value).toBe(10)
  })

  it('reports a single page for an empty list rather than zero', () => {
    const pager = usePager()
    expect(pager.totalPages.value).toBe(1)
    expect(pager.offset.value).toBe(0)
  })

  it('turns the page number into an offset the API can use', () => {
    const pager = usePager()
    pager.total.value = 95
    pager.page.value = 4
    expect(pager.offset.value).toBe(30)
    expect(pager.totalPages.value).toBe(10)
  })

  it('returns to the first page when the page size changes', async () => {
    const pager = usePager()
    pager.total.value = 95
    pager.page.value = 7
    pager.pageSize.value = 25
    await nextTick()
    expect(pager.page.value).toBe(1)
  })

  it('clamps a page that has fallen off the end, and says that it moved', () => {
    const pager = usePager()
    pager.total.value = 95
    pager.page.value = 10

    pager.total.value = 11 // most of the list was deleted
    expect(pager.clamp()).toBe(true)
    expect(pager.page.value).toBe(2)
    expect(pager.clamp()).toBe(false) // already in range, so nothing to report
  })
})

describe('usePagedList', () => {
  it('slices the source to one page', () => {
    const source = ref(rows(24))
    const list = usePagedList(source)

    expect(list.total.value).toBe(24)
    expect(list.totalPages.value).toBe(3)
    expect(list.items.value).toHaveLength(10)
    expect(list.items.value[0]!.id).toBe(1)
  })

  it('moves the window with the page', async () => {
    const source = ref(rows(24))
    const list = usePagedList(source)

    list.page.value = 3
    await nextTick()
    expect(list.offset.value).toBe(20)
    expect(list.items.value.map((r) => r.id)).toEqual([21, 22, 23, 24])
  })

  it('tracks a list that grows at the top', async () => {
    const source = ref(rows(5))
    const list = usePagedList(source)

    source.value = [{ id: 99 }, ...source.value]
    await nextTick()
    expect(list.total.value).toBe(6)
    expect(list.items.value[0]!.id).toBe(99)
  })

  it('follows the list back down when the last page is emptied', async () => {
    const source = ref(rows(11))
    const list = usePagedList(source)

    list.page.value = 2
    await nextTick()
    expect(list.items.value).toHaveLength(1)

    // Deleting that last row would otherwise strand the reader on a blank page.
    source.value = rows(10)
    await nextTick()
    expect(list.page.value).toBe(1)
    expect(list.items.value).toHaveLength(10)
  })
})

describe('AdminPagination', () => {
  const mount = (props: Record<string, unknown>) =>
    mountSuspended(AdminPagination, { props: { page: 1, pageSize: 10, total: 0, ...props } })

  it('renders nothing for an empty list', async () => {
    const wrapper = await mount({ total: 0 })
    expect(wrapper.text()).toBe('')
  })

  it('states the range in human terms', async () => {
    const wrapper = await mount({ total: 24, page: 2, label: 'projects' })
    expect(wrapper.text()).toContain('Showing 11–20 of 24 projects')
  })

  it('does not run the range past the end of the list', async () => {
    const wrapper = await mount({ total: 24, page: 3 })
    expect(wrapper.text()).toContain('Showing 21–24 of 24')
  })

  it('hides the page buttons when everything fits on one page', async () => {
    const wrapper = await mount({ total: 8 })
    expect(wrapper.find('nav').exists()).toBe(false)
    expect(wrapper.text()).toContain('Showing 1–8 of 8')
  })

  it('marks the current page for assistive technology', async () => {
    const wrapper = await mount({ total: 24, page: 2 })
    const current = wrapper.find('[aria-current="page"]')
    expect(current.text()).toBe('2')
  })

  it('disables the previous control on the first page', async () => {
    const wrapper = await mount({ total: 24, page: 1 })
    const previous = wrapper.get('[aria-label="Previous page"]')
    expect((previous.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('asks for the next page instead of changing it itself', async () => {
    const wrapper = await mount({ total: 24, page: 1 })
    await wrapper.get('[aria-label="Next page"]').trigger('click')
    expect(wrapper.emitted('update:page')).toEqual([[2]])
  })

  it('collapses a long run of pages around the current one', async () => {
    const wrapper = await mount({ total: 500, page: 25 })
    const labels = wrapper.findAll('nav button').map((b) => b.text()).filter(Boolean)
    expect(labels).toEqual(['1', '24', '25', '26', '50'])
    expect(wrapper.text()).toContain('…')
  })

  it('reports a new page size', async () => {
    const wrapper = await mount({ total: 24 })
    const select = wrapper.get('select')
    await select.setValue('25')
    expect(wrapper.emitted('update:pageSize')).toEqual([[25]])
  })
})

describe('AdminConfirmDelete', () => {
  // The dialog teleports to <body> and mounts are not torn down between tests,
  // so clear it out first or a query finds the previous test's dialog.
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  const dialog = () => document.querySelector('[role="dialog"]')

  it('does not open a dialog until it is asked to', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, { props: { label: 'Bubble White' } })
    expect(dialog()).toBeNull()
    expect(wrapper.get('button').attributes('aria-label')).toBe('Delete Bubble White')
  })

  it('names the row in the dialog so the wrong one is not deleted', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, { props: { label: 'Bubble White' } })
    await wrapper.get('button').trigger('click')

    expect(dialog()?.textContent).toContain('Bubble White')
    expect(dialog()?.getAttribute('aria-modal')).toBe('true')
  })

  it('warns that the delete is final', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, { props: { label: 'A' } })
    await wrapper.get('button').trigger('click')
    expect(dialog()?.textContent).toContain('This cannot be undone.')
  })

  it('shows what else a cascading delete takes with it', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, {
      props: { label: 'B', note: 'The case study goes too.' },
    })
    await wrapper.get('button').trigger('click')
    expect(dialog()?.textContent).toContain('The case study goes too.')
  })

  it('does not emit anything when it is dismissed', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, { props: { label: 'Bubble White' } })
    await wrapper.get('button').trigger('click')

    const cancel = [...document.querySelectorAll('[role="dialog"] button')].find(
      (b) => b.textContent?.trim() === 'Cancel',
    ) as HTMLButtonElement
    cancel.click()
    await nextTick()

    expect(wrapper.emitted('confirm')).toBeUndefined()
    expect(dialog()).toBeNull()
  })

  it('emits once when the delete is confirmed, and closes', async () => {
    const wrapper = await mountSuspended(AdminConfirmDelete, { props: { label: 'Bubble White' } })
    await wrapper.get('button').trigger('click')

    const confirm = [...document.querySelectorAll('[role="dialog"] button')].find(
      (b) => b.textContent?.trim() === 'Delete',
    ) as HTMLButtonElement
    confirm.click()
    await nextTick()

    expect(wrapper.emitted('confirm')).toHaveLength(1)
    expect(dialog()).toBeNull()
  })
})
