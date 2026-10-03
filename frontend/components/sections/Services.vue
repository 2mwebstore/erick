<script setup lang="ts">
const { t } = useI18n()
const { services, site } = useSiteContent()

const wordpressSeoStack = computed(() => site.value.wordpressSeoStack)
</script>

<template>
  <UiSection
    id="services"
    :eyebrow="t('sections.services.eyebrow')"
    :title="t('sections.services.title')"
    :subtitle="t('sections.services.subtitle')"
  >
    <div class="grid gap-x-12 gap-y-10 sm:grid-cols-2">
      <article
        v-for="(service, i) in services"
        :key="service.slug"
        class="reveal flex gap-4"
        :class="i % 2 === 0 ? 'reveal-left' : 'reveal-right'"
        :style="{ '--reveal-delay': `${Math.floor(i / 2) * 90}ms` }"
      >
        <Icon :name="service.icon" class="mt-0.5 size-5 shrink-0 text-accent" aria-hidden="true" />
        <div>
          <h3 class="text-base font-semibold tracking-tight text-fg">{{ service.title }}</h3>
          <UiProse :value="service.description" class="mt-2 text-sm" />
          <ul v-if="service.slug === 'wordpress-seo'" class="mt-4 flex flex-wrap gap-1.5">
            <li v-for="item in wordpressSeoStack" :key="item">
              <UiTag :label="item" />
            </li>
          </ul>
        </div>
      </article>
    </div>
  </UiSection>
</template>
