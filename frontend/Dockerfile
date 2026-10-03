# syntax=docker/dockerfile:1

# ── Build ────────────────────────────────────────────────────────────────────
# The site is prerendered, so the canonical origin is baked in at build time.
# NUXT_PUBLIC_SITE_URL must be the real production origin here, or every
# canonical tag and Open Graph URL in the output will point somewhere else.
FROM node:22-alpine AS build

WORKDIR /app

# Dependencies are copied first so a source-only change reuses this layer.
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts

COPY . .

ARG NUXT_PUBLIC_SITE_URL=http://localhost:3000
ENV NUXT_PUBLIC_SITE_URL=${NUXT_PUBLIC_SITE_URL}
ENV NODE_ENV=production

RUN npm run postinstall && npm run build

# ── Runtime ──────────────────────────────────────────────────────────────────
# Only .output ships: no source, no node_modules, no build toolchain.
FROM node:22-alpine AS runtime

RUN apk add --no-cache wget

WORKDIR /app

ENV NODE_ENV=production \
    NITRO_PORT=3000 \
    NITRO_HOST=0.0.0.0

COPY --from=build --chown=node:node /app/.output ./.output

USER node

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:3000/ || exit 1

CMD ["node", ".output/server/index.mjs"]
