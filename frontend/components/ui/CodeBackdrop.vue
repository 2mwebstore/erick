<script setup lang="ts">
/**
 * Code-editor decoration behind the hero.
 *
 * Two editor panes, the technologies drifting past as brand-coloured chips, and
 * a few loose tokens. It is scenery, so all of it is `aria-hidden`, takes no
 * pointer events and sits behind everything — but none of it is invented: the
 * pane is built from the real settings and the chips are the real stack, so the
 * decoration cannot drift away from what the page says about itself (§27).
 * Change either in the admin panel and the background changes with it.
 *
 * Placed in the strip beside the portrait and in the margins around the text,
 * because that is where the hero has nothing in it — a panel behind the
 * headline is a panel you have to read through. Each slot carries the
 * breakpoint at which its space actually exists, which is measured rather than
 * assumed: one breakpoint below 2xl the right-hand strip is narrower than a
 * pane, and the pane went straight through the portrait.
 */
const { site, projects, capabilities } = useSiteContent()

/**
 * What the editor pane shows.
 *
 * Real settings, and deliberately not the stack: the hero already prints the
 * stack in the row under the buttons, and the pane was repeating three of the
 * same words a few hundred pixels away. Location is on no other part of this
 * screen, and the line is dropped rather than printed empty when it is unset.
 */
const fields = computed(() => {
  const out: { key: string; value: string; kind: 'str' | 'num' }[] = [
    { key: 'name', value: site.value.name, kind: 'str' },
  ]
  if (site.value.location) out.push({ key: 'from', value: site.value.location, kind: 'str' })
  out.push({ key: 'since', value: String(site.value.startYear), kind: 'num' })
  return out
})

/**
 * The two cards, and the timing each floats on.
 *
 * Duration and delay are handed to the card and to its shadow as one custom
 * property, because the two have to agree exactly: the shadow is widest and
 * faintest at the moment the card is highest, and a hundred milliseconds of
 * drift between them reads as a mistake rather than as a shadow.
 */
const PANES = {
  file: { at: 'right-[1%] top-[6%] w-[17rem]', duration: '9s', delay: '-2s' },
  shell: { at: 'right-[1%] top-[31%] w-[15rem]', duration: '12s', delay: '-5s' },
}

/**
 * Where a chip may sit, and how it moves.
 *
 * Every slot has its own duration and a negative delay, so the chips are at
 * different points of their cycle from the first frame and never drift in
 * formation. Nothing here is on a beat anyone can follow.
 */
const SLOTS = [
  // The band above the hero text, free at every width.
  { at: 'left-[18%] top-[3%] hidden md:inline-flex', delay: '-1s', duration: '11s' },
  { at: 'right-[26%] top-[4%] hidden lg:inline-flex', delay: '-4s', duration: '13s' },
  // The band under the pillars, free once the columns sit side by side.
  { at: 'left-[30%] bottom-[4%] hidden lg:inline-flex', delay: '-6s', duration: '12s' },
  { at: 'right-[17%] bottom-[5%] hidden lg:inline-flex', delay: '-10s', duration: '15s' },
  { at: 'left-[52%] top-[2%] hidden lg:inline-flex', delay: '-12s', duration: '10s' },
  { at: 'right-[38%] bottom-[8%] hidden xl:inline-flex', delay: '-2s', duration: '16s' },
  // Beside the portrait, under the panes.
  { at: 'right-[3%] top-[57%] hidden xl:inline-flex', delay: '-3s', duration: '10s' },
  /*
     The left margin, which only exists once the viewport is wider than the
     container plus a chip: at 2xl the margin is 96px and a chip is 100, so it
     would have sat on the tagline.
  */
  { at: 'left-[3%] top-[44%] hidden min-[1620px]:inline-flex', delay: '-8s', duration: '9s' },
]

/** "Nuxt", "Nuxt 4" and "nuxt" are one technology, not three. */
const fold = (tech: string) => tech.toLowerCase().replace(/\s+\d+(\.\d+)?$/, '').trim()

/**
 * The technologies the chips float, and the reason they are not the hero stack.
 *
 * The stack is already printed in full in the row under the buttons, so
 * floating the same eight words around it put every one of them on screen
 * twice. These come from the technologies listed on the real projects instead —
 * the same data, one layer further in — with anything already in that row
 * filtered out, including a compound like "Laravel / PHP" that names one of
 * them. The stack is the fallback for a site that has no projects yet, where
 * there is nothing to duplicate.
 */
const chips = computed(() => {
  const onScreen = new Set(site.value.heroStack.map(fold))
  const named = (tech: string) =>
    onScreen.has(fold(tech)) || fold(tech).split(/[\s/]+/).some((word) => onScreen.has(word))

  const seen = new Set<string>()
  const pool: { tech: string; meta: NonNullable<ReturnType<typeof techIconMeta>> }[] = []

  const consider = (tech: string) => {
    const key = fold(tech)
    if (seen.has(key)) return
    seen.add(key)
    const meta = techIconMeta(tech)
    if (meta) pool.push({ tech, meta })
  }

  for (const project of projects.value) {
    for (const tech of project.technologies ?? []) if (!named(tech)) consider(tech)
  }
  for (const capability of capabilities.value) {
    for (const tech of capability.items ?? []) if (!named(tech)) consider(tech)
  }
  // Nothing of its own to show: fall back to the stack rather than nothing.
  if (!pool.length) for (const tech of site.value.heroStack) consider(tech)

  return pool.slice(0, SLOTS.length).map((chip, i) => ({ ...chip, slot: SLOTS[i]! }))
})

/**
 * Loose tokens for the corners. Plain language constructs, and ones from this
 * stack rather than a more fashionable one: the decoration should not claim a
 * technology that is not used here.
 */
const TOKENS = [
  { text: '<script setup>', class: 'left-[26%] top-[4%] hidden md:block', delay: '0s', duration: '11s' },
  // In the band under the pillars: the left margin it used to sit in does not
  // exist until 2xl, and below that it was printed across a pillar card.
  { text: 'go func() {}', class: 'left-[6%] bottom-[9%] hidden lg:block', delay: '-3s', duration: '14s' },
  { text: 'await $fetch()', class: 'right-[2%] top-[48%] hidden xl:block', delay: '-6s', duration: '12s' },
  { text: 'SELECT * FROM', class: 'left-[44%] bottom-[3%] hidden xl:block', delay: '-9s', duration: '15s' },
]
</script>

<template>
  <div class="code-backdrop pointer-events-none absolute inset-0 -z-10" aria-hidden="true">
    <!--
      Each card sits on its own stage: the card, and the shadow it casts on the
      ground under it. The stage is what is positioned; the card moves inside
      it, so the shadow stays where the ground is.
    -->
    <div
      class="float-stage absolute hidden 2xl:block"
      :class="PANES.file.at"
      :style="{ '--float-duration': PANES.file.duration, '--float-delay': PANES.file.delay }"
    >
      <span class="ground-shadow" />
      <!-- The editor pane: real values, so it stays true as the settings change. -->
      <figure class="code-pane float3d">
        <figcaption class="code-pane__bar">
          <span class="code-pane__dot code-pane__dot--a" />
          <span class="code-pane__dot code-pane__dot--b" />
          <span class="code-pane__dot code-pane__dot--c" />
          <span class="ml-2 truncate">developer.ts</span>
        </figcaption>
        <pre class="code-pane__body"><span class="tok-key">const</span> <span class="tok-name">developer</span> = {
<template v-for="field in fields" :key="field.key">  <span class="tok-prop">{{ field.key }}</span>: <span :class="field.kind === 'num' ? 'tok-num' : 'tok-str'">{{ field.kind === 'num' ? field.value : `"${field.value}"` }}</span>,
</template>};</pre>
      </figure>
    </div>

    <div
      class="float-stage absolute hidden 2xl:block"
      :class="PANES.shell.at"
      :style="{ '--float-duration': PANES.shell.duration, '--float-delay': PANES.shell.delay }"
    >
      <span class="ground-shadow" />
      <!-- The shell pane: the commands this project is actually built with. No
           output, and no timings — an invented build time is an invented fact. -->
      <figure class="code-pane float3d">
        <figcaption class="code-pane__bar">
          <span class="code-pane__dot code-pane__dot--a" />
          <span class="code-pane__dot code-pane__dot--b" />
          <span class="code-pane__dot code-pane__dot--c" />
          <span class="ml-2 truncate">~ deploy</span>
        </figcaption>
        <!-- <pre class="code-pane__body"><span class="tok-prompt">$</span> go build ./cmd/api
<span class="tok-prompt">$</span> nuxt build
<span class="tok-prompt">$</span> docker compose up -d<span class="caret" /></pre> -->
      </figure>
    </div>

    <!-- The stack itself, drifting past in its own colours. -->
    <span
      v-for="chip in chips"
      :key="chip.tech"
      class="tech-chip float3d absolute items-center gap-1.5"
      :class="chip.slot.at"
      :style="{
        '--tc': chip.meta.color,
        '--tc-dark': chip.meta.dark ?? chip.meta.color,
        '--float-delay': chip.slot.delay,
        '--float-duration': chip.slot.duration,
      }"
    >
      <Icon :name="chip.meta.icon" class="tech-icon size-3.5 shrink-0" />
      {{ chip.tech }}
    </span>

    <span
      v-for="token in TOKENS"
      :key="token.text"
      class="code-token float3d absolute"
      :class="token.class"
      :style="{ '--float-delay': token.delay, '--float-duration': token.duration }"
    >{{ token.text }}</span>
  </div>
</template>

<style scoped>
/*
   Quiet enough to be a background. The panes carry real text, so they are the
   one thing here that must never compete with the headline for attention.
*/
.code-backdrop {
  /*
     The room the cards float in. `perspective` on the parent is what makes a
     rotation read as depth rather than as a squashed rectangle; 1000px is far
     enough back that the tilt is gentle across a pane this size.

     It is not clipped here — the hero section already has `overflow-hidden`, so
     nothing escapes, and an overflow of its own would force this container back
     to `transform-style: flat` and flatten everything inside it.
  */
  perspective: 1000px;
  perspective-origin: 50% 35%;
  transform-style: preserve-3d;

  --code-strength: 0.58;

  /*
     Syntax colours, by role rather than by decoration: a keyword, a property, a
     string, a number. Chosen to stay legible at the strength above on the pale
     theme, where a background element has the least contrast to spare.
  */
  --syn-key: #8250df;
  --syn-name: #1a3d63;
  --syn-prop: #0a66c2;
  --syn-str: #0a7b52;
  --syn-num: #b35309;
}

:root.dark .code-backdrop {
  /* Dim ink on a near-black ground reads fainter than it measures. */
  --code-strength: 0.72;

  --syn-key: #c297ff;
  --syn-name: #e6eef7;
  --syn-prop: #84abc9;
  --syn-str: #4fc08d;
  --syn-num: #ffab70;
}

/*
   Glass: a translucent gradient fill, a pale edge catching the light along the
   top, and the page behind it blurred. `backdrop-filter` is what makes it glass
   rather than a grey box — the aurora moving behind a pane is visible through
   it, diffused.

   The white edge is barely there on the pale theme, so the border mixes toward
   the page's own ink there and toward white in the dark, which is where glass
   actually looks like glass.
*/
.code-pane {
  opacity: var(--code-strength);
  border-radius: 0.7rem;
  border: 1px solid color-mix(in srgb, var(--fg) 12%, transparent);
  background-color: color-mix(in srgb, var(--surface) 55%, transparent);
  background-image: linear-gradient(
    135deg,
    rgb(255 255 255 / 40%) 0%,
    rgb(255 255 255 / 8%) 42%,
    rgb(255 255 255 / 0%) 100%
  );
  backdrop-filter: blur(14px) saturate(130%);
  -webkit-backdrop-filter: blur(14px) saturate(130%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 45%),
    0 12px 32px rgb(10 25 49 / 12%);
  overflow: hidden;
}

:root.dark .code-pane {
  border-color: rgb(255 255 255 / 14%);
  background-image: linear-gradient(
    135deg,
    rgb(255 255 255 / 12%) 0%,
    rgb(255 255 255 / 3%) 45%,
    rgb(255 255 255 / 0%) 100%
  );
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 12%),
    0 16px 40px rgb(0 0 0 / 35%);
}

.code-pane__bar {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  border-bottom: 1px solid var(--border);
  padding: 0.4rem 0.6rem;
  font-family: var(--font-mono);
  font-size: 0.625rem;
  color: var(--fg-subtle);
}

/* The window buttons, in the colours a window's buttons are. */
.code-pane__dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 9999px;
  background-color: var(--border-strong);
}

.code-pane__dot--a {
  background-color: #ff5f57;
}

.code-pane__dot--b {
  background-color: #febc2e;
}

.code-pane__dot--c {
  background-color: #28c840;
}

.code-pane__body {
  margin: 0;
  padding: 0.7rem 0.75rem 0.85rem;
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  line-height: 1.7;
  color: var(--fg-muted);
  white-space: pre;
  overflow: hidden;
}

.tok-key {
  color: var(--syn-key);
}

.tok-name {
  color: var(--syn-name);
}

.tok-prop {
  color: var(--syn-prop);
}

.tok-str {
  color: var(--syn-str);
}

.tok-num {
  color: var(--syn-num);
}

.tok-prompt {
  color: var(--syn-str);
}

/* The one thing in a terminal that is supposed to blink. */
.caret {
  display: inline-block;
  width: 0.45em;
  height: 1em;
  margin-left: 0.15em;
  vertical-align: text-bottom;
  background-color: var(--syn-prop);
  animation: code-caret 1.2s steps(2, start) infinite;
}

@keyframes code-caret {
  to {
    opacity: 0;
  }
}

/*
   A technology chip in its own brand colour: the mark at full strength, the
   border and fill mixed down from it so the colour is present without the chip
   becoming a button. --tc / --tc-dark arrive from utils/tech-icons.ts, the same
   pair the technology row under the buttons uses.
*/
/*
   No `display` here. A scoped rule carries the component's data attribute, so
   `.tech-chip[data-v-…]` outranks Tailwind's `.hidden` and every chip stayed on
   screen at 390px, sitting across the eyebrow and the portrait. The slot's own
   `hidden md:inline-flex` is what decides whether a chip exists at a width, and
   it can only do that if nothing else sets display.
*/
.tech-chip {
  --chip: var(--tc, var(--fg-subtle));

  border-radius: 9999px;
  border: 1px solid color-mix(in srgb, var(--chip) 55%, transparent);
  background-color: color-mix(in srgb, var(--chip) 12%, color-mix(in srgb, var(--surface) 60%, transparent));
  background-image: linear-gradient(135deg, rgb(255 255 255 / 35%), rgb(255 255 255 / 5%));
  backdrop-filter: blur(8px) saturate(125%);
  -webkit-backdrop-filter: blur(8px) saturate(125%);
  padding: 0.25rem 0.6rem;
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  line-height: 1;
  color: color-mix(in srgb, var(--chip) 72%, var(--fg));
  white-space: nowrap;
  opacity: 0.9;
  /* The colour it casts on the page behind it. */
  box-shadow: 0 0 1.75rem color-mix(in srgb, var(--chip) 30%, transparent);
}

:root.dark .tech-chip {
  --chip: var(--tc-dark, var(--tc, var(--fg-subtle)));

  background-color: color-mix(in srgb, var(--chip) 16%, color-mix(in srgb, var(--surface) 55%, transparent));
  background-image: linear-gradient(135deg, rgb(255 255 255 / 14%), rgb(255 255 255 / 2%));
  color: color-mix(in srgb, var(--chip) 78%, var(--fg));
  opacity: 0.95;
}

.code-token {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--fg-subtle);
  opacity: calc(var(--code-strength) * 0.85);
  white-space: nowrap;
}

/*
   The float.

   One keyframe set for everything that floats: it rises, tips on both axes and
   grows very slightly as it comes toward the viewer, then settles back. Scale
   and rotation together are what sell it — a card that only moves up and down
   reads as a sliding rectangle, and one that also turns a degree and a half
   reads as an object with a near edge and a far one.

   The cycle runs to 100% rather than alternating, so the tilt can go one way on
   the way up and the other on the way down instead of retracing its path.
   Transform only, so it composites and never triggers layout.
*/
.float3d {
  animation: float3d var(--float-duration, 12s) ease-in-out var(--float-delay, 0s) infinite;
  transform-style: preserve-3d;
  will-change: transform;
}

@keyframes float3d {
  0% {
    transform: translateY(0.55rem) rotateX(2.2deg) rotateY(-3.5deg) scale(0.985);
  }

  35% {
    transform: translateY(-0.35rem) rotateX(-1deg) rotateY(1.5deg) scale(1.004);
  }

  50% {
    transform: translateY(-0.95rem) rotateX(-2.4deg) rotateY(3.8deg) scale(1.018);
  }

  75% {
    transform: translateY(-0.2rem) rotateX(0.6deg) rotateY(0.5deg) scale(1.001);
  }

  100% {
    transform: translateY(0.55rem) rotateX(2.2deg) rotateY(-3.5deg) scale(0.985);
  }
}

/*
   The ground shadow, on the same clock as the card above it.

   It is the part that makes the height readable: high up, the shadow spreads
   and washes out; low down, it draws in and darkens. Same duration, same delay
   and the same keyframe percentages as `float3d`, because a shadow that lags
   its object by even a fraction of a second stops looking like a shadow.
*/
.ground-shadow {
  position: absolute;
  bottom: -0.85rem;
  left: 50%;
  width: 74%;
  height: 0.9rem;
  border-radius: 9999px;
  background: radial-gradient(ellipse at center, rgb(10 25 49 / 55%) 0%, rgb(10 25 49 / 0%) 70%);
  filter: blur(7px);
  transform: translateX(-50%) scale(0.8);
  opacity: 0.6;
  animation: float3d-shadow var(--float-duration, 12s) ease-in-out var(--float-delay, 0s) infinite;
  will-change: transform, opacity;
}

:root.dark .ground-shadow {
  background: radial-gradient(ellipse at center, rgb(0 0 0 / 70%) 0%, rgb(0 0 0 / 0%) 70%);
}

@keyframes float3d-shadow {
  0% {
    transform: translateX(-50%) scale(0.8);
    opacity: 0.6;
  }

  35% {
    transform: translateX(-50%) scale(1.06);
    opacity: 0.33;
  }

  50% {
    transform: translateX(-50%) scale(1.2);
    opacity: 0.2;
  }

  75% {
    transform: translateX(-50%) scale(0.95);
    opacity: 0.44;
  }

  100% {
    transform: translateX(-50%) scale(0.8);
    opacity: 0.6;
  }
}

/*
   Still, not slower. The card rests level, and its shadow rests at the size and
   strength it has when the card is down — the two have to agree even when
   neither is moving.
*/
@media (prefers-reduced-motion: reduce) {
  .float3d,
  .ground-shadow,
  .caret {
    animation: none;
  }

  .float3d {
    transform: none;
  }
}

/* Decoration with no information in it: it should not cost a sheet of paper. */
@media print {
  .code-backdrop {
    display: none;
  }
}
</style>
