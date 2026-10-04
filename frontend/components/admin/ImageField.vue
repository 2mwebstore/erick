<script setup lang="ts">
import type { ApiError } from '~/composables/useAdmin'

/**
 * An image field that takes either an upload or a link.
 *
 * Uploading stores the file in R2 through the API and puts its public URL in
 * the field; pasting a link (or a path to a file in public/) works the same as
 * before. Either way the field holds one address, so nothing that displays the
 * image has to know which it was.
 */
const props = withDefaults(
  defineProps<{
    id: string
    label: string
    modelValue: string | null | undefined
    /** Where uploads go in the bucket. */
    folder: 'projects' | 'portraits' | 'profiles'
    error?: string
    /** Logos keep their shape; photos fill the preview. */
    fit?: 'cover' | 'contain'
  }>(),
  { error: undefined, fit: 'cover' },
)

const emit = defineEmits<{ 'update:modelValue': [string] }>()

const { api } = useAdmin()
const uploads = useUploadSettings()

const value = computed({
  get: () => props.modelValue ?? '',
  set: (v: string) => emit('update:modelValue', v.trim()),
})

const uploading = ref(false)
const uploadError = ref('')
const previewFailed = ref(false)
watch(
  () => props.modelValue,
  () => {
    previewFailed.value = false
  },
)

const fileInput = ref<HTMLInputElement | null>(null)

const maxLabel = computed(() => {
  const max = uploads.value.maxBytes ?? 0
  return max >= 1 << 20 ? `${Math.round((max / (1 << 20)) * 10) / 10} MB` : `${Math.ceil(max / 1024)} KB`
})

const accept = computed(() => (uploads.value.types ?? []).join(','))

const hint = computed(() =>
  uploads.value.enabled
    ? `Upload a JPEG, PNG, WebP, GIF or AVIF image up to ${maxLabel.value} — it is stored in R2 — or paste an image link. ` +
      'When you save, an uploaded image you replaced or cleared is deleted from R2, unless something else uses it.'
    : 'Paste an image link, or a path to a file in public/ such as /portrait.jpg.',
)

/**
 * The last file uploaded through this field, until something replaces it.
 *
 * Saving the form releases whatever image it replaced, but a file uploaded and
 * then replaced or cleared before any save was never saved anywhere, so no
 * save would release it. This field releases it itself. The API deletes only a
 * file nothing uses, so if it was saved in the meantime it is kept.
 */
const lastUpload = ref('')

function discard(url: string) {
  if (!url) return
  // Fire and forget: a file left behind costs storage, not correctness.
  api('/api/admin/uploads/discard', { method: 'POST', body: { url } }).catch(() => {})
}

function clear() {
  if (value.value && value.value === lastUpload.value) {
    discard(lastUpload.value)
    lastUpload.value = ''
  }
  value.value = ''
}

const shownError = computed(() => props.error || uploadError.value)

const describedBy = computed(() => (shownError.value ? `${props.id}-error` : `${props.id}-hint`))

async function upload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  // Cleared so choosing the same file again still fires a change.
  input.value = ''
  if (!file) return

  uploadError.value = ''
  const { maxBytes, types } = uploads.value
  if (maxBytes && file.size > maxBytes) {
    uploadError.value = `Images can be up to ${maxLabel.value}.`
    return
  }
  // The API decides from the file's bytes; this only saves a pointless upload.
  if (types?.length && file.type && !types.includes(file.type)) {
    uploadError.value = 'Only JPEG, PNG, WebP, GIF or AVIF images can be uploaded.'
    return
  }

  const form = new FormData()
  form.append('folder', props.folder)
  form.append('file', file)

  uploading.value = true
  try {
    const res = await api<{ url: string }>('/api/admin/uploads/image', { method: 'POST', body: form })
    if (lastUpload.value && props.modelValue === lastUpload.value) discard(lastUpload.value)
    lastUpload.value = res.url
    emit('update:modelValue', res.url)
  } catch (e) {
    const err = e as ApiError
    uploadError.value = err.errors?.file ?? err.errors?.folder ?? err.message
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div>
    <label :for="id" class="block text-[0.8125rem] font-medium text-fg">{{ label }}</label>

    <div class="mt-1.5 flex items-start gap-3">
      <!-- Preview: the fastest way to see that an address is wrong. -->
      <span
        class="inline-flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-surface"
      >
        <img
          v-if="value && !previewFailed"
          :src="value"
          alt=""
          :class="['size-full', fit === 'contain' ? 'object-contain p-2' : 'object-cover']"
          loading="lazy"
          @error="previewFailed = true"
        >
        <Icon
          v-else
          :name="previewFailed ? 'lucide:image-off' : 'lucide:image'"
          :class="['size-4', previewFailed ? 'text-red-500' : 'text-fg-subtle']"
          aria-hidden="true"
        />
      </span>

      <div class="min-w-0 flex-1">
        <div class="flex gap-2">
          <input
            :id="id"
            v-model="value"
            type="text"
            inputmode="url"
            placeholder="https://… or /image.png"
            :aria-invalid="shownError ? 'true' : undefined"
            :aria-describedby="describedBy"
            :class="[
              'w-full min-w-0 rounded-md border bg-bg px-3 py-2 text-sm text-fg placeholder:text-fg-subtle',
              'transition-colors focus:outline-none focus-visible:border-accent',
              shownError ? 'border-red-500 dark:border-red-400' : 'border-border-strong hover:border-fg-subtle',
            ]"
          >
          <template v-if="uploads.enabled">
            <input
              ref="fileInput"
              type="file"
              :accept="accept"
              class="sr-only"
              tabindex="-1"
              aria-hidden="true"
              @change="upload"
            >
            <UiButton
              variant="secondary"
              size="sm"
              :disabled="uploading"
              :aria-label="`Upload an image for ${label}`"
              @click="fileInput?.click()"
            >
              <Icon
                :name="uploading ? 'lucide:loader-circle' : 'lucide:upload'"
                :class="['size-3.5', uploading && 'animate-spin']"
                aria-hidden="true"
              />
              {{ uploading ? 'Uploading…' : 'Upload' }}
            </UiButton>
          </template>
          <UiButton
            v-if="value"
            variant="ghost"
            size="sm"
            :aria-label="`Clear ${label}`"
            @click="clear"
          >
            <Icon name="lucide:x" class="size-3.5" aria-hidden="true" />
          </UiButton>
        </div>

        <p v-if="shownError" :id="`${id}-error`" role="alert" class="mt-1 text-xs text-red-600 dark:text-red-400">
          {{ shownError }}
        </p>
        <p v-else :id="`${id}-hint`" class="mt-1 text-xs text-fg-subtle">
          <template v-if="previewFailed">This address does not load as an image. </template>{{ hint }}
        </p>
      </div>
    </div>
  </div>
</template>
