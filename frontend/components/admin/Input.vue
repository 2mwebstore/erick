<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string | number | null | undefined
    type?: string
    error?: string
    hint?: string
    placeholder?: string
    /**
     * Passed through to the field itself.
     *
     * Declared rather than left to fall through: the root of this component is
     * the wrapping div, so an undeclared `autocomplete` landed there instead of
     * on the input, where a password manager never saw it. The sign-in form has
     * been passing one since it was written.
     */
    autocomplete?: string
    required?: boolean
    rows?: number
    /** Renders a textarea instead of an input. */
    multiline?: boolean
  }>(),
  {
    type: 'text',
    rows: 4,
    autocomplete: undefined,
    error: undefined,
    hint: undefined,
    placeholder: undefined,
    required: false,
    multiline: false,
  },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const value = computed({
  get: () => (props.modelValue ?? '') as string,
  set: (v: string) => emit('update:modelValue', v),
})

const describedBy = computed(() =>
  [props.error ? `${props.id}-error` : '', props.hint ? `${props.id}-hint` : '']
    .filter(Boolean)
    .join(' ') || undefined,
)

const fieldClass = computed(() => [
  'w-full rounded-md border bg-bg px-3 py-2 text-sm text-fg placeholder:text-fg-subtle',
  'transition-colors focus:outline-none focus-visible:border-accent',
  props.error ? 'border-red-500 dark:border-red-400' : 'border-border-strong hover:border-fg-subtle',
])
</script>

<template>
  <div>
    <label :for="id" class="block text-[0.8125rem] font-medium text-fg">
      {{ label }}
      <span v-if="required" class="text-fg-subtle" aria-hidden="true">*</span>
    </label>

    <textarea
      v-if="multiline"
      :id="id"
      v-model="value"
      :rows="rows"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="[...fieldClass, 'mt-1.5 resize-y leading-relaxed']"
    />
    <input
      v-else
      :id="id"
      v-model="value"
      :type="type"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="[...fieldClass, 'mt-1.5']"
    >

    <p v-if="hint && !error" :id="`${id}-hint`" class="mt-1 text-xs text-fg-subtle">{{ hint }}</p>
    <p v-if="error" :id="`${id}-error`" class="mt-1 text-xs text-red-600 dark:text-red-400">
      {{ error }}
    </p>
  </div>
</template>
