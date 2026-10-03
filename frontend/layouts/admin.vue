<script setup lang="ts">
const { user, isAdmin, staleApi, logout } = useAdmin()
const route = useRoute()

const nav = computed(() => {
  const items = [
    { label: 'Overview', to: '/admin', icon: 'lucide:layout-dashboard' },
    { label: 'Projects', to: '/admin/projects', icon: 'lucide:folder-open' },
    { label: 'Experience', to: '/admin/experience', icon: 'lucide:git-commit-horizontal' },
    { label: 'Skills', to: '/admin/capabilities', icon: 'lucide:layers' },
    { label: 'Services', to: '/admin/services', icon: 'lucide:briefcase' },
    { label: 'Positioning', to: '/admin/pillars', icon: 'lucide:list-ordered' },
    { label: 'Principles', to: '/admin/principles', icon: 'lucide:compass' },
    { label: 'Messages', to: '/admin/messages', icon: 'lucide:inbox' },
    { label: 'Settings', to: '/admin/settings', icon: 'lucide:settings' },
    // Your own account, so every role gets it.
    { label: 'Your account', to: '/admin/account', icon: 'lucide:user-round' },
  ]

  // Account management and the audit trail are admin-only, so an editor is not
  // shown links that would only 403.
  if (isAdmin.value) {
    items.push({ label: 'Users', to: '/admin/users', icon: 'lucide:users' })
    items.push({ label: 'Audit log', to: '/admin/audit', icon: 'lucide:scroll-text' })
  }

  return items
})

const isActive = (to: string) =>
  to === '/admin' ? route.path === '/admin' : route.path.startsWith(to)

useHead({
  titleTemplate: (title?: string) => (title ? `${title} — Admin` : 'Admin'),
  meta: [{ name: 'robots', content: 'noindex, nofollow' }],
})
</script>

<template>
  <div class="flex min-h-dvh flex-col bg-bg text-fg">
    <header class="sticky top-0 z-40 border-b border-border bg-bg/90 backdrop-blur-md">
      <div class="mx-auto flex h-14 max-w-[84rem] items-center justify-between gap-6 px-6">
        <div class="flex items-center gap-3">
          <NuxtLink to="/admin" class="text-sm font-semibold tracking-tight text-fg">
            Admin
          </NuxtLink>
          <span class="font-mono text-[0.625rem] tracking-[0.14em] text-fg-subtle uppercase">
            {{ user?.role }}
          </span>
        </div>

        <div class="flex items-center gap-2">
          <NavigationThemeToggle />
          <UiButton to="/" variant="ghost" size="sm">
            <Icon name="lucide:external-link" class="size-3.5" aria-hidden="true" />
            View site
          </UiButton>
          <UiButton variant="secondary" size="sm" @click="logout">Sign out</UiButton>
        </div>
      </div>
    </header>

    <div class="mx-auto flex w-full max-w-[84rem] flex-1 flex-col gap-8 px-6 py-8 lg:flex-row lg:gap-12">
      <nav aria-label="Admin sections" class="lg:w-52 lg:shrink-0">
        <ul class="flex gap-1 overflow-x-auto lg:flex-col lg:overflow-visible">
          <li v-for="item in nav" :key="item.to">
            <NuxtLink
              :to="item.to"
              class="inline-flex items-center gap-2 rounded-md px-3 py-2 text-[0.8125rem] font-medium whitespace-nowrap transition-colors"
              :class="
                isActive(item.to)
                  ? 'bg-surface text-fg'
                  : 'text-fg-muted hover:bg-surface/60 hover:text-fg'
              "
              :aria-current="isActive(item.to) ? 'page' : undefined"
            >
              <Icon :name="item.icon" class="size-4 shrink-0" aria-hidden="true" />
              {{ item.label }}
            </NuxtLink>
          </li>
        </ul>
      </nav>

      <main class="min-w-0 flex-1">
        <!--
          A server left running through a deploy rejects the current request
          bodies with "Malformed request.", which says nothing about why. This
          says it once, at the top, before anything is typed into a form that
          cannot save.
        -->
        <div
          v-if="staleApi"
          role="alert"
          class="mb-6 rounded-lg border border-amber-500/40 bg-amber-500/10 px-5 py-4 text-sm leading-relaxed text-fg"
        >
          <p class="font-medium">The API is running an older build than this page.</p>
          <p class="mt-1 text-fg-muted">
            Saving will fail with “Malformed request.” until it is restarted.
            Restart the Go API
            (<code class="font-mono text-xs">cd backend &amp;&amp; go run ./cmd/api</code>),
            then reload this page.
          </p>
        </div>

        <slot />
      </main>

      <UiToasts />
    </div>
  </div>
</template>
