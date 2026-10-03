<script setup lang="ts">
useReveal()
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <a
      href="#main"
      class="sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-3 focus:z-[60] focus:rounded-md focus:bg-accent-solid focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-accent-fg"
    >
      Skip to content
    </a>

    <LayoutHeader />

    <main id="main" class="flex-1">
      <slot />
    </main>
  </div>
</template>
