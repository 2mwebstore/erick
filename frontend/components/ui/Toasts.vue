<script setup lang="ts">
/**
 * The toast stack.
 *
 * Two live regions rather than one: a success is polite and waits its turn, an
 * error is assertive and interrupts. Putting both in one region would force a
 * choice between announcing failures late and announcing everything rudely.
 *
 * Teleported to <body> so no positioned or filtered ancestor can contain it —
 * the same trap the mobile menu fell into.
 */
const { toasts, dismiss } = useToast()

const STYLES: Record<string, { icon: string; classes: string }> = {
  success: {
    icon: 'lucide:check-circle-2',
    classes: 'border-accent/35 bg-bg text-fg',
  },
  error: {
    icon: 'lucide:alert-circle',
    classes: 'border-red-500/40 bg-bg text-fg',
  },
  info: {
    icon: 'lucide:info',
    classes: 'border-border-strong bg-bg text-fg',
  },
}

const errors = computed(() => toasts.value.filter((t) => t.kind === 'error'))
const others = computed(() => toasts.value.filter((t) => t.kind !== 'error'))
</script>

<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed top-4 right-4 z-[60] flex w-[min(22rem,calc(100vw-2rem))] flex-col gap-2"
    >
      <!--
        Both regions are always present. A live region that is added to the page
        at the same moment as its content is frequently not announced at all —
        it has to be there first, empty, and then change.
      -->
      <div role="alert" aria-live="assertive" class="contents">
        <UiToast
          v-for="toast in errors"
          :key="toast.id"
          :toast="toast"
          :icon="STYLES[toast.kind]!.icon"
          :classes="STYLES[toast.kind]!.classes"
          @dismiss="dismiss(toast.id)"
        />
      </div>

      <div role="status" aria-live="polite" class="contents">
        <UiToast
          v-for="toast in others"
          :key="toast.id"
          :toast="toast"
          :icon="STYLES[toast.kind]!.icon"
          :classes="STYLES[toast.kind]!.classes"
          @dismiss="dismiss(toast.id)"
        />
      </div>
    </div>
  </Teleport>
</template>
