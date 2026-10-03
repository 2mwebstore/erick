<script setup lang="ts">

const { t, tm, rt } = useI18n()

const { site } = useSiteContent()

/**
 * The production chain. Its stages are words rather than content, so they live
 * in the locale files: there is nothing here for an editor to change, only
 * something for a translator to.
 */
const lifecycle = computed(() => {
  const stages = tm('lifecycle')
  return Array.isArray(stages) ? stages.map((stage) => rt(stage as string)) : []
})

const aboutParagraphs = computed(() => site.value.aboutParagraphs)
</script>

<template>
  <!--
    The section heads with the name rather than a translated slogan. It is the
    one heading here that is a fact about a person rather than a claim about the
    work, so it belongs in Settings → Identity, where it can be corrected, and
    not in the locale files, where only a translator could reach it. `name` is
    translatable, so the Khmer page shows the Khmer spelling once one is typed.
  -->
  <UiSection
    id="about"
    :eyebrow="t('sections.about.eyebrow')"
    :title="site.name"
    title-class="headline-gradient"
    :subtitle="site.positioning"
  >
    <div class="grid gap-14 lg:grid-cols-12 lg:gap-16">
      <div class="reveal reveal-left space-y-5 lg:col-span-7">
        <p
          v-for="(paragraph, i) in aboutParagraphs"
          :key="i"
          class="text-[0.9375rem] leading-relaxed text-fg-muted sm:text-base"
        >
          {{ paragraph }}
        </p>

        <div class="hairline mt-8 pt-8">
          <p class="eyebrow">{{ t('resume.productionLifecycle') }}</p>
          <ul class="mt-4 flex flex-wrap gap-x-2 gap-y-2">
            <li v-for="(stage, i) in lifecycle" :key="stage" class="flex items-center gap-2">
              <span class="font-mono text-xs text-fg-muted">{{ stage }}</span>
              <Icon
                v-if="i < lifecycle.length - 1"
                name="lucide:chevron-right"
                class="size-3 text-border-strong"
                aria-hidden="true"
              />
            </li>
          </ul>
        </div>
      </div>
    </div>
  </UiSection>
</template>
