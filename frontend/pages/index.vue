<script setup lang="ts">
const { site } = useSiteContent()

usePageSeo({
  title: site.value.role,
  description: site.value.description,
  path: '/',
})

useWebPageSchema({
  name: `${site.value.name} — ${site.value.role}`,
  description: site.value.description,
  path: '/',
})

// The homepage title is the site title, not "Role — Name"
useHead({ title: site.value.role })
</script>

<template>
  <div>
    <!-- §38 — deliberately few sections; depth lives in the case studies -->
    <SectionsHero />
    <SectionsAbout />
    <SectionsExperience />
    <SectionsWork />
    <SectionsCapabilities />
    <SectionsServices />
    <SectionsPrinciples />
    <SectionsContact />
  </div>
</template>
