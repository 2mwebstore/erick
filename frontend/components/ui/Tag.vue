<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    label: string
    /** Set false on dense lists where glyphs would read as a logo wall (§3). */
    icon?: boolean
    /** Set false to keep the mark monochrome. */
    color?: boolean
  }>(),
  { icon: true, color: true },
)

const meta = computed(() => (props.icon ? techIconMeta(props.label) : undefined))

/**
 * Brand colours are passed as custom properties rather than a direct `color`,
 * so the dark-mode override in main.css can pick the lighter variant without
 * needing a second inline style.
 */
const colorVars = computed(() => {
  if (!props.color || !meta.value?.color) return undefined
  return {
    '--tc': meta.value.color,
    '--tc-dark': meta.value.dark ?? meta.value.color,
  }
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded border border-border bg-surface px-2 py-0.5 font-mono text-[0.6875rem] tracking-tight text-fg-muted"
  >
    <Icon
      v-if="meta"
      :name="meta.icon"
      :class="['size-3 shrink-0', colorVars ? 'tech-icon' : 'text-accent opacity-80']"
      :style="colorVars"
      aria-hidden="true"
    />
    {{ label }}
  </span>
</template>
