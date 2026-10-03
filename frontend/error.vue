<script setup lang="ts">
import type { NuxtError } from '#app'

const { t } = useI18n()

const props = defineProps<{ error: NuxtError }>()

const is404 = computed(() => props.error?.statusCode === 404)

useHead({ title: is404.value ? 'Page not found' : 'Something went wrong' })
useSeoMeta({ robots: 'noindex, follow' })
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <LayoutHeader />

    <main id="main" class="flex flex-1 items-center">
      <div class="mx-auto w-full max-w-2xl px-6 py-24 lg:px-8">
        <p class="eyebrow">Error {{ error?.statusCode ?? 500 }}</p>
        <h1 class="mt-4 text-3xl font-semibold tracking-tight text-fg sm:text-4xl">
          {{ is404 ? 'That page does not exist.' : 'Something went wrong.' }}
        </h1>
        <p class="mt-4 text-base leading-relaxed text-fg-muted">
          {{
            is404
              ? 'The page may have moved, or the link may be incomplete. The work and contact pages are below.'
              : 'The error has been logged. Please try again, or get in touch if it keeps happening.'
          }}
        </p>

        <div class="mt-9 flex flex-wrap gap-3">
          <UiButton to="/">{{ t('actions.backToHome') }}</UiButton>
          <UiButton to="/work" variant="secondary">{{ t('actions.viewWork') }}</UiButton>
          <UiButton to="/#contact" variant="ghost">{{ t('nav.contact') }}</UiButton>
        </div>
      </div>
    </main>

  </div>
</template>
