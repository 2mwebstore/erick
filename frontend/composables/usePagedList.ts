/**
 * Pagination for the admin lists.
 *
 * Two entry points, because the lists are backed two different ways:
 *
 *   usePagedList — the content lists (projects, experience, skills, services,
 *   users) arrive in one payload, so they are sliced in the browser.
 *
 *   usePager — messages and the audit log are paged by the API, which already
 *   takes `limit`/`offset` and returns a filtered total. Those lists can grow
 *   without bound, so they must never be fetched whole.
 */

export const DEFAULT_PAGE_SIZE = 10
export const PAGE_SIZES = [10, 25, 50, 100]

export interface Pager {
  page: Ref<number>
  pageSize: Ref<number>
  total: Ref<number>
  totalPages: ComputedRef<number>
  offset: ComputedRef<number>
  clamp: () => boolean
  reset: () => void
}

export function usePager(initialSize: number = DEFAULT_PAGE_SIZE): Pager {
  const page = ref(1)
  const pageSize = ref(initialSize)
  const total = ref(0)

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const offset = computed(() => (page.value - 1) * pageSize.value)

  /**
   * Pulls the page number back inside the list.
   *
   * Deleting the last row on the last page would otherwise strand the reader on
   * an empty page. Returns whether it moved, so a server-backed list knows it
   * has to refetch.
   */
  function clamp(): boolean {
    const next = Math.min(Math.max(1, page.value), totalPages.value)
    if (next === page.value) return false
    page.value = next
    return true
  }

  // A different page size makes the current page number meaningless, so go back
  // to the top rather than to a page that may no longer exist.
  watch(pageSize, () => {
    page.value = 1
  })

  return { page, pageSize, total, totalPages, offset, clamp, reset: () => (page.value = 1) }
}

export function usePagedList<T>(
  source: MaybeRefOrGetter<T[]>,
  initialSize: number = DEFAULT_PAGE_SIZE,
) {
  const pager = usePager(initialSize)

  // Tracks length rather than identity: rows are sometimes replaced in place
  // after a save, which leaves the array the same object.
  watchEffect(() => {
    pager.total.value = toValue(source).length
  })

  watch(pager.totalPages, () => pager.clamp(), { immediate: true })

  const items = computed(() => {
    const all = toValue(source)
    return all.slice(pager.offset.value, pager.offset.value + pager.pageSize.value)
  })

  return { ...pager, items }
}
