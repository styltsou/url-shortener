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
VITE_SHORT_DOMAIN=your-short-domain.example
```

## Deployment Steps

1. Provision PostgreSQL and run migrations from `server/migrations`.
2. Provision Redis and optionally ClickHouse. These can be managed services or separate VPS/Coolify/Dokploy services.
3. Create a Clerk production app and configure OAuth callback URLs for the frontend domain.
4. Deploy the backend container from `server/Dockerfile`.
5. Deploy the frontend container from `client/Dockerfile`, or run `pnpm build` and serve `client/dist`.
6. Configure DNS:
   - App domain to frontend host.
   - API domain to backend host.
   - Short-link domain to backend host if separate from the API domain.
7. Set `BASE_URL` to the public short-link domain.
8. Set `CORS_ALLOWED_ORIGINS` to the exact frontend origin.

## Coolify / Dokploy Notes

Deploy the backend and frontend as **separate applications**. Do not use `docker-compose.yml` for production; it is a local development stack with development credentials and bundled databases.

For both applications:

- Docker build context: repository root
- Dockerfile path:
  - Backend: `server/Dockerfile`
  - Frontend: `client/Dockerfile`

Backend runtime environment:

- Set the backend variables from "Required Backend Environment".
- Expose port `8080`.
- Use `/api/v1/health/ready` as the health check path.

Frontend build environment:

The frontend is a Vite static build, so public frontend variables are compiled into the image at build time. Configure these as build arguments in Coolify/Dokploy:

```env
VITE_API_BASE_URL=https://your-api.example
VITE_CLERK_PUBLISHABLE_KEY=pk_live_...
VITE_SHORT_DOMAIN=your-short-domain.example
```

Expose frontend port `80`. The frontend image includes an Nginx SPA fallback so client-side routes work after refresh.

## Local Development Compose

`docker-compose.yml` is for local development only. It starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- ClickHouse on `localhost:9000` and `localhost:8123`
- Backend on `localhost:8080`
- Frontend on `localhost:5173`

The dev compose disables API rate limiting with `RATE_LIMIT_REQUESTS=0` to avoid interrupting local testing.

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
