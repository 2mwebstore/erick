<script setup lang="ts">
import { navItems, sectionIds } from '~/content/navigation'
const { site } = useSiteContent()

const { t } = useI18n()
const localePath = useLocalePath()

/** Anchors hang off whichever language's home page is current. */
const home = computed(() => localePath('/'))
const anchor = useSectionLink()

const route = useRoute()
const isHome = computed(() => route.path === '/')

const { active } = useScrollSpy(sectionIds)
const scrolled = ref(false)
const menuOpen = ref(false)

// The header shrinks on scroll, so the drawer below it has to move with it.
// One source of truth: a hard-coded offset in the drawer matched neither state
// and left a gap or an overlap.
const HEIGHT_TOP = '4rem' // h-16
const HEIGHT_SCROLLED = '3.5rem' // h-14
const headerHeight = computed(() => (scrolled.value ? HEIGHT_SCROLLED : HEIGHT_TOP))

onMounted(() => {
  let frame = 0
  const onScroll = () => {
    if (frame) return
    frame = requestAnimationFrame(() => {
      scrolled.value = window.scrollY > 16
      frame = 0
    })
  }
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })

  // The drawer is hidden by an `lg:hidden` class, which does not run any of its
  // cleanup. Rotating a phone to landscape with the menu open used to hide the
  // panel while leaving the scroll lock on `<html>` in place, so the page could
  // not be scrolled and nothing on screen explained why.
  //
  // 1024px, matching `lg:` in the template: five navigation items, the language
  // switcher, the theme button and two calls to action do not fit in 768, and
  // what fell off the right-hand edge there was the Resume button.
  const wide = window.matchMedia('(min-width: 1024px)')
  const onBreakpoint = (e: MediaQueryListEvent | MediaQueryList) => {
    if (e.matches) menuOpen.value = false
  }
  onBreakpoint(wide)
  wide.addEventListener('change', onBreakpoint)

  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onScroll)
    wide.removeEventListener('change', onBreakpoint)
    if (frame) cancelAnimationFrame(frame)
  })
})

// Close the drawer on navigation so a hash link never leaves it open.
watch(() => route.fullPath, () => (menuOpen.value = false))

/**
 * A clicked item lights up straight away.
 *
 * The scroll spy is authoritative, but it only reports once the page has
 * actually arrived — with smooth scrolling that is most of a second of the
 * previous item still looking selected. The click wins until the spy agrees, or
 * until it is clear the spy never will: a short section under a tall header may
 * never reach the observation band, and without the timeout the highlight would
 * stay pinned to it for the rest of the visit.
 */
const clicked = ref('')
let clickedTimer: ReturnType<typeof setTimeout> | undefined

function pick(section: string) {
  clicked.value = section
  clearTimeout(clickedTimer)
  clickedTimer = setTimeout(() => (clicked.value = ''), 1200)
}

watch(active, (value) => {
  if (value === clicked.value) {
    clearTimeout(clickedTimer)
    clicked.value = ''
  }
})

onBeforeUnmount(() => clearTimeout(clickedTimer))

/** Only highlight in-page sections while on the homepage. */
const activeSection = computed(() => {
  if (!isHome.value) return ''
  return clicked.value || active.value
})
</script>

<template>
  <header
    class="sticky top-0 z-50 border-b transition-[height,background-color,border-color,backdrop-filter] duration-200"
    :class="
      scrolled
        ? 'h-14 border-border bg-bg/85 backdrop-blur-md supports-[backdrop-filter]:bg-bg/70'
        : 'h-16 border-transparent bg-bg'
    "
  >
    <div class="mx-auto flex h-full max-w-[84rem] items-center justify-between gap-3 px-6 sm:gap-6 lg:px-8">
      <NuxtLink
        :to="home"
        class="min-w-0 truncate text-sm font-semibold tracking-tight text-fg transition-colors hover:text-accent"
      >
        {{ site.name }}
        <span class="sr-only">— {{ site.role }}, {{ t('nav.home') }}</span>
      </NuxtLink>

      <nav :aria-label="t('nav.primary')" class="hidden lg:block">
        <ul class="flex items-center gap-1">
          <li v-for="item in navItems" :key="item.key">
            <NuxtLink
              :to="anchor(item.hash)"
              class="relative rounded-md px-3 py-2 text-[0.8125rem] font-medium transition-colors"
              :class="
                activeSection === item.section
                  ? 'text-fg'
                  : 'text-fg-muted hover:text-fg'
              "
              :aria-current="activeSection === item.section ? 'true' : undefined"
              @click="pick(item.section)"
            >
              {{ t(`nav.${item.key}`) }}
              <span
                v-if="activeSection === item.section"
                class="absolute inset-x-3 -bottom-px h-px bg-accent-solid"
                aria-hidden="true"
              />
            </NuxtLink>
          </li>
        </ul>
      </nav>

      <div class="flex shrink-0 items-center gap-1.5 sm:gap-2">
        <NavigationLanguageSwitcher />
        <NavigationThemeToggle />
        <!--
          Hidden by the wrapper, not by a class on the buttons themselves.
          UiButton always emits `inline-flex`, and a `hidden` passed in from
          here is the same kind of utility at the same specificity, so which one
          wins is decided by the order Tailwind happens to emit them in — it
          emitted `inline-flex` last, and both buttons stayed on screen at every
          width, squeezing the name down to "K..." on a phone.
        -->
        <div class="hidden items-center gap-2 lg:flex">
          <UiButton :to="localePath('/resume')" variant="secondary" size="sm">
            {{ t('actions.resume') }}
          </UiButton>
          <UiButton :to="anchor('contact')" size="sm">
            {{ t('actions.letsTalk') }}
          </UiButton>
        </div>

        <button
          type="button"
          class="inline-flex size-9 items-center justify-center rounded-md border border-transparent text-fg-muted transition-colors hover:border-border hover:bg-surface hover:text-fg lg:hidden"
          :aria-expanded="menuOpen"
          aria-controls="mobile-menu"
          :aria-label="menuOpen ? t('nav.closeMenu') : t('nav.openMenu')"
          @click="menuOpen = !menuOpen"
        >
          <Icon :name="menuOpen ? 'lucide:x' : 'lucide:menu'" class="size-5" aria-hidden="true" />
        </button>
      </div>
    </div>

    <NavigationMobile
      :open="menuOpen"
      :active="activeSection"
      :offset="headerHeight"
      @close="menuOpen = false"
      @pick="pick"
    />
  </header>
</template>
