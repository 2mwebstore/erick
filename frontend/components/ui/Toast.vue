<script setup lang="ts">
import type { Toast } from '~/composables/useToast'

defineProps<{ toast: Toast; icon: string; classes: string }>()
const emit = defineEmits<{ dismiss: [] }>()
</script>

<template>
  <!--
    `pointer-events-auto` because the stack around it is pass-through: a column
    of invisible boxes down the corner of the page would otherwise swallow
    clicks meant for the content underneath.
  -->
  <div
    class="toast pointer-events-auto flex items-start gap-3 rounded-lg border p-3.5 shadow-lg backdrop-blur-sm"
    :class="classes"
  >
    <Icon
      :name="icon"
      class="mt-0.5 size-4 shrink-0"
      :class="toast.kind === 'error' ? 'text-red-500' : 'text-accent'"
      aria-hidden="true"
    />
    <p class="min-w-0 flex-1 text-sm leading-relaxed">{{ toast.message }}</p>
    <button
      type="button"
      class="-m-1 shrink-0 rounded p-1 text-fg-subtle transition-colors hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
      aria-label="Dismiss"
      @click="emit('dismiss')"
    >
      <Icon name="lucide:x" class="size-3.5" aria-hidden="true" />
    </button>
  </div>
</template>

<style scoped>
.toast {
  animation: toast-in 220ms cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateY(-6px) scale(0.98);
  }
}

/* Arriving without moving still reads as arriving. */
@media (prefers-reduced-motion: reduce) {
  .toast {
    animation: none;
  }
}
</style>
