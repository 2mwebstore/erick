<script setup lang="ts">
withDefaults(
  defineProps<{
    id?: string
    eyebrow?: string
    title?: string
    subtitle?: string
    /** Renders the top hairline divider. */
    divider?: boolean
    /** Wider container for the work grid. */
    wide?: boolean
    as?: string
    /** Use h1 when this section is the page's main heading. */
    titleTag?: 'h1' | 'h2'
    /**
     * Replaces the title's colour, e.g. with the gradient the name is painted
     * in. It replaces rather than adds, because `text-fg` is a utility and
     * would outrank a colour set by any class passed alongside it.
     */
    titleClass?: string
  }>(),
  {
    divider: true,
    as: 'section',
    id: undefined,
    eyebrow: undefined,
    title: undefined,
    subtitle: undefined,
    wide: false,
    titleTag: 'h2',
    titleClass: '',
  },
)
</script>

<template>
  <component :is="as" :id="id" :class="['scroll-mt-24', divider ? 'hairline' : '']">
    <div :class="['mx-auto w-full px-6 py-20 sm:py-24 lg:px-8', wide ? 'max-w-[84rem]' : 'max-w-6xl']">
      <header v-if="eyebrow || title || subtitle" class="reveal max-w-3xl">
        <p v-if="eyebrow" class="eyebrow">{{ eyebrow }}</p>
        <component
          :is="titleTag"
          v-if="title"
          class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl"
          :class="[titleTag === 'h1' ? 'sm:text-5xl sm:tracking-[-0.03em]' : '', titleClass || 'text-fg']"
        >
          {{ title }}
        </component>
        <p v-if="subtitle" class="mt-4 text-base leading-relaxed text-fg-muted sm:text-lg">
          {{ subtitle }}
        </p>
      </header>

      <div :class="eyebrow || title ? 'mt-12 sm:mt-16' : ''">
        <slot />
      </div>
    </div>
  </component>
</template>
