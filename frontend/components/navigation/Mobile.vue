<script setup lang="ts">
import { navItems } from '~/content/navigation'
const { site } = useSiteContent()

const { t } = useI18n()
const localePath = useLocalePath()
const home = computed(() => localePath('/'))
const anchor = (hash: string) => `${home.value === '/' ? '' : home.value}/#${hash}`

const props = defineProps<{
  open: boolean
  active?: string
  /** Current header height, so the panel sits flush under it in both states. */
  offset: string
}>()
const emit = defineEmits<{ close: []; pick: [section: string] }>()

const panel = ref<HTMLElement | null>(null)

watch(
  () => props.open,
  (open) => {
    if (!import.meta.client) return
    document.documentElement.style.overflow = open ? 'hidden' : ''
    if (open) nextTick(() => panel.value?.querySelector<HTMLElement>('a, button')?.focus())
  },
)

onBeforeUnmount(() => {
  if (import.meta.client) document.documentElement.style.overflow = ''
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
</script>

<template>
  <!--
    Teleported out of <header> deliberately.

    The header gets `backdrop-blur` once the page is scrolled, and an element
    with a backdrop-filter becomes the containing block for its `position:
    fixed` descendants. Nested here, the panel was being laid out inside a
    56px-tall header instead of the viewport: `top` ended up below `bottom`, the
    height clamped to zero, and the menu simply did not appear once you had
    scrolled. At the top of the page there was no filter and it worked, which is
    what made it look intermittent.
  -->
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        id="mobile-menu"
        ref="panel"
        class="fixed inset-x-0 bottom-0 z-40 overflow-y-auto border-t border-border bg-bg px-6 pb-10 lg:hidden"
        :style="{ top: offset }"
        @keydown="onKeydown"
      >
        <nav :aria-label="t('nav.mobile')">
          <ul class="divide-y divide-border">
            <li v-for="item in navItems" :key="item.key">
              <NuxtLink
                :to="anchor(item.hash)"
                class="flex items-center justify-between py-4 text-lg font-medium tracking-tight transition-colors"
                :class="active === item.section ? 'text-accent' : 'text-fg hover:text-accent'"
                @click="emit('pick', item.section), emit('close')"
              >
                {{ t(`nav.${item.key}`) }}
                <Icon name="lucide:arrow-up-right" class="size-4 text-fg-subtle" aria-hidden="true" />
              </NuxtLink>
            </li>
          </ul>
        </nav>

        <div class="mt-8 flex flex-col gap-3">
          <UiButton :to="localePath('/resume')" variant="secondary" block @click="emit('close')">
            {{ t('actions.resume') }}
          </UiButton>
          <UiButton :to="anchor('contact')" block @click="emit('close')">
            {{ t('actions.letsTalk') }}
          </UiButton>
        </div>

        <p class="mt-8 font-mono text-[0.6875rem] tracking-[0.14em] text-fg-subtle uppercase">
          {{ site.email }}
        </p>
      </div>
    </Transition>
  </Teleport>
</template>
