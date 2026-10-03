<script setup lang="ts">
const props = defineProps<{
  id: string
  label: string
  error?: string
  hint?: string
  required?: boolean
}>()

const describedBy = computed(() =>
  [props.error ? `${props.id}-error` : '', props.hint ? `${props.id}-hint` : '']
    .filter(Boolean)
    .join(' ') || undefined,
)
</script>

<template>
  <div>
    <label :for="id" class="block text-[0.8125rem] font-medium text-fg">
      {{ label }}
      <span v-if="required" class="text-fg-subtle" aria-hidden="true">*</span>
    </label>

    <div class="mt-2">
      <slot :described-by="describedBy" :invalid="Boolean(error)" />
    </div>

    <p v-if="hint && !error" :id="`${id}-hint`" class="mt-1.5 text-xs text-fg-subtle">
      {{ hint }}
    </p>
    <p v-if="error" :id="`${id}-error`" class="mt-1.5 flex items-center gap-1.5 text-xs text-red-600 dark:text-red-400">
      <Icon name="lucide:alert-circle" class="size-3.5 shrink-0" aria-hidden="true" />
      {{ error }}
    </p>
  </div>
</template>
