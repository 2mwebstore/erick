<script setup lang="ts">
/** Inline save feedback, announced politely so it is not missed. */
defineProps<{ state: 'idle' | 'saving' | 'saved' | 'error'; message?: string }>()
</script>

<template>
  <p aria-live="polite" role="status" class="text-xs">
    <span v-if="state === 'saving'" class="inline-flex items-center gap-1.5 text-fg-muted">
      <Icon name="lucide:loader-circle" class="size-3.5 animate-spin" aria-hidden="true" />
      Saving…
    </span>
    <span v-else-if="state === 'saved'" class="inline-flex items-center gap-1.5 text-fg">
      <Icon name="lucide:check" class="size-3.5 text-accent" aria-hidden="true" />
      {{ message || 'Saved.' }}
    </span>
    <span v-else-if="state === 'error'" class="inline-flex items-center gap-1.5 text-red-600 dark:text-red-400">
      <Icon name="lucide:alert-circle" class="size-3.5" aria-hidden="true" />
      {{ message || 'Could not save.' }}
    </span>
  </p>
</template>
