<script setup lang="ts">
/**
 * Delete control with its own confirmation step.
 *
 * Replaces window.confirm, which cannot be styled, cannot say what is about to
 * be lost beyond one line of plain text, and is blocked outright in some
 * browsers. Deletes here are not undoable — there is no trash — so the dialog
 * names the row and opens with the focus on Cancel, never on Delete.
 */
withDefaults(
  defineProps<{
    /** The row being deleted, shown in the dialog so it is unambiguous. */
    label: string
    /** What else goes with it, when a delete cascades. */
    note?: string
    /** `icon` for a row of controls, `button` where there is room for a word. */
    variant?: 'icon' | 'button'
    disabled?: boolean
  }>(),
  { note: undefined, variant: 'icon', disabled: false },
)

const emit = defineEmits<{ confirm: [] }>()

const open = ref(false)
const trigger = ref<HTMLButtonElement | null>(null)
const cancelButton = ref<HTMLButtonElement | null>(null)
const confirmButton = ref<HTMLButtonElement | null>(null)
const headingId = useId()

async function show() {
  open.value = true
  await nextTick()
  cancelButton.value?.focus()
}

function close() {
  if (!open.value) return
  open.value = false
  trigger.value?.focus()
}

function accept() {
  open.value = false
  emit('confirm')
  trigger.value?.focus()
}

/**
 * Keeps Tab inside the dialog. With only two controls the whole trap is the two
 * wrap-around cases, but without it Tab walks into the page behind the overlay,
 * which for a modal is a trap of a different kind.
 */
function onTab(event: KeyboardEvent) {
  const forward = !event.shiftKey
  const atEnd = forward && document.activeElement === confirmButton.value
  const atStart = !forward && document.activeElement === cancelButton.value
  if (!atEnd && !atStart) return

  event.preventDefault()
  ;(forward ? cancelButton : confirmButton).value?.focus()
}
</script>

<template>
  <button
    ref="trigger"
    type="button"
    :disabled="disabled"
    :class="
      variant === 'icon'
        ? 'inline-flex size-8 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:border-red-400 hover:text-red-500 disabled:pointer-events-none disabled:opacity-40'
        : 'inline-flex h-9 items-center justify-center gap-2 rounded-md border border-transparent px-3.5 text-[0.8125rem] font-medium text-fg-muted transition-colors hover:bg-surface hover:text-red-500 disabled:pointer-events-none disabled:opacity-40'
    "
    :aria-label="variant === 'icon' ? `Delete ${label}` : undefined"
    @click="show"
  >
    <Icon name="lucide:trash-2" :class="variant === 'icon' ? 'size-4' : 'size-3.5'" aria-hidden="true" />
    <template v-if="variant === 'button'">Delete</template>
  </button>

  <Teleport v-if="open" to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      @keydown.esc="close"
      @keydown.tab="onTab"
    >
      <!-- Inert backdrop: clicking away is a cancel, which is the safe outcome. -->
      <div class="absolute inset-0 bg-black/50" @click="close" />

      <div
        role="dialog"
        aria-modal="true"
        :aria-labelledby="headingId"
        class="relative w-full max-w-sm rounded-lg border border-border bg-bg p-6 shadow-xl"
      >
        <h2 :id="headingId" class="text-base font-semibold tracking-tight text-fg">
          Delete “{{ label }}”?
        </h2>
        <p class="mt-2 text-sm text-fg-muted">
          {{ note ? note : 'This cannot be undone.' }}
        </p>

        <div class="mt-6 flex justify-end gap-2">
          <button
            ref="cancelButton"
            type="button"
            class="inline-flex h-9 items-center rounded-md border border-border-strong px-3.5 text-[0.8125rem] font-medium text-fg transition-colors hover:bg-surface"
            @click="close"
          >
            Cancel
          </button>
          <button
            ref="confirmButton"
            type="button"
            class="inline-flex h-9 items-center rounded-md border border-red-600 bg-red-600 px-3.5 text-[0.8125rem] font-medium text-white transition-colors hover:border-red-700 hover:bg-red-700"
            @click="accept"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
