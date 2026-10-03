<script setup lang="ts">
/**
 * One field in both languages.
 *
 * English is the field itself and is required where the entity requires it.
 * Khmer is an overlay stored separately, and is always optional: leaving it
 * empty means the site shows the English, which is the right default while the
 * translation is still being written.
 *
 * The Khmer box is bound straight into the entity's `translations` object, so
 * the existing single Save button writes both languages in one request.
 */
type Translations = Record<string, Record<string, string>>

const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string | number | null | undefined
    /** Field name as the API knows it, e.g. `title` or `caseStudy.overview`. */
    field: string
    locale?: string
    type?: string
    error?: string
    hint?: string
    required?: boolean
    rows?: number
    multiline?: boolean
    /** Formatting toolbar and preview, for the longer description fields. */
    rich?: boolean
  }>(),
  {
    locale: 'km',
    type: 'text',
    rows: 4,
    error: undefined,
    hint: undefined,
    required: false,
    multiline: false,
    rich: false,
  },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

/**
 * The row's whole translations object, as `v-model:translations`.
 *
 * Writing a new object rather than mutating one reached through a prop: several
 * of these share a row, so the parent has to own the value or they would each
 * be editing a copy.
 */
const translations = defineModel<Translations | undefined>('translations', { default: undefined })

const english = computed({
  get: () => (props.modelValue ?? '') as string,
  set: (value: string) => emit('update:modelValue', value),
})

const translated = computed({
  get: () => translations.value?.[props.locale]?.[props.field] ?? '',
  set: (value: string) => {
    const current = translations.value ?? {}
    translations.value = {
      ...current,
      [props.locale]: { ...current[props.locale], [props.field]: value },
    }
  },
})
</script>

<template>
  <div class="grid gap-3 lg:grid-cols-2">
    <AdminRichTextEditor
      v-if="rich"
      :id="id"
      v-model="english"
      :label="label"
      :rows="rows"
      :error="error"
      :hint="hint"
      :required="required"
    />
    <AdminInput
      v-else
      :id="id"
      v-model="english"
      :label="label"
      :type="type"
      :error="error"
      :hint="hint"
      :required="required"
      :rows="rows"
      :multiline="multiline"
    />

    <AdminRichTextEditor
      v-if="rich"
      :id="`${id}-${locale}`"
      v-model="translated"
      :label="`${label} (ខ្មែរ)`"
      :rows="rows"
      lang="km"
      hint="Optional — empty shows the English."
    />
    <AdminInput
      v-else
      :id="`${id}-${locale}`"
      v-model="translated"
      :label="`${label} (ខ្មែរ)`"
      :type="type"
      :rows="rows"
      :multiline="multiline"
      lang="km"
      hint="Optional — empty shows the English."
    />
  </div>
</template>
