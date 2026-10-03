<script setup lang="ts">
/**
 * Pager for the admin lists.
 *
 * Shows the range even when there is only one page — "8 of 8" is information —
 * but hides the page buttons, which would do nothing.
 */
const props = withDefaults(
  defineProps<{
    page: number
    pageSize: number
    total: number
    /** Plural noun for the range line, e.g. "projects". */
    label?: string
    sizes?: number[]
  }>(),
  { label: 'items', sizes: () => PAGE_SIZES },
)

const emit = defineEmits<{ 'update:page': [number]; 'update:pageSize': [number] }>()

const selectId = useId()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const current = computed(() => Math.min(Math.max(props.page, 1), totalPages.value))

const first = computed(() => (props.total === 0 ? 0 : (current.value - 1) * props.pageSize + 1))
const last = computed(() => Math.min(current.value * props.pageSize, props.total))

/** 1 … 4 5 6 … 12 — always the ends, always the neighbours. */
const steps = computed<(number | 'gap')[]>(() => {
  const n = totalPages.value
  if (n <= 7) return Array.from({ length: n }, (_, i) => i + 1)

  const wanted = new Set<number>([1, n])
  for (let i = current.value - 1; i <= current.value + 1; i++) {
    if (i >= 1 && i <= n) wanted.add(i)
  }

  const out: (number | 'gap')[] = []
  let previous = 0
  for (const value of [...wanted].sort((a, b) => a - b)) {
    if (previous && value - previous > 1) out.push('gap')
    out.push(value)
    previous = value
  }
  return out
})

function go(page: number) {
  const next = Math.min(Math.max(page, 1), totalPages.value)
  if (next !== props.page) emit('update:page', next)
}

const stepClass =
  'inline-flex h-8 min-w-8 items-center justify-center rounded-md border px-2 text-xs font-medium tabular-nums transition-colors disabled:pointer-events-none disabled:opacity-40'
</script>

<template>
  <div
    v-if="total > 0"
    class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 border-t border-border pt-4"
  >
    <p class="text-xs text-fg-subtle tabular-nums">
      Showing {{ first }}–{{ last }} of {{ total }} {{ label }}
    </p>

    <div class="flex flex-wrap items-center gap-x-4 gap-y-2">
      <div class="flex items-center gap-2">
        <label :for="selectId" class="text-xs text-fg-subtle">Per page</label>
        <select
          :id="selectId"
          class="h-8 rounded-md border border-border bg-bg px-2 text-xs text-fg"
          :value="pageSize"
          @change="emit('update:pageSize', Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="size in sizes" :key="size" :value="size">{{ size }}</option>
        </select>
      </div>

      <nav v-if="totalPages > 1" aria-label="Pagination" class="flex items-center gap-1">
        <button
          type="button"
          :class="[stepClass, 'border-border text-fg-muted hover:text-fg']"
          :disabled="current === 1"
          aria-label="Previous page"
          @click="go(current - 1)"
        >
          <Icon name="lucide:chevron-left" class="size-4" aria-hidden="true" />
        </button>

        <template v-for="(step, i) in steps" :key="`${step}-${i}`">
          <span v-if="step === 'gap'" class="px-1 text-xs text-fg-subtle" aria-hidden="true">…</span>
          <button
            v-else
            type="button"
            :class="[
              stepClass,
              step === current
                ? 'border-accent-solid bg-accent-solid text-accent-fg'
                : 'border-border text-fg-muted hover:text-fg',
            ]"
            :aria-label="`Page ${step}`"
            :aria-current="step === current ? 'page' : undefined"
            @click="go(step)"
          >
            {{ step }}
          </button>
        </template>

        <button
          type="button"
          :class="[stepClass, 'border-border text-fg-muted hover:text-fg']"
          :disabled="current === totalPages"
          aria-label="Next page"
          @click="go(current + 1)"
        >
          <Icon name="lucide:chevron-right" class="size-4" aria-hidden="true" />
        </button>
      </nav>
    </div>
  </div>
</template>
