# syntax=docker/dockerfile:1

# ── Build ────────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS build

WORKDIR /src

# Module files first so dependency download is cached separately from source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev

# CGO is off and the binary is stripped, so the runtime image needs no libc and
# stays under 15 MB.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/api ./cmd/api

# adminctl ships alongside the server so accounts can be managed in place:
#   docker compose exec api adminctl create -email … -name …
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/adminctl ./cmd/adminctl

# ── Runtime ──────────────────────────────────────────────────────────────────
FROM alpine:3.21 AS runtime

# ca-certificates for any outbound TLS; wget is busybox's, used by the health check.
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 api \
    && adduser -S -u 10001 -G api api

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/adminctl /usr/local/bin/adminctl

USER api:api

ENV HOST=0.0.0.0 \
    PORT=8080 \
    APP_ENV=production

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
