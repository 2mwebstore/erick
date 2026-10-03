#!/usr/bin/env bash
#
# One-shot production check (§33). Exits non-zero if anything is wrong, so it
# works as a cron alert as well as a manual smoke test.
#
#   ./scripts/healthcheck.sh https://your-domain.com
#
# Cron, every five minutes, only mailing on failure:
#   */5 * * * * cd /srv/portfolio && ./scripts/healthcheck.sh https://your-domain.com >/dev/null
set -Eeuo pipefail

ORIGIN="${1:-${NUXT_PUBLIC_SITE_URL:-http://localhost:3000}}"
ORIGIN="${ORIGIN%/}"
FAILURES=0

check() {
  local label="$1" url="$2" expected="${3:-200}"
  local status
  status="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$url" || echo "000")"

  if [[ "$status" == "$expected" ]]; then
    printf '  ok    %-28s %s\n' "$label" "$status"
  else
    printf '  FAIL  %-28s %s (expected %s)\n' "$label" "$status" "$expected"
    FAILURES=$((FAILURES + 1))
  fi
}

echo "checking $ORIGIN"
check "homepage"        "$ORIGIN/"
check "work index"      "$ORIGIN/work"
check "case study"      "$ORIGIN/work/portfolio"
check "resume"          "$ORIGIN/resume"
check "frontend health" "$ORIGIN/health"
check "sitemap"         "$ORIGIN/sitemap.xml"
check "robots"          "$ORIGIN/robots.txt"
check "404 handling"    "$ORIGIN/work/does-not-exist" 404

# The contact endpoint must reject an empty body rather than accept it.
check "contact rejects empty" "$ORIGIN/api/contact" 400

if [[ "$ORIGIN" == https://* ]]; then
  HOST="${ORIGIN#https://}"
  echo "checking TLS for $HOST"
  EXPIRY="$(echo | openssl s_client -servername "$HOST" -connect "$HOST:443" 2>/dev/null \
    | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2 || true)"

  if [[ -z "$EXPIRY" ]]; then
    echo "  FAIL  certificate            could not be read"
    FAILURES=$((FAILURES + 1))
  else
    EXPIRY_EPOCH="$(date -j -f '%b %d %T %Y %Z' "$EXPIRY" +%s 2>/dev/null || date -d "$EXPIRY" +%s)"
    DAYS_LEFT=$(( (EXPIRY_EPOCH - $(date +%s)) / 86400 ))

    if (( DAYS_LEFT < 14 )); then
      printf '  FAIL  certificate            expires in %s days (%s)\n' "$DAYS_LEFT" "$EXPIRY"
      FAILURES=$((FAILURES + 1))
    else
      printf '  ok    certificate            %s days left\n' "$DAYS_LEFT"
    fi
  fi
fi

# Container health, when running on the host itself.
if command -v docker >/dev/null 2>&1 && docker compose ps --quiet >/dev/null 2>&1; then
  echo "container status"
  docker compose ps --format '  {{.Service}}: {{.State}} {{.Status}}' 2>/dev/null || true

  UNHEALTHY="$(docker compose ps --format '{{.Service}} {{.Status}}' 2>/dev/null | grep -ci 'unhealthy' || true)"
  if (( UNHEALTHY > 0 )); then
    echo "  FAIL  $UNHEALTHY container(s) unhealthy"
    FAILURES=$((FAILURES + 1))
  fi
fi

echo
if (( FAILURES > 0 )); then
  echo "$FAILURES check(s) failed"
  exit 1
fi
echo "all checks passed"
