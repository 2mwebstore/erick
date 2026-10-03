<script setup lang="ts">
/**
 * Placeholder project visual.
 *
 * Generated geometry rather than a stock photo (§11) — it is honest about being
 * a placeholder, stays on-brand in both themes, and costs no network request.
 * Replace it by setting `image` on the project in content/projects.ts.
 */
const props = defineProps<{ slug: string; label?: string }>()

const seed = computed(() => slugSeed(props.slug))

/** Deterministic node positions so a project's visual never changes between builds. */
const nodes = computed(() => {
  const s = seed.value
  return Array.from({ length: 5 }, (_, i) => ({
    x: 14 + ((s * (i + 3)) % 68),
    y: 16 + ((s * (i + 7)) % 64),
    r: 1.5 + ((s + i) % 3) * 0.6,
    accent: i === (s % 5),
  }))
})

const links = computed(() => {
  const n = nodes.value
  return n.slice(0, -1).map((node, i) => ({ x1: node.x, y1: node.y, x2: n[i + 1]!.x, y2: n[i + 1]!.y }))
})
</script>

<template>
  <div class="relative overflow-hidden rounded-lg border border-border bg-surface">
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="h-full w-full" aria-hidden="true">
      <defs>
        <pattern :id="`grid-${slug}`" width="6.25" height="6.25" patternUnits="userSpaceOnUse">
          <path d="M6.25 0 L0 0 0 6.25" fill="none" stroke="var(--border)" stroke-width="0.25" />
        </pattern>
      </defs>
      <rect width="100" height="100" :fill="`url(#grid-${slug})`" opacity="0.7" />
      <g stroke="var(--border-strong)" stroke-width="0.4">
        <line v-for="(l, i) in links" :key="i" :x1="l.x1" :y1="l.y1" :x2="l.x2" :y2="l.y2" />
      </g>
      <g>
        <circle
          v-for="(n, i) in nodes"
          :key="i"
          :cx="n.x"
          :cy="n.y"
          :r="n.r"
          :fill="n.accent ? 'var(--accent)' : 'var(--fg-subtle)'"
        />
      </g>
    </svg>

    <p
      v-if="label"
      class="absolute bottom-3 left-3 font-mono text-[0.625rem] tracking-[0.14em] text-fg-subtle uppercase"
    >
      {{ label }}
    </p>
  </div>
</template>
