<script setup lang="ts">
/**
 * Your own account, and the one thing anyone needs to do to it.
 *
 * Separate from /admin/users, which is administrators managing other people's
 * accounts: this page is about the account you are signed in with, so every
 * role can reach it. Until now the only way to change a password was
 * `adminctl passwd` on the server, which is why one ended up written down.
 */
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { user, csrf, api } = useAdmin()
const toast = useToast()

/** Matches services/auth.go: length does the work, the character rule is light. */
const MIN_LENGTH = 8
const MAX_BYTES = 72

const form = reactive({ current: '', next: '', confirm: '' })
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

const bytes = (value: string) => new TextEncoder().encode(value).length

/**
 * Checked here as well as on the server, because a password the server will
 * reject is worth saying so about before the current one is typed in vain.
 * The server remains the authority; this only saves a round trip.
 */
function validate(): boolean {
  const found: Record<string, string> = {}

  if (!form.current) found.current = 'Enter your current password.'
  if ([...form.next].length < MIN_LENGTH) found.next = `Use at least ${MIN_LENGTH} characters.`
  // bcrypt truncates silently past 72 bytes, so the server refuses it outright.
  else if (bytes(form.next) > MAX_BYTES) found.next = 'Passwords are limited to 72 bytes.'
  else if (form.next === form.current) found.next = 'That is the password you already have.'
  else if (form.confirm !== form.next) found.confirm = 'This does not match the new password.'

  errors.value = found
  return Object.keys(found).length === 0
}

async function submit() {
  if (state.value === 'saving') return
  message.value = ''

  if (!validate()) {
    state.value = 'error'
    message.value = 'Please check the form.'
    return
  }

  state.value = 'saving'

  try {
    await api('/api/auth/password', {
      method: 'POST',
      body: { current: form.current, next: form.next },
    })

    // The server revoked every session, including this one, so there is nothing
    // left to stay on the page for. Clear what is held here and send them back
    // to sign in with the new password.
    form.current = ''
    form.next = ''
    form.confirm = ''
    state.value = 'saved'
    user.value = null
    csrf.value = ''
    toast.success('Password changed. Please sign in again.')
    await navigateTo('/admin/login')
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    /*
       The server reports a rejected new password under `password`, because the
       same validator is used when an account is created. This form calls that
       field `next`, so the message is moved to the box it is about rather than
       shown against nothing.
    */
    const { password, ...rest } = err.errors ?? {}
    errors.value = password ? { ...rest, next: password } : rest
    message.value = err.message ?? 'Could not change the password.'
    toast.error(message.value)
  }
}

useHead({ title: 'Your account' })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight text-fg">Your account</h1>
      <p class="mt-2 text-sm text-fg-muted">
        Signed in as <span class="font-medium text-fg">{{ user?.email }}</span>
        <span v-if="user?.role"> · {{ user.role }}</span>
      </p>
    </header>

    <AdminPanel
      title="Change password"
      description="Changing it signs you out everywhere, on this device and any other."
    >
      <form class="max-w-md space-y-5" novalidate @submit.prevent="submit">
        <AdminInput
          id="pw-current"
          v-model="form.current"
          label="Current password"
          type="password"
          required
          autocomplete="current-password"
          :error="errors.current"
        />

        <AdminInput
          id="pw-next"
          v-model="form.next"
          label="New password"
          type="password"
          required
          autocomplete="new-password"
          :error="errors.next"
          hint="At least 8 characters. Length does the work, so a passphrase beats a short password with symbols in it."
        />

        <AdminInput
          id="pw-confirm"
          v-model="form.confirm"
          label="Confirm new password"
          type="password"
          required
          autocomplete="new-password"
          :error="errors.confirm"
        />

        <div class="flex items-center gap-4">
          <UiButton type="submit" :disabled="state === 'saving'">
            <Icon
              v-if="state === 'saving'"
              name="lucide:loader-circle"
              class="size-4 animate-spin"
              aria-hidden="true"
            />
            {{ state === 'saving' ? 'Changing…' : 'Change password' }}
          </UiButton>
          <AdminStatus :state="state" :message="message" />
        </div>
      </form>
    </AdminPanel>

    <p class="text-xs leading-relaxed text-fg-subtle">
      Locked out instead? A password can be reset on the server with
      <code class="font-mono">go run ./cmd/adminctl passwd -email …</code>, which prompts rather than
      taking the password as an argument.
    </p>
  </div>
</template>
