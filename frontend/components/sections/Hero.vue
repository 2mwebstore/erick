<script setup lang="ts">
const { site, pillars } = useSiteContent()

const { t } = useI18n()
const localePath = useLocalePath()
const home = computed(() => localePath('/'))
const anchor = (hash: string) => `${home.value === '/' ? '' : home.value}/#${hash}`
</script>

<template>
  <section class="relative isolate overflow-hidden">
    <UiAurora />
    <UiCodeBackdrop />

    <div class="mx-auto max-w-[84rem] px-6 pt-16 pb-20 sm:pt-24 sm:pb-28 lg:px-8">
      <div class="grid items-center gap-14 lg:grid-cols-12 lg:gap-16">
        <!-- Staggered entrance: role, headline, tagline, then the calls to
             action. One pass on load; the headline's colour is the only thing
             here that keeps moving afterwards. -->
        <div class="lg:col-span-7">
          <p class="reveal eyebrow" style="--reveal-delay: 60ms">{{ site.role }}</p>

          <!--
            The colour travels through the half of the headline that carries the
            name, and the opening words stay solid ink; with no quieter half to
            animate, the whole line takes it instead.

            One line, and the space between the halves written out: Vue drops
            whitespace between two elements, and "Hi, I'mKONG CHANSILA" is what
            that looks like. A plain space rather than &nbsp;, so a narrow screen
            may still wrap between the two.
          -->
          <h1
            :class="[
              'reveal reveal-blur mt-5 text-4xl font-semibold tracking-[-0.03em] text-fg sm:text-5xl lg:text-[3.5rem] lg:leading-[1.04]',
              site.headlineTail ? '' : 'headline-gradient',
            ]"
            style="--reveal-delay: 140ms"
          >{{ site.headline }}{{ ' ' }}<span v-if="site.headlineTail" class="headline-gradient">{{ site.headlineTail }}</span></h1>

          <p
            class="reveal mt-7 max-w-xl text-base leading-relaxed text-fg-muted sm:text-lg"
            style="--reveal-delay: 260ms"
          >
            {{ site.tagline }}
          </p>

          <p
            class="reveal mt-6 font-mono text-xs tracking-[0.12em] text-fg-subtle uppercase"
            style="--reveal-delay: 340ms"
          >
            {{ t('hero.fullTime') }} · {{ yearRange(site.startYear) }}
          </p>

          <div class="reveal mt-9 flex flex-wrap gap-3" style="--reveal-delay: 420ms">
            <UiButton :to="anchor('work')">
              {{ t('actions.viewSelectedWork') }}
              <Icon name="lucide:arrow-down-right" class="size-4" aria-hidden="true" />
            </UiButton>
            <UiButton :to="anchor('contact')" variant="secondary">{{ t('actions.letsTalk') }}</UiButton>
          </div>

          <!-- Compact technology line (§10) — not the full stack -->
          <ul class="reveal mt-10 flex flex-wrap items-center gap-x-4 gap-y-2" style="--reveal-delay: 500ms">
            <li
              v-for="tech in site.heroStack"
              :key="tech"
              class="inline-flex items-center gap-1.5 font-mono text-xs text-fg-subtle"
            >
              <Icon
                v-if="techIconMeta(tech)"
                :name="techIconMeta(tech)!.icon"
                class="tech-icon size-3.5 shrink-0"
                :style="{
                  '--tc': techIconMeta(tech)!.color,
                  '--tc-dark': techIconMeta(tech)!.dark ?? techIconMeta(tech)!.color,
                }"
                aria-hidden="true"
              />
              {{ tech }}
            </li>
          </ul>
        </div>

        <!-- Portrait beside the intro (§11). Stacks naturally on narrow screens. -->
        <div class="reveal reveal-scale flex flex-col gap-6 lg:col-span-5" style="--reveal-delay: 320ms">
          <div class="mx-auto w-full max-w-[16rem] sm:max-w-[18rem] lg:mx-0">
            <UiPortrait :src="site.portrait" :alt="site.portraitAlt" />
          </div>
        </div>
      </div>

      <!-- Three-point positioning (§1) -->
      <dl v-if="pillars.length" class="mt-20 grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:mt-24 sm:grid-cols-3">
        <div v-for="(pillar, i) in pillars" :key="pillar.id ?? pillar.slug" class="bg-bg p-6 sm:p-7">
          <dt class="flex items-baseline gap-3">
            <!-- Numbered by position, so reordering renumbers them. -->
            <span class="font-mono text-xs text-accent">{{ String(i + 1).padStart(2, '0') }}</span>
            <span class="text-base font-semibold tracking-tight text-fg">{{ pillar.title }}</span>
          </dt>
          <dd class="mt-3"><UiProse :value="pillar.description" class="text-sm" /></dd>
        </div>
      </dl>
    </div>
  </section>
</template>
