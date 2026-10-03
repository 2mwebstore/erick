# Monitoring

## Endpoints

| Endpoint | Service | Meaning |
| --- | --- | --- |
| `GET /health` | Nuxt | The site process is alive |
| `GET /health` | Go API | The API is alive, plus database reachability |

The two are deliberately separate. An uptime monitor should be able to tell "the
site is down" apart from "the contact form is down" — collapsing them into one
signal loses that distinction.

The API returns **503** when MySQL is unreachable, so a database outage registers
as an outage rather than a healthy process that cannot actually serve requests.

```bash
curl -s https://your-domain.com/health
docker compose exec api wget -qO- http://127.0.0.1:8080/health
```

```json
{
  "status": "ok",
  "version": "a1b2c3d",
  "uptime": "3h12m4s",
  "checks": { "api": "ok", "database": "ok" },
  "time": "2026-09-23T12:00:00Z"
}
```

## What to watch

| Signal | Threshold | Where |
| --- | --- | --- |
| Site uptime | any failure | External monitor on `/` |
| API health | any 503 | External monitor or cron |
| TLS expiry | under 14 days | `scripts/healthcheck.sh` |
| Disk usage | over 80% | Host monitor |
| Memory | over 85% sustained | Host monitor |
| CPU | over 80% for 5 min | Host monitor |
| Container state | any `unhealthy` | `docker compose ps` |
| HTTP 5xx rate | any sustained increase | Nginx access log |
| Application errors | any `level=ERROR` | `docker compose logs api` |
| Rate-limit hits | a sudden spike | `rate limit exceeded` log lines |
| Backup success | a missing daily run | `journalctl -t portfolio-backup` |

## The bundled check

```bash
./scripts/healthcheck.sh https://your-domain.com
```

Checks every route, the sitemap, robots.txt, 404 handling, that the contact
endpoint rejects an empty body, TLS expiry, and container health. Exits non-zero
on any failure, so it works as a cron alert as well as a manual smoke test.

```cron
*/5 * * * * cd /srv/portfolio && ./scripts/healthcheck.sh https://your-domain.com >/dev/null
```

An external monitor (Better Stack, Uptime Kuma, Cloudflare Health Checks) is
still worth having — a check running on the host cannot tell you the host is
unreachable.

## Logs

The API logs one structured line per request: JSON in production, text locally.
Every line carries a `request_id` that is also returned as `X-Request-ID`, so a
visitor's report can be traced to an exact request.

```bash
docker compose logs -f api
docker compose logs api | grep '"level":"ERROR"'
docker compose logs api | grep 'rate limit exceeded'
docker compose logs --since 30m api | grep '"status":5'
```

No message bodies and no email addresses appear in logs — that is intentional.

Cap log growth so a noisy week cannot fill the disk:

```yaml
# per service in docker-compose.yml
logging:
  driver: json-file
  options:
    max-size: '10m'
    max-file: '5'
```

## Useful queries

```sql
-- submissions per day, last fortnight
SELECT DATE(created_at) AS day, COUNT(*) AS messages
FROM contact_messages
WHERE created_at >= NOW() - INTERVAL 14 DAY
GROUP BY day ORDER BY day DESC;

-- repeated submissions from one address, a sign of abuse or a broken form
SELECT ip_address, COUNT(*) AS n, MAX(created_at) AS latest
FROM contact_messages
WHERE created_at >= NOW() - INTERVAL 7 DAY
GROUP BY ip_address HAVING n > 3 ORDER BY n DESC;

-- what people are actually asking for
SELECT project_type, COUNT(*) AS n
FROM contact_messages GROUP BY project_type ORDER BY n DESC;
```

## When something breaks

**Site returns 502** — the web container is down or unhealthy.

```bash
docker compose ps
docker compose logs --tail=100 web
docker compose up -d web
```

**Contact form returns an error but pages load** — the API or MySQL is down. The
site itself is unaffected, which is the point of prerendering.

```bash
docker compose exec api wget -qO- http://127.0.0.1:8080/health
docker compose logs --tail=100 api db
```

**Everything is slow** — check the host before the application.

```bash
df -h; free -m; docker stats --no-stream
```
