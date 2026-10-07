import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',

  // Keep the flat project layout documented in docs/ARCHITECTURE.md
  // (components/, pages/, composables/ at the frontend root) rather than
  // Nuxt 4's nested app/ directory.
  srcDir: '.',
  dir: { app: 'app' },

  modules: ['@nuxtjs/i18n', '@nuxt/icon', '@nuxt/image', '@nuxt/fonts', '@nuxt/eslint'],

  css: ['~/assets/css/main.css'],

  vite: {
    plugins: [tailwindcss()],
  },

  typescript: {
    strict: true,
    typeCheck: false, // run explicitly via `npm run typecheck`
  },

  icon: {
    mode: 'svg', // inline SVG, no client-side icon fetching
    // Brand marks (simple-icons) alongside the UI set (lucide).
    serverBundle: { collections: ['lucide', 'simple-icons'] },
    clientBundle: {
      scan: true,
      sizeLimitKb: 256,
    },
  },

  image: {
    format: ['avif', 'webp'],
    quality: 78,
    screens: { sm: 640, md: 768, lg: 1024, xl: 1280, xxl: 1536 },
  },

  i18n: {
    // English lives at /, Khmer at /km/. Distinct URLs rather than one address
    // serving two languages: a search engine can index both, and the swr cache
    // keys on the path, so a cached English page can never be served to a
    // Khmer request.
    strategy: 'prefix_except_default',
    defaultLocale: 'en',
    // Origin for the hreflang alternates, which must be absolute URLs. Same value
    // as siteUrl; NUXT_PUBLIC_I18N_BASE_URL overrides it at runtime.
    baseUrl: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
    locales: [
      { code: 'en', language: 'en-US', name: 'English' },
      { code: 'km', language: 'km-KH', name: 'ភាសាខ្មែរ' },
    ],
    // Messages are bundled through i18n/i18n.config.ts, not lazy-loaded files.
    vueI18n: './i18n.config.ts',
    // Off deliberately. Redirecting on Accept-Language would be cached by the
    // swr rules below and then applied to everyone, and a visitor who picked a
    // language should not have it overridden on their next visit.
    detectBrowserLanguage: false,
  },

  fonts: {
    families: [
      { name: 'Inter', provider: 'google', weights: [400, 500, 600, 700], subsets: ['latin', 'latin-ext'] },
      { name: 'JetBrains Mono', provider: 'google', weights: [400, 500], subsets: ['latin'] },
      // Khmer. Kantumruy Pro carries Latin too, so mixed sentences keep one voice.
      { name: 'Kantumruy Pro', provider: 'google', weights: [400, 500, 600, 700], subsets: ['khmer', 'latin'] },
    ],
    defaults: {
      fallbacks: { 'sans-serif': ['Helvetica Neue', 'Arial'] },
    },
  },

  runtimeConfig: {
    // Server-only: the Go API is never exposed to the browser.
    apiBaseUrl: process.env.NUXT_API_BASE_URL || 'http://localhost:8080',
    apiTimeoutMs: process.env.NUXT_API_TIMEOUT_MS || '8000',
    public: {
      // Set NUXT_PUBLIC_SITE_URL to the real origin before deploying.
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
      contactEnabled: process.env.NUXT_PUBLIC_CONTACT_ENABLED !== 'false',
      // Google Search Console ownership code(s); read at runtime, so setting it
      // needs a restart, not a rebuild. See docs/SEARCH_CONSOLE.md.
      googleSiteVerification: process.env.NUXT_PUBLIC_GOOGLE_SITE_VERIFICATION || '',
    },
  },

  nitro: {
    compressPublicAssets: { gzip: true, brotli: true },

    // Works around a Nuxt payload-cache collision. Nuxt mounts the runtime
    // payload cache at `cache:nuxt:payload:` (a directory). Rendering `/` calls
    // setItem('/'), which unstorage normalises to '', so the full key collapses
    // to `cache:nuxt:payload` — one separator short of its own mountpoint. It
    // therefore misses that mount, falls through to the parent `cache` mount,
    // and is written as a *file* at .nuxt/cache/nuxt/payload. Every other route
    // then tries to write inside that file and fails with ENOTDIR (and if a
    // route renders first, `/` fails with EISDIR instead).
    //
    // Mounting `cache:nuxt` sends that collapsed key to its own base, while the
    // longer, more specific `cache:nuxt:payload:` mount still wins for the
    // per-route keys — so the file and the directory no longer fight over one
    // path. Runtime payload extraction is forced on by the swr rules below (it
    // is not `experimental.payloadExtraction`), so it cannot simply be turned
    // off without giving up the caching.
    //
    // Dev only, deliberately: the built server mounts no `cache` storage at
    // all, so these keys land in memory there and never collide. Declaring it
    // under `storage` as well would point production at a build directory that
    // is not deployed.
    devStorage: {
      'cache:nuxt': { driver: 'fs', base: '.nuxt/cache/nuxt-root' },
    },

    routeRules: {
      // Content lives in MySQL now, so pages cannot be built from files at build
      // time. Stale-while-revalidate is the compromise: a visitor is served from
      // cache immediately and the page is refreshed behind them, so the database
      // is off the critical path for all but the first request in each window.
      '/': { swr: 300 },
      '/work': { swr: 300 },
      '/work/**': { swr: 300 },
      '/resume': { swr: 600 },
      // The Khmer routes are separate documents and cache separately.
      '/km': { swr: 300 },
      '/km/work': { swr: 300 },
      '/km/work/**': { swr: 300 },
      '/km/resume': { swr: 600 },
      '/sitemap.xml': { swr: 3600 },

      // The old monogram icon, removed when the photo icon set replaced it.
      // Google and browsers that saw it keep asking for this address; a
      // permanent redirect hands them the current icon rather than the
      // not-found page. Nothing on the site links here any more.
      '/favicon.svg': { redirect: { to: '/favicon-96x96.png', statusCode: 301 } },

      // Never cache anything authenticated or written.
      // The admin panel is client-rendered: it needs no SEO, and rendering it on
      // the server would mean forwarding session cookies through SSR for no gain.
      '/admin/**': {
        ssr: false,
        swr: false,
        headers: { 'x-robots-tag': 'noindex, nofollow', 'cache-control': 'no-store' },
      },
      '/api/admin/**': { swr: false },
      '/api/auth/**': { swr: false },
      '/api/contact': { swr: false },
    },
  },

  app: {
    head: {
      // <html lang> is set per route by useLocaleHead in app.vue.
      script: [
        {
          // Apply the stored theme before first paint to avoid a flash.
          innerHTML:
            '(function(){try{var s=localStorage.getItem("theme");var m=window.matchMedia("(prefers-color-scheme: dark)").matches;var d=s==="dark"||(s!=="light"&&m);document.documentElement.classList.toggle("dark",d);document.documentElement.style.colorScheme=d?"dark":"light"}catch(e){}})()',
          tagPosition: 'head',
        },
      ],
      // The icon set in public/. favicon.ico also answers browsers and crawlers
      // that ask for /favicon.ico without reading these tags. Google Search
      // wants a favicon a multiple of 48px square, which the 48 and 96 px PNGs
      // are. The home-screen icons are listed in site.webmanifest, and Windows
      // tiles in browserconfig.xml. tests/favicon.test.ts checks that every
      // file named here and in the manifest exists.
      link: [
        { rel: 'icon', href: '/favicon.ico', sizes: '48x48' },
        { rel: 'icon', type: 'image/png', sizes: '96x96', href: '/favicon-96x96.png' },
        { rel: 'icon', type: 'image/png', sizes: '48x48', href: '/favicon-48x48.png' },
        { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png' },
        { rel: 'icon', type: 'image/png', sizes: '16x16', href: '/favicon-16x16.png' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' },
      ],
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'theme-color', content: '#0a1931', media: '(prefers-color-scheme: dark)' },
        { name: 'theme-color', content: '#f6fafd', media: '(prefers-color-scheme: light)' },
      ],
    },
  },

  features: { inlineStyles: false },

  experimental: { payloadExtraction: true },
})
