<script setup lang="ts">
useTheme()

const { site } = useSiteContent()

usePersonSchema()

const { public: pub } = useRuntimeConfig()

// Sets <html lang>/<html dir> and emits the hreflang alternates, so the two
// language versions are linked for search engines rather than looking like
// duplicate pages.
const localeHead = useLocaleHead({ lang: true, dir: true, seo: true })

useHead(() => ({
  htmlAttrs: localeHead.value.htmlAttrs,
  link: localeHead.value.link,
  meta: localeHead.value.meta,
}))

useHead({
  titleTemplate: (title?: string) =>
    title ? `${title} — ${site.value.name}` : `${site.value.name} — ${site.value.role}`,
})

useSeoMeta({
  ogSiteName: site.value.name,
  ogType: 'website',
  twitterCard: 'summary_large_image',
})

// Canonical origin is environment-driven; see .env.example
if (import.meta.server && pub.siteUrl.startsWith('http://localhost')) {
  console.warn('[seo] NUXT_PUBLIC_SITE_URL is unset — canonical URLs will point at localhost.')
}
</script>

<template>
  <div class="min-h-dvh bg-bg text-fg antialiased">
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>
