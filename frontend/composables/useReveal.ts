/**
 * One-shot reveal on scroll (§25).
 *
 * Adds `.is-visible` to `.reveal` elements the first time they enter the
 * viewport, then stops observing them. No parallax, no scroll listeners, and a
 * no-op when the visitor prefers reduced motion.
 */
export function useReveal() {
  onMounted(() => {
    if (!('IntersectionObserver' in window)) return
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    const observer = new IntersectionObserver(
      (entries, obs) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue

          entry.target.classList.add('is-visible')
          obs.unobserve(entry.target)

          // Drop the compositor layer once the transition is over. `will-change`
          // left on every revealed element keeps a layer alive for the rest of
          // the visit, which on a long page is a real amount of memory for an
          // animation that has already finished.
          entry.target.addEventListener(
            'transitionend',
            () => entry.target.classList.add('is-settled'),
            { once: true },
          )
        }
      },
      { rootMargin: '0px 0px -10% 0px', threshold: 0.05 },
    )

    const observeAll = () => {
      document.querySelectorAll('.reveal:not(.is-visible)').forEach((el) => observer.observe(el))
    }

    observeAll()
    // Re-scan after navigation adds new sections to the DOM.
    const stop = watch(() => useRoute().fullPath, () => nextTick(observeAll))

    onBeforeUnmount(() => {
      observer.disconnect()
      stop()
    })
  })
}
