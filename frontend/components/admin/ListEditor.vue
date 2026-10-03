<script setup lang="ts">
/** Edits an ordered list of short strings — features, challenges, tags. */
const props = defineProps<{
  label: string
  /** Optional because an unfilled case-study list arrives as undefined. */
  modelValue?: string[]
  placeholder?: string
  hint?: string
  error?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [string[]] }>()

const items = computed(() => props.modelValue ?? [])

function update(index: number, value: string) {
  const next = [...items.value]
  next[index] = value
  emit('update:modelValue', next)
}

function add() {
  emit('update:modelValue', [...items.value, ''])
}

function remove(index: number) {
  emit('update:modelValue', items.value.filter((_, i) => i !== index))
}

function move(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= items.value.length) return
  const next = [...items.value]
  ;[next[index], next[target]] = [next[target]!, next[index]!]
  emit('update:modelValue', next)
}
</script>

<template>
  <fieldset>
    <legend class="text-[0.8125rem] font-medium text-fg">{{ label }}</legend>
    <p v-if="hint" class="mt-1 text-xs text-fg-subtle">{{ hint }}</p>

    <ul class="mt-2 space-y-2">
      <li v-for="(item, i) in items" :key="i" class="flex items-start gap-2">
        <input
          :value="item"
          :placeholder="placeholder"
          :aria-label="`${label} item ${i + 1}`"
          class="w-full rounded-md border border-border-strong bg-bg px-3 py-2 text-sm text-fg transition-colors hover:border-fg-subtle focus:outline-none focus-visible:border-accent"
          @input="update(i, ($event.target as HTMLInputElement).value)"
        >
        <div class="flex shrink-0 gap-1">
          <button
            type="button"
            class="inline-flex size-9 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:border-border-strong hover:text-fg disabled:opacity-40"
            :disabled="i === 0"
            :aria-label="`Move ${label} item ${i + 1} up`"
            @click="move(i, -1)"
          >
            <Icon name="lucide:chevron-up" class="size-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="inline-flex size-9 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:border-border-strong hover:text-fg disabled:opacity-40"
            :disabled="i === items.length - 1"
            :aria-label="`Move ${label} item ${i + 1} down`"
            @click="move(i, 1)"
          >
            <Icon name="lucide:chevron-down" class="size-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="inline-flex size-9 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:border-red-400 hover:text-red-500"
            :aria-label="`Remove ${label} item ${i + 1}`"
            @click="remove(i)"
          >
            <Icon name="lucide:trash-2" class="size-4" aria-hidden="true" />
          </button>
        </div>
      </li>
    </ul>

    <p v-if="error" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ error }}</p>

    <UiButton variant="secondary" size="sm" class="mt-2" @click="add">
      <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
      Add
    </UiButton>
  </fieldset>
</template>
