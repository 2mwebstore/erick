<script setup lang="ts">
const { t } = useI18n()
/**
 * Visible marker for content that has not been written yet (§27).
 *
 * Never invents prose to fill the gap. `hint` is the author's own note from the
 * content file, shown so the gap is self-documenting while editing.
 */
const props = defineProps<{ value: string; file?: string }>()
const hint = computed(() => pendingHint(props.value))
</script>

<template>
  <p
    data-pending
    class="flex flex-col gap-1 rounded-md border border-dashed border-border-strong bg-surface px-4 py-3 text-sm text-fg-subtle"
  >
    <span class="font-mono text-[0.6875rem] tracking-[0.14em] uppercase">{{ t('content.pending') }}</span>
    <span>{{ hint }}</span>
    <span v-if="props.file" class="font-mono text-[0.6875rem]">{{ props.file }}</span>
  </p>
</template>
