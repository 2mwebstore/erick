<script setup lang="ts">
/**
 * Formatting editor for description fields.
 *
 * Writes Markdown rather than HTML, and stores it as plain text. That keeps the
 * `TODO:` placeholder convention working on the raw value, keeps the CSV export
 * and the seed files readable, and means the site never has to trust stored
 * markup — `renderRichText` escapes everything before it puts any tag back.
 *
 * A textarea with a toolbar, not a contenteditable. Authors keep their
 * selection, their undo history and their keyboard, and the value in the box is
 * exactly the value that is saved; the preview below says what it will look
 * like. A contenteditable would have to be kept in sync with the model on every
 * keystroke and would still be the weaker editor.
 */
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string | null | undefined
    rows?: number
    error?: string
    hint?: string
    required?: boolean
    placeholder?: string
    lang?: string
  }>(),
  {
    rows: 5,
    error: undefined,
    hint: undefined,
    required: false,
    placeholder: undefined,
    lang: undefined,
  },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const field = ref<HTMLTextAreaElement | null>(null)
const showPreview = ref(false)

const value = computed({
  get: () => (props.modelValue ?? '') as string,
  set: (next: string) => emit('update:modelValue', next),
})

const preview = computed(() => renderRichText(value.value))

const describedBy = computed(
  () =>
    [props.error ? `${props.id}-error` : '', props.hint ? `${props.id}-hint` : '']
      .filter(Boolean)
      .join(' ') || undefined,
)

interface Action {
  key: string
  label: string
  icon: string
  /** Wraps the selection, or prefixes each selected line when `linewise`. */
  before: string
  after?: string
  linewise?: boolean
  placeholder: string
}

const ACTIONS: Action[] = [
  { key: 'bold', label: 'Bold', icon: 'lucide:bold', before: '**', after: '**', placeholder: 'bold text' },
  { key: 'italic', label: 'Italic', icon: 'lucide:italic', before: '*', after: '*', placeholder: 'italic text' },
  { key: 'code', label: 'Code', icon: 'lucide:code', before: '`', after: '`', placeholder: 'code' },
  { key: 'link', label: 'Link', icon: 'lucide:link', before: '[', after: '](https://)', placeholder: 'link text' },
  { key: 'bullet', label: 'Bulleted list', icon: 'lucide:list', before: '- ', linewise: true, placeholder: 'item' },
  { key: 'number', label: 'Numbered list', icon: 'lucide:list-ordered', before: '1. ', linewise: true, placeholder: 'item' },
]

/**
 * Applies an action to the selection and puts the caret back where the author
 * expects it — inside what they just wrapped, or after it when nothing was
 * selected and a placeholder was inserted for them to type over.
 */
async function apply(action: Action) {
  const el = field.value
  if (!el) return

  const start = el.selectionStart
  const end = el.selectionEnd
  const text = value.value
  const selected = text.slice(start, end)

  let replacement: string
  let caretStart: number
  let caretEnd: number

  if (action.linewise) {
    const lineStart = text.lastIndexOf('\n', start - 1) + 1
    const body = selected || action.placeholder
    replacement = body
      .split('\n')
      .map((line, i) => (action.key === 'number' ? `${i + 1}. ` : action.before) + line)
      .join('\n')

    const prefix = text.slice(lineStart, start)
    value.value = text.slice(0, lineStart) + prefix + replacement + text.slice(end)
    caretStart = lineStart + prefix.length + replacement.length
    caretEnd = caretStart
  } else {
    const body = selected || action.placeholder
    replacement = action.before + body + (action.after ?? '')
    value.value = text.slice(0, start) + replacement + text.slice(end)

    // Select the body, so typing replaces the placeholder immediately.
    caretStart = start + action.before.length
    caretEnd = caretStart + body.length
  }

  await nextTick()
  el.focus()
  el.setSelectionRange(caretStart, caretEnd)
}

/** Cmd/Ctrl+B and +I, because nobody reaches for a toolbar for those. */
function onKeydown(event: KeyboardEvent) {
  if (!event.metaKey && !event.ctrlKey) return
  const action = ACTIONS.find(
    (a) => (a.key === 'bold' && event.key === 'b') || (a.key === 'italic' && event.key === 'i'),
  )
  if (!action) return
  event.preventDefault()
  apply(action)
}

const fieldClass = computed(() => [
  'w-full rounded-b-md border border-t-0 bg-bg px-3 py-2 font-mono text-sm leading-relaxed text-fg',
  'placeholder:text-fg-subtle transition-colors focus:outline-none focus-visible:border-accent',
  props.error ? 'border-red-500 dark:border-red-400' : 'border-border-strong hover:border-fg-subtle',
])
</script>

<template>
  <div>
    <label :for="id" class="block text-[0.8125rem] font-medium text-fg">
      {{ label }}
      <span v-if="required" class="text-fg-subtle" aria-hidden="true">*</span>
    </label>

    <div class="mt-1.5">
      <div
        class="flex flex-wrap items-center gap-0.5 rounded-t-md border border-border-strong bg-surface px-1.5 py-1"
      >
        <button
          v-for="action in ACTIONS"
          :key="action.key"
          type="button"
          class="inline-flex size-7 items-center justify-center rounded text-fg-muted transition-colors hover:bg-bg hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-accent"
          :title="action.label"
          :aria-label="action.label"
          @click="apply(action)"
        >
          <Icon :name="action.icon" class="size-3.5" aria-hidden="true" />
        </button>

        <button
          type="button"
          class="ml-auto inline-flex h-7 items-center gap-1.5 rounded px-2 text-[0.6875rem] font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-accent"
          :class="showPreview ? 'bg-bg text-fg' : 'text-fg-muted hover:text-fg'"
          :aria-pressed="showPreview"
          @click="showPreview = !showPreview"
        >
          <Icon :name="showPreview ? 'lucide:pencil' : 'lucide:eye'" class="size-3.5" aria-hidden="true" />
          {{ showPreview ? 'Edit' : 'Preview' }}
        </button>
      </div>

      <textarea
        v-show="!showPreview"
        :id="id"
        ref="field"
        v-model="value"
        :rows="rows"
        :lang="lang"
        :placeholder="placeholder"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="describedBy"
        :class="fieldClass"
        @keydown="onKeydown"
      />

      <!--
        Safe to render: the preview runs the same function the site does, which
        escapes the author's text before putting any tag back. Nothing here can
        be markup the author wrote.
      -->
      <!-- eslint-disable vue/no-v-html -->
      <div
        v-show="showPreview"
        class="admin-richtext min-h-[6rem] rounded-b-md border border-t-0 border-border-strong bg-bg px-3 py-2 text-sm leading-relaxed text-fg-muted"
        :lang="lang"
        v-html="preview || '<p class=&quot;opacity-60&quot;>Nothing to preview yet.</p>'"
      />
      <!-- eslint-enable vue/no-v-html -->
    </div>

    <p v-if="hint && !error" :id="`${id}-hint`" class="mt-1 text-xs text-fg-subtle">{{ hint }}</p>
    <p v-if="error" :id="`${id}-error`" class="mt-1 text-xs text-red-600 dark:text-red-400">
      {{ error }}
    </p>
  </div>
</template>

<style scoped>
/* The preview is generated markup, so it is styled here rather than by class. */
.admin-richtext :deep(p + p) {
  margin-top: 0.75em;
}
.admin-richtext :deep(strong) {
  color: var(--fg);
  font-weight: 600;
}
.admin-richtext :deep(code) {
  font-family: var(--font-mono);
  font-size: 0.875em;
  color: var(--fg);
}
.admin-richtext :deep(a) {
  color: var(--accent);
  text-decoration: underline;
}
.admin-richtext :deep(ul),
.admin-richtext :deep(ol) {
  margin: 0.25em 0 0 1.25em;
  list-style: revert;
}
</style>
