type ThemeChoice = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'theme'

/**
 * Dark / light / system theme with a persisted preference (§4).
 *
 * The resolved class is applied to <html> before first paint by the inline
 * script in nuxt.config.ts; this composable keeps it in sync afterwards.
 */
export function useTheme() {
  const choice = useState<ThemeChoice>('theme-choice', () => 'system')
  const systemDark = useState<boolean>('theme-system-dark', () => false)

  const isDark = computed(() =>
    choice.value === 'system' ? systemDark.value : choice.value === 'dark',
  )

  function apply() {
    if (!import.meta.client) return
    const root = document.documentElement
    root.classList.toggle('dark', isDark.value)
    root.style.colorScheme = isDark.value ? 'dark' : 'light'
  }

  function set(next: ThemeChoice) {
    choice.value = next
    if (!import.meta.client) return
    try {
      if (next === 'system') localStorage.removeItem(STORAGE_KEY)
      else localStorage.setItem(STORAGE_KEY, next)
    } catch {
      // Storage can be unavailable (private mode, blocked cookies) — the theme
      // still applies for this page view.
    }
    apply()
  }

  /** Cycles light → dark → system, which keeps the control to one button. */
  function cycle() {
    set(choice.value === 'light' ? 'dark' : choice.value === 'dark' ? 'system' : 'light')
  }

  onMounted(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    systemDark.value = mq.matches

    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      choice.value = stored === 'dark' || stored === 'light' ? stored : 'system'
    } catch {
      choice.value = 'system'
    }

    const onChange = (e: MediaQueryListEvent) => {
      systemDark.value = e.matches
      apply()
    }
    mq.addEventListener('change', onChange)
    onBeforeUnmount(() => mq.removeEventListener('change', onChange))

    apply()
  })

  return { choice, isDark, set, cycle }
}
