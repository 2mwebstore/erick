<script setup lang="ts">
/**
 * Language switch: a globe that opens the list of languages.
 *
 * The languages themselves stay real links — each one is its own URL (`/` and
 * `/km/`), so a link is what they actually are: shareable, openable in a new
 * tab, and crawlable. `switchLocalePath` keeps the visitor on the page they
 * were reading instead of sending them back to the home page.
 *
 * The panel is hidden with `v-show` rather than removed, so both addresses are
 * in the markup that a crawler reads whether or not the dropdown is ever
 * opened, and whether or not JavaScript runs at all.
 */
const { locale, locales, t } = useI18n()
const switchLocalePath = useSwitchLocalePath()

type LocaleCode = Parameters<typeof switchLocalePath>[0]

/**
 * A flag is a country, not a language, so it decorates the name rather than
 * replacing it: the trigger is a globe, which says "language" without claiming
 * a country, and the flags appear beside the names inside. Khmer is spoken in
 * Cambodia; English here is labelled with the UK flag, which is a choice, not a
 * fact, and is why the text is what anyone actually reads.
 */
const FLAGS: Record<string, string> = {
  en: 'circle-flags:gb',
  km: 'circle-flags:kh',
}

const options = computed(() =>
  locales.value.map((l) => ({
    code: l.code,
    name: l.name ?? l.code,
    short: l.code === 'en' ? 'EN' : 'ខ្មែរ',
    flag: FLAGS[l.code],
    to: switchLocalePath(l.code as LocaleCode),
  })),
)

const current = computed(() => options.value.find((o) => o.code === locale.value))

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const panel = ref<HTMLElement | null>(null)
const panelId = useId()

async function toggle() {
  open.value = !open.value
  if (!open.value) return
  // Move into the list on open, so the keyboard is already where the choices
  // are and Tab walks them in order.
  await nextTick()
  panel.value?.querySelector('a')?.focus()
}

function close(focusTrigger = false) {
  if (!open.value) return
  open.value = false
  if (focusTrigger) trigger.value?.focus()
}

onMounted(() => {
  // Pointerdown rather than click: a press that starts outside should dismiss
  // the panel immediately, the way every other menu on the platform does.
  const onPointerDown = (event: PointerEvent) => {
    if (!root.value?.contains(event.target as Node)) close()
  }
  const onKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') close(true)
  }
  // Tabbing past the last link takes focus out of the panel, and a panel that
  // stays open behind the focus ring is a panel the keyboard has abandoned.
  const onFocusIn = (event: FocusEvent) => {
    if (!root.value?.contains(event.target as Node)) close()
  }

  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('keydown', onKeydown)
  document.addEventListener('focusin', onFocusIn)

  onBeforeUnmount(() => {
    document.removeEventListener('pointerdown', onPointerDown)
    document.removeEventListener('keydown', onKeydown)
    document.removeEventListener('focusin', onFocusIn)
  })
})

// Choosing a language navigates, and the panel must not survive the page it
// was opened on.
watch(() => useRoute().fullPath, () => (open.value = false))
</script>

<template>
  <div ref="root" class="relative">
    <button
      ref="trigger"
      type="button"
      class="inline-flex h-8 items-center gap-1.5 rounded-md border border-border px-2 text-[0.6875rem] font-medium tracking-tight text-fg-muted transition-colors hover:border-border-strong hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
      :class="open ? 'border-border-strong text-fg' : ''"
      :aria-expanded="open"
      :aria-controls="panelId"
      :aria-label="`${t('nav.language')}: ${current?.name ?? locale}`"
      @click="toggle"
    >
      <Icon name="lucide:globe" class="size-4 shrink-0" aria-hidden="true" />
      <span class="hidden sm:inline" aria-hidden="true">{{ current?.short }}</span>
      <Icon
        name="lucide:chevron-down"
        class="size-3 shrink-0 transition-transform duration-200"
        :class="open ? 'rotate-180' : ''"
        aria-hidden="true"
      />
    </button>

    <div
      v-show="open"
      :id="panelId"
      ref="panel"
      class="absolute right-0 top-full z-50 mt-2 min-w-[10.5rem] overflow-hidden rounded-lg border border-border bg-surface p-1 shadow-lg shadow-black/5 dark:shadow-black/40"
    >
      <NuxtLink
        v-for="option in options"
        :key="option.code"
        :to="option.to"
        class="flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[0.8125rem] transition-colors"
        :class="
          option.code === locale
            ? 'bg-bg font-medium text-fg'
            : 'text-fg-muted hover:bg-bg hover:text-fg'
        "
        :lang="option.code"
        :aria-current="option.code === locale ? 'true' : undefined"
        @click="close()"
      >
        <Icon v-if="option.flag" :name="option.flag" class="size-4 shrink-0" aria-hidden="true" />
        <span class="flex-1">{{ option.name }}</span>
        <Icon
          v-if="option.code === locale"
          name="lucide:check"
          class="size-3.5 shrink-0 text-accent"
          aria-hidden="true"
        />
      </NuxtLink>
    </div>
  </div>
</template>
