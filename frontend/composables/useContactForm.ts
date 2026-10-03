import type { ContactPayload, ContactResponse } from '~/types'
import type { ContactErrors } from '~/utils/validation'

type Status = 'idle' | 'submitting' | 'success' | 'error'

export function useContactForm() {
  const form = reactive<ContactPayload>({
    name: '',
    email: '',
    phone: '',
    subject: '',
    projectType: '',
    message: '',
    company: '', // honeypot — real visitors never see or fill this
  })

  const errors = ref<ContactErrors>({})
  const status = ref<Status>('idle')
  const feedback = ref('')

  const messageLength = computed(() => form.message.trim().length)

  function clearError(field: keyof ContactPayload) {
    if (errors.value[field]) errors.value = { ...errors.value, [field]: undefined }
  }

  async function submit() {
    if (status.value === 'submitting') return

    const found = validateContact(form)
    errors.value = found

    if (Object.keys(found).some((k) => found[k as keyof ContactErrors])) {
      status.value = 'error'
      feedback.value = 'Please correct the highlighted fields.'
      // Move focus to the first invalid control for keyboard and screen-reader users.
      if (import.meta.client) {
        const first = Object.keys(found)[0]
        nextTick(() => document.getElementById(`contact-${first}`)?.focus())
      }
      return
    }

    status.value = 'submitting'
    feedback.value = ''

    try {
      const res = await $fetch<ContactResponse>('/api/contact', {
        method: 'POST',
        body: {
          ...form,
          name: form.name.trim(),
          email: form.email.trim(),
          phone: form.phone?.trim() ?? '',
          subject: form.subject?.trim() ?? '',
          message: form.message.trim(),
        },
      })

      status.value = 'success'
      feedback.value = res.message
      form.name = ''
      form.email = ''
      form.phone = ''
      form.subject = ''
      form.projectType = ''
      form.message = ''
    } catch (error) {
      const data = (error as { data?: ContactResponse })?.data
      status.value = 'error'
      errors.value = (data?.errors ?? {}) as ContactErrors
      feedback.value =
        data?.message ?? 'Something went wrong sending your message. Please email me directly instead.'
    }
  }

  return { form, errors, status, feedback, messageLength, submit, clearError }
}
