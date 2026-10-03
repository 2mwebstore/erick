<script setup lang="ts">
/**
 * Aurora ribbon background (§25).
 *
 * Angled bands of light sweeping slowly across the top of a section, heavily
 * blurred and at low opacity. Long offset cycles (38–52s) mean the movement
 * reads as ambient rather than animated; the spec warns against distracting
 * background motion, so nothing here crosses the viewport quickly or repeats
 * on a noticeable beat.
 *
 * Pure CSS: no canvas, no scroll listener, no JavaScript at runtime. It costs
 * only compositing, and it stops moving entirely under prefers-reduced-motion.
 */
withDefaults(defineProps<{ intensity?: 'normal' | 'soft' }>(), { intensity: 'normal' })
</script>

<template>
  <div
    class="aurora pointer-events-none absolute inset-0 -z-10 overflow-hidden"
    :class="intensity === 'soft' ? 'aurora--soft' : ''"
    aria-hidden="true"
  >
    <span class="aurora__ribbon aurora__ribbon--a" />
    <span class="aurora__ribbon aurora__ribbon--b" />
    <span class="aurora__ribbon aurora__ribbon--c" />
    <span class="aurora__grain" />
  </div>
</template>

<style scoped>
.aurora {
  --aurora-opacity: 0.28;
  --aurora-blur: 70px;
  /* Fade out before the section ends so ribbons never collide with the next
     section's hairline divider. */
  mask-image: linear-gradient(to bottom, #000 0%, #000 48%, transparent 92%);
}

:root.dark .aurora {
  /* Blurred light reads dimmer on a near-black ground. */
  --aurora-opacity: 0.4;
  --aurora-blur: 78px;
}

.aurora--soft {
  --aurora-opacity: 0.14;
}

:root.dark .aurora--soft {
  --aurora-opacity: 0.2;
}

.aurora__ribbon {
  position: absolute;
  display: block;
  left: -35%;
  width: 170%;
  height: 14rem;
  filter: blur(var(--aurora-blur));
  opacity: var(--aurora-opacity);
  will-change: transform;
  /* Soft ends, so a ribbon never shows a hard edge where it is cut off. */
  background-image: linear-gradient(
    100deg,
    transparent 0%,
    var(--aurora-tint) 28%,
    var(--aurora-tint) 62%,
    transparent 100%
  );
}

.aurora__ribbon--a {
  top: -4rem;
  /*
     The palette's mid blue. It is the one colour here that cannot be text in
     either theme — too light for the pale end, too dark for the ink — so it
     gets the job with no contrast requirement, which is also the job it is
     best at.
  */
  --aurora-tint: var(--wash-a);
  transform: rotate(-14deg);
  animation: aurora-sweep-a 38s ease-in-out infinite alternate;
}

.aurora__ribbon--b {
  top: 4rem;
  height: 11rem;
  /* The second wash, so the bands read as depth rather than one flat tint. */
  --aurora-tint: var(--wash-b);
  opacity: calc(var(--aurora-opacity) * 0.7);
  transform: rotate(-8deg);
  animation: aurora-sweep-b 52s ease-in-out infinite alternate;
}

.aurora__ribbon--c {
  top: 11rem;
  height: 9rem;
  /* A neutral third band keeps the accent from dominating the page (§3). */
  --aurora-tint: var(--border-strong);
  opacity: calc(var(--aurora-opacity) * 1.1);
  transform: rotate(-19deg);
  animation: aurora-sweep-c 45s ease-in-out infinite alternate;
}

/* Fine grain over the ribbons: breaks up the banding that large blurred
   gradients show on 8-bit displays, and keeps the result feeling printed
   rather than airbrushed. */
.aurora__grain {
  position: absolute;
  inset: 0;
  opacity: 0.35;
  mix-blend-mode: overlay;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3'/%3E%3C/filter%3E%3Crect width='120' height='120' filter='url(%23n)' opacity='0.22'/%3E%3C/svg%3E");
}

@keyframes aurora-sweep-a {
  from {
    transform: translate3d(-6%, 0, 0) rotate(-14deg) scaleY(1);
  }
  to {
    transform: translate3d(6%, 1.5rem, 0) rotate(-11deg) scaleY(1.18);
  }
}

@keyframes aurora-sweep-b {
  from {
    transform: translate3d(5%, 0, 0) rotate(-8deg) scaleY(1.1);
  }
  to {
    transform: translate3d(-7%, -1rem, 0) rotate(-12deg) scaleY(0.9);
  }
}

@keyframes aurora-sweep-c {
  from {
    transform: translate3d(-3%, 0.5rem, 0) rotate(-19deg) scaleY(0.95);
  }
  to {
    transform: translate3d(7%, -0.5rem, 0) rotate(-16deg) scaleY(1.2);
  }
}

/* A still image, not a slower one. */
@media (prefers-reduced-motion: reduce) {
  .aurora__ribbon {
    animation: none !important;
  }
}
</style>
