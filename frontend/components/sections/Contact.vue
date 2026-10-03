<script setup lang="ts">
import { projectTypes } from '~/content/services'
import { LIMITS } from '~/utils/validation'

const { t } = useI18n()

const { form, errors, status, feedback, messageLength, submit, clearError } = useContactForm()
const { public: pub } = useRuntimeConfig()
const { site } = useSiteContent()

// Either the build-time flag or the CMS setting can switch the form off.
const contactEnabled = computed(() => pub.contactEnabled && site.value.contactEnabled)

const inputClass = (invalid: boolean) => [
  'w-full rounded-lg border bg-bg px-3.5 py-2.5 text-sm text-fg placeholder:text-fg-subtle',
  'transition-[border-color,box-shadow] duration-200 focus:outline-none',
  'focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/20',
  invalid ? 'border-red-500 dark:border-red-400' : 'border-border-strong hover:border-fg-subtle',
]

/**
 * The contact card only lists details that exist.
 *
 * An empty row reading "Phone" with nothing beside it is worse than no row, and
 * §27 applies here as much as anywhere: these are filled in from the admin
 * panel, not invented.
 */
const details = computed(() =>
  [
    {
      key: 'email',
      icon: 'lucide:mail',
      label: t('contact.email'),
      value: site.value.email,
      href: site.value.email ? `mailto:${site.value.email}` : undefined,
    },
    {
      key: 'phone',
      icon: 'lucide:phone',
      label: t('contact.phone'),
      value: site.value.phone,
      // tel: cannot contain spaces and the stored value is formatted for reading.
      href: site.value.phone ? `tel:${site.value.phone.replace(/[^\d+]/g, '')}` : undefined,
    },
    {
      key: 'location',
      icon: 'lucide:map-pin',
      label: t('contact.location'),
      value: site.value.location,
      href: undefined,
    },
  ].filter((row) => hasContent(row.value)),
)
</script>

<template>
  <UiSection
    id="contact"
    class="relative isolate overflow-hidden"
    :eyebrow="t('sections.contact.eyebrow')"
    :title="t('sections.contact.title')"
    :subtitle="t('sections.contact.subtitle')"
  >
    <UiAurora intensity="soft" />

    <div class="grid gap-8 lg:grid-cols-12 lg:gap-10">
      <!-- ── Form ─────────────────────────────────────────────────────── -->
      <div class="reveal reveal-left lg:col-span-7">
        <div class="rounded-xl border border-border bg-surface/40 p-6 backdrop-blur-[2px] sm:p-8">
          <h3 class="text-lg font-semibold tracking-tight text-fg">
            {{ t('sections.contact.title') }}
          </h3>

          <form
            v-if="contactEnabled"
            novalidate
            class="mt-6 space-y-5"
            @submit.prevent="submit"
          >
            <div class="grid gap-5 sm:grid-cols-2">
              <UiField id="contact-name" :label="t('contact.name')" required :error="errors.name">
                <template #default="{ describedBy, invalid }">
                  <input
                    id="contact-name"
                    v-model="form.name"
                    type="text"
                    name="name"
                    autocomplete="name"
                    :placeholder="t('contact.placeholder.name')"
                    :maxlength="LIMITS.name.max"
                    :aria-invalid="invalid || undefined"
                    :aria-describedby="describedBy"
                    :class="inputClass(invalid)"
                    @input="clearError('name')"
                  >
                </template>
              </UiField>

              <UiField id="contact-email" :label="t('contact.email')" required :error="errors.email">
                <template #default="{ describedBy, invalid }">
                  <input
                    id="contact-email"
                    v-model="form.email"
                    type="email"
                    name="email"
                    autocomplete="email"
                    inputmode="email"
                    :placeholder="t('contact.placeholder.email')"
                    :maxlength="LIMITS.email.max"
                    :aria-invalid="invalid || undefined"
                    :aria-describedby="describedBy"
                    :class="inputClass(invalid)"
                    @input="clearError('email')"
                  >
                </template>
              </UiField>
            </div>

            <UiField id="contact-phone" :label="t('contact.phone')" :error="errors.phone">
              <template #default="{ describedBy, invalid }">
                <input
                  id="contact-phone"
                  v-model="form.phone"
                  type="tel"
                  name="phone"
                  autocomplete="tel"
                  inputmode="tel"
                  :placeholder="t('contact.placeholder.phone')"
                  :maxlength="LIMITS.phone.max"
                  :aria-invalid="invalid || undefined"
                  :aria-describedby="describedBy"
                  :class="inputClass(invalid)"
                  @input="clearError('phone')"
                >
              </template>
            </UiField>

            <UiField id="contact-subject" :label="t('contact.subject')" :error="errors.subject">
              <template #default="{ describedBy, invalid }">
                <input
                  id="contact-subject"
                  v-model="form.subject"
                  type="text"
                  name="subject"
                  :placeholder="t('contact.placeholder.subject')"
                  :maxlength="LIMITS.subject.max"
                  :aria-invalid="invalid || undefined"
                  :aria-describedby="describedBy"
                  :class="inputClass(invalid)"
                  @input="clearError('subject')"
                >
              </template>
            </UiField>

            <UiField
              id="contact-projectType"
              :label="t('contact.projectType')"
              required
              :error="errors.projectType"
            >
              <template #default="{ describedBy, invalid }">
                <select
                  id="contact-projectType"
                  v-model="form.projectType"
                  name="projectType"
                  :aria-invalid="invalid || undefined"
                  :aria-describedby="describedBy"
                  :class="inputClass(invalid)"
                  @change="clearError('projectType')"
                >
                  <option value="" disabled>{{ t('contact.selectProjectType') }}</option>
                  <!-- The value stays the canonical English: it is what the
                       API validates against and what lands in the inbox. Only
                       the label is translated. -->
                  <option v-for="type in projectTypes" :key="type" :value="type">
                    {{ t(`contact.type.${type}`) }}
                  </option>
                </select>
              </template>
            </UiField>

            <UiField
              id="contact-message"
              :label="t('contact.message')"
              required
              :error="errors.message"
              :hint="`${messageLength} / ${LIMITS.message.max}`"
            >
              <template #default="{ describedBy, invalid }">
                <textarea
                  id="contact-message"
                  v-model="form.message"
                  name="message"
                  rows="6"
                  :placeholder="t('contact.placeholder.message')"
                  :maxlength="LIMITS.message.max"
                  :aria-invalid="invalid || undefined"
                  :aria-describedby="describedBy"
                  :class="[...inputClass(invalid), 'resize-y']"
                  @input="clearError('message')"
                />
              </template>
            </UiField>

            <!-- Honeypot: hidden from sight and from assistive technology, -->
            <!-- so only automated submissions ever fill it (§20 spam protection). -->
            <div class="absolute -left-[9999px] h-0 w-0 overflow-hidden" aria-hidden="true">
              <label for="contact-company">{{ t('resume.company') }}</label>
              <input
                id="contact-company"
                v-model="form.company"
                type="text"
                name="company"
                tabindex="-1"
                autocomplete="off"
              >
            </div>

            <UiButton type="submit" block :disabled="status === 'submitting'">
              <Icon
                v-if="status === 'submitting'"
                name="lucide:loader-circle"
                class="size-4 animate-spin"
                aria-hidden="true"
              />
              <Icon v-else name="lucide:send" class="size-4" aria-hidden="true" />
              {{ status === 'submitting' ? t('contact.sending') : t('contact.send') }}
            </UiButton>

            <p aria-live="polite" role="status" class="text-sm">
              <span v-if="status === 'success'" class="inline-flex items-center gap-1.5 text-fg">
                <Icon name="lucide:check-circle-2" class="size-4 text-accent" aria-hidden="true" />
                {{ feedback }}
              </span>
              <span
                v-else-if="status === 'error' && feedback"
                class="inline-flex items-center gap-1.5 text-red-600 dark:text-red-400"
              >
                <Icon name="lucide:alert-circle" class="size-4" aria-hidden="true" />
                {{ feedback }}
              </span>
            </p>
          </form>

          <p v-else class="mt-6 text-sm text-fg-muted">
            {{ t('contact.closed') }}
            <a :href="`mailto:${site.email}`" class="text-accent hover:underline">{{ site.email }}</a>
          </p>
        </div>
      </div>

      <!-- ── Details ──────────────────────────────────────────────────── -->
      <div class="reveal reveal-right lg:col-span-5" style="--reveal-delay: 120ms">
        <h3 class="text-lg font-semibold tracking-tight text-fg">
          {{ t('contact.information') }}
        </h3>

        <ul v-if="details.length" class="mt-6 space-y-3">
          <li
            v-for="row in details"
            :key="row.key"
            class="rounded-xl border border-border bg-surface/40 p-4 transition-colors duration-200 hover:border-border-strong"
          >
            <component
              :is="row.href ? 'a' : 'div'"
              :href="row.href"
              class="flex items-center gap-4"
              :class="row.href ? 'group' : ''"
            >
              <span
                class="inline-flex size-11 shrink-0 items-center justify-center rounded-lg border border-accent/25 bg-accent/10 text-accent"
              >
                <Icon :name="row.icon" class="size-5" aria-hidden="true" />
              </span>
              <span class="min-w-0">
                <span class="block text-xs text-fg-subtle">{{ row.label }}</span>
                <span
                  class="block truncate text-[0.9375rem] text-fg transition-colors"
                  :class="row.href ? 'group-hover:text-accent' : ''"
                >
                  {{ row.value }}
                </span>
              </span>
            </component>
          </li>
        </ul>

        <template v-if="site.profiles.length">
          <h3 class="mt-10 text-lg font-semibold tracking-tight text-fg">
            {{ t('contact.followMe') }}
          </h3>
          <UiProfileLinks :profiles="site.profiles" class="mt-4" />
        </template>

        <p
          v-if="hasContent(site.availability)"
          class="mt-10 rounded-xl border border-accent/25 bg-accent/5 px-5 py-4 text-sm leading-relaxed text-fg-muted"
        >
          {{ site.availability }}
        </p>

        <div class="mt-10">
          <p class="eyebrow">{{ t('contact.usefulToInclude') }}</p>
          <ul class="mt-4 space-y-2 text-sm leading-relaxed text-fg-muted">
            <li>{{ t('contact.hint.what') }}</li>
            <li>{{ t('contact.hint.who') }}</li>
            <li>{{ t('contact.hint.stack') }}</li>
            <li>{{ t('contact.hint.timeline') }}</li>
          </ul>
        </div>
      </div>
    </div>
  </UiSection>
</template>
