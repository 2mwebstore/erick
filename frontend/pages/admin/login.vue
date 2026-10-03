<script setup lang="ts">
definePageMeta({ layout: false })

const { login, ensureSession } = useAdmin()
const route = useRoute()

const email = ref('')
const password = ref('')
const status = ref<'idle' | 'submitting' | 'error'>('idle')
const message = ref('')

// Already signed in? Skip the form.
onMounted(async () => {
  if (await ensureSession()) await navigateTo(nextPath())
})

function nextPath() {
  const next = route.query.next
  // Only ever redirect within this site: an open redirect on a login page is a
  // gift to a phisher.
  return typeof next === 'string' && next.startsWith('/admin') ? next : '/admin'
}

async function submit() {
  if (status.value === 'submitting') return

  status.value = 'submitting'
  message.value = ''

  try {
    await login(email.value, password.value)
    await navigateTo(nextPath())
  } catch (error) {
    status.value = 'error'
    message.value =
      (error as { data?: { message?: string } })?.data?.message ??
      'Those details do not match an account.'
    password.value = ''
  }
}

useHead({ title: 'Sign in — Admin', meta: [{ name: 'robots', content: 'noindex, nofollow' }] })
</script>

<template>
  <div class="relative isolate flex min-h-dvh items-center justify-center overflow-hidden bg-bg px-6">
    <UiAurora intensity="soft" />

    <div class="w-full max-w-sm">
      <p class="eyebrow">Portfolio</p>
      <h1 class="mt-3 text-2xl font-semibold tracking-tight text-fg">Sign in</h1>
      <p class="mt-2 text-sm text-fg-muted">Content management for the portfolio site.</p>

      <form class="mt-8 space-y-5" novalidate @submit.prevent="submit">
        <AdminInput
          id="login-email"
          v-model="email"
          label="Email"
          type="email"
          required
          autocomplete="username"
        />
        <AdminInput
          id="login-password"
          v-model="password"
          label="Password"
          type="password"
          required
          autocomplete="current-password"
        />

        <UiButton type="submit" block :disabled="status === 'submitting'">
          <Icon
            v-if="status === 'submitting'"
            name="lucide:loader-circle"
            class="size-4 animate-spin"
            aria-hidden="true"
          />
          {{ status === 'submitting' ? 'Signing in…' : 'Sign in' }}
        </UiButton>

        <p
          v-if="status === 'error'"
          aria-live="polite"
          role="status"
          class="flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400"
        >
          <Icon name="lucide:alert-circle" class="size-4 shrink-0" aria-hidden="true" />
          {{ message }}
        </p>
      </form>

      <p class="mt-8 text-xs leading-relaxed text-fg-subtle">
        No account yet? Create the first one on the server:
        <code class="font-mono">adminctl create -email … -name …</code>
      </p>
    </div>
  </div>
</template>
