# Deployment Guide

This guide describes a practical MVP deployment for the URL shortener.

## Recommended MVP Topology

- **Frontend:** Vercel, Netlify, or any static host that can serve the Vite build.
- **Backend:** Fly.io, Render, Railway, or a small container host.
- **Database:** Managed PostgreSQL.
- **Redis:** Managed Redis. Optional at runtime, but recommended for redirect latency.
- **ClickHouse:** Optional for MVP analytics. If unset or unavailable, analytics degrades safely.
- **Auth:** Clerk production application.

## Required Backend Environment

```env
APP_ENV=production
PORT=8080
BASE_URL=https://your-short-domain.example
POSTGRES_CONNECTION_STRING=postgres://...
REDIS_URL=host:6379
REDIS_USERNAME=
REDIS_PASSWORD=...
REDIS_DB=0
CLERK_SECRET_KEY=sk_live_...
CORS_ALLOWED_ORIGINS=https://your-frontend.example
CORS_EXPOSED_HEADERS=Link,X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset,Retry-After
RATE_LIMIT_REQUESTS=120
RATE_LIMIT_WINDOW_SECONDS=60
SERVER_READ_TIMEOUT=15
SERVER_WRITE_TIMEOUT=15
SERVER_IDLE_TIMEOUT=60
```

Optional analytics:

```env
CLICKHOUSE_URL=host:9000
CLICKHOUSE_USERNAME=default
CLICKHOUSE_PASSWORD=...
CLICKHOUSE_TABLE_NAME=link4it.click_events
```

## Required Frontend Environment

```env
VITE_API_BASE_URL=https://your-api.example
VITE_CLERK_PUBLISHABLE_KEY=pk_live_...
```

## Deployment Steps

1. Provision PostgreSQL and run migrations from `server/migrations`.
2. Provision Redis and optionally ClickHouse.
3. Create a Clerk production app and configure OAuth callback URLs for the frontend domain.
4. Deploy the backend container from `server/Dockerfile`.
5. Deploy the frontend container from `client/Dockerfile`, or run `pnpm build` and serve `client/dist`.
6. Configure DNS:
   - App domain to frontend host.
   - API domain to backend host.
   - Short-link domain to backend host if separate from the API domain.
7. Set `BASE_URL` to the public short-link domain.
8. Set `CORS_ALLOWED_ORIGINS` to the exact frontend origin.

## Health Checks

Use these for uptime checks and platform health probes:

- `GET /api/v1/health/live` checks that the API process is running.
- `GET /api/v1/health/ready` checks required dependencies.

Readiness behavior:

- PostgreSQL unavailable: returns `503`.
- Redis unavailable: returns `200` with `degraded: true` because redirects still work through Postgres.

## Post-Deploy Smoke Test

```bash
curl -i https://your-api.example/api/v1/health/live
curl -i https://your-api.example/api/v1/health/ready
curl -i https://your-api.example/api/v1/reference
```

Then use the UI:

1. Sign in with Clerk.
2. Create a link with a tag.
3. Visit the short URL and confirm it redirects.
4. Update the link expiration and clear it again.
5. Confirm rate-limit and request ID headers are present:

```bash
curl -I https://your-api.example/api/v1/health/live
```

Expected operational headers include:

- `X-Request-ID`
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

## CI

GitHub Actions runs on pushes to `main` and pull requests:

- `go test ./...` in `server`
- `pnpm check` in `client`
- OpenAPI YAML parsing for `docs/openapi/openapi.yaml`
