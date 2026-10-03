<script setup lang="ts">
/**
 * Portrait slot for the hero (§11, Option D).
 *
 * Renders the real photo when `siteConfig.portrait` points at one, and an
 * obviously-editable frame when it does not — the same rule the rest of the site
 * follows: show the gap, never fill it with a stock face.
 *
 * `sizes` must stay in px or vw. Nuxt Image parses this string itself and does
 * not understand `rem`: given one it produces no width candidates, and rather
 * than falling back it renders an <img> with an empty srcset and no src at all.
 * This was written as `16rem` and the portrait simply never appeared, with no
 * error anywhere. 256px is the same width.
 */
const props = defineProps<{
  src?: string | null
  alt?: string
}>()

const hasPhoto = computed(() => hasContent(props.src))
</script>

<template>
  <figure class="relative overflow-hidden rounded-xl border border-border bg-surface">
    <div class="aspect-4/5">
      <NuxtImg
        v-if="hasPhoto"
        :src="props.src!"
        :alt="props.alt ?? ''"
        class="size-full object-cover"
        sizes="(max-width: 1024px) 60vw, 256px"
        format="webp"
        quality="82"
        preload
      />

      <div
        v-else
        class="flex size-full flex-col items-center justify-center gap-3 border border-dashed border-border-strong px-4 text-center"
      >
        <Icon name="lucide:user-round" class="size-8 text-fg-subtle" aria-hidden="true" />
        <p class="font-mono text-[0.625rem] tracking-[0.14em] text-fg-subtle uppercase">
          Your photo
        </p>
        <p class="max-w-[14rem] text-xs leading-relaxed text-fg-subtle">
          Set
          <code class="font-mono">portrait</code>
          in Admin → Settings, to an image URL or a file you have put in
          <code class="font-mono">public/</code>
        </p>
      </div>
    </div>

    <!-- Hairline accent along the bottom edge, tying the frame to the
         diagram panel beneath it. -->
    <span class="absolute inset-x-0 bottom-0 h-px bg-accent-solid/40" aria-hidden="true" />
  </figure>
</template>
