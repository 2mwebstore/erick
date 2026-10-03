<script setup lang="ts">
import type { AdminUser } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { api, isAdmin } = useAdmin()
const toast = useToast()

const users = ref<AdminUser[]>([])
const loading = ref(true)

const { page, pageSize, total, items } = usePagedList(users)
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

const form = reactive({ email: '', name: '', password: '', role: 'editor' })

async function load() {
  try {
    const res = await api<{ users: AdminUser[] }>('/api/admin/users')
    users.value = res.users
  } catch (e) {
    state.value = 'error'
    message.value = (e as { message?: string }).message ?? 'Could not load users.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function create() {
  state.value = 'saving'
  errors.value = {}

  try {
    await api('/api/admin/users', { method: 'POST', body: { ...form } })
    state.value = 'saved'
    message.value = 'Account created.'
    toast.success(`Created the account for ${form.email}.`)
    form.email = ''
    form.name = ''
    form.password = ''
    await load()
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    errors.value = err.errors ?? {}
    message.value = err.message ?? 'Could not create that account.'
    toast.error(message.value)
  }
}

const when = (iso?: string) =>
  iso ? new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }) : 'never'

useHead({ title: 'Users' })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight text-fg">Users</h1>
      <p class="mt-2 text-sm text-fg-muted">
        Editors can change content. Admins can also manage accounts and read the audit log.
      </p>
    </header>

    <p v-if="!isAdmin" class="text-sm text-fg-muted">This section requires an admin account.</p>

    <template v-else>
      <AdminPanel title="Accounts">
        <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>
        <table v-else class="w-full text-left text-sm">
          <thead>
            <tr class="border-b border-border text-xs text-fg-subtle">
              <th scope="col" class="pb-2 font-medium">Email</th>
              <th scope="col" class="pb-2 font-medium">Name</th>
              <th scope="col" class="pb-2 font-medium">Role</th>
              <th scope="col" class="pb-2 font-medium">Last login</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr v-for="user in items" :key="user.id">
              <td class="py-2.5 text-fg">{{ user.email }}</td>
              <td class="py-2.5 text-fg-muted">{{ user.name }}</td>
              <td class="py-2.5">
                <span class="font-mono text-[0.6875rem] tracking-[0.1em] text-fg-subtle uppercase">
                  {{ user.role }}
                </span>
              </td>
              <td class="py-2.5 text-xs text-fg-subtle">{{ when(user.last_login_at) }}</td>
            </tr>
          </tbody>
        </table>

        <AdminPagination
          v-if="!loading"
          v-model:page="page"
          v-model:page-size="pageSize"
          :total="total"
          label="accounts"
          class="mt-4"
        />
      </AdminPanel>

      <AdminPanel
        title="Add an account"
        description="Passwords must be at least 12 characters and mix letters with something else."
      >
        <form class="space-y-4" novalidate @submit.prevent="create">
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminInput id="u-email" v-model="form.email" label="Email" type="email" required :error="errors.email" />
            <AdminInput id="u-name" v-model="form.name" label="Name" required :error="errors.name" />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminInput
              id="u-password"
              v-model="form.password"
              label="Password"
              type="password"
              required
              :error="errors.password"
            />
            <div>
              <label for="u-role" class="block text-[0.8125rem] font-medium text-fg">Role</label>
              <select
                id="u-role"
                v-model="form.role"
                class="mt-1.5 w-full rounded-md border border-border-strong bg-bg px-3 py-2 text-sm text-fg"
              >
                <option value="editor">Editor</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>

          <div class="flex items-center gap-4">
            <UiButton type="submit" size="sm" :disabled="state === 'saving'">Create account</UiButton>
            <AdminStatus :state="state" :message="message" />
          </div>
        </form>
      </AdminPanel>
    </template>
  </div>
</template>
