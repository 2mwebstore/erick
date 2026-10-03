/**
 * Marks the navigation item for the section currently in view (§9).
 *
 * Uses IntersectionObserver rather than scroll maths so it costs nothing while
 * the page is idle.
 */
export function useScrollSpy(ids: string[], options: { rootMargin?: string } = {}) {
  const active = ref<string>('')

  onMounted(() => {
    const elements = ids
      .map((id) => document.getElementById(id))
      .filter((el): el is HTMLElement => el !== null)

    if (!elements.length || !('IntersectionObserver' in window)) return

    /**
     * Resolve ties by position on the page, not by position in the nav.
     *
     * Two sections can straddle the observation band at once, and the one to
     * highlight is whichever is higher up the document. Reading that from the
     * DOM rather than from the order of `ids` means the nav can be reordered,
     * or a section added to the page but not the nav, without the highlight
     * silently picking the wrong one.
     */
    const order = new Map(
      [...elements]
        .sort((a, b) => (a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING ? -1 : 1))
        .map((el, index) => [el.id, index]),
    )

    const visible = new Set<string>()

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) visible.add(entry.target.id)
          else visible.delete(entry.target.id)
        }

        let best = ''
        for (const id of visible) {
          if (!best || (order.get(id) ?? 0) < (order.get(best) ?? 0)) best = id
        }

        // Assign even when nothing is in view. The previous version only ever
        // assigned a non-empty value, so the highlight stuck to the last
        // section seen: scrolling back up to the hero left "About" lit, and
        // any stretch of page the observer did not cover kept whichever item
        // happened to be active when it was entered.
        active.value = best
      },
      {
        rootMargin: options.rootMargin ?? '-88px 0px -55% 0px',
        threshold: 0,
      },
    )

    elements.forEach((el) => observer.observe(el))
    onBeforeUnmount(() => observer.disconnect())
  })

  return { active }
}
