<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <section class="rounded-lg border border-border bg-surface/40">
    <header class="border-b border-border px-5 py-4">
      <h2 class="text-sm font-semibold tracking-tight text-fg">{{ title }}</h2>
      <p v-if="description" class="mt-1 text-xs leading-relaxed text-fg-muted">{{ description }}</p>
    </header>
    <div class="p-5">
      <slot />
    </div>
  </section>
</template>
