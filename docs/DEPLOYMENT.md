# Deployment Guide

This guide describes a practical MVP deployment for the URL shortener.

## Recommended MVP Topology

- **Frontend:** Vercel, Netlify, or any static host that can serve the Vite build.
- **Backend:** Fly.io, Render, Railway, or a small container host.
- **Database:** Managed PostgreSQL.
- **Redis:** Managed Redis recommended for redirect latency. `REDIS_URL` is required by configuration; the API degrades safely if that endpoint is temporarily unavailable.
- **ClickHouse:** Stores click analytics.
- **RabbitMQ:** A quorum queue with a persistent volume stores accepted click events until the worker writes them to ClickHouse.
- **Click worker:** Run `./click-worker` from the backend image as a separate application.
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
TRUSTED_PROXY_CIDRS= # comma-separated CIDRs of proxies directly connecting to the API
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
RABBITMQ_URL=amqp://user:password@rabbitmq:5672/
```

When `RABBITMQ_URL` is set, redirect requests hand clicks to the API's bounded in-memory queue and return immediately. Its background workers publish persistent messages to RabbitMQ and wait for publisher confirms. The click worker inserts batches into ClickHouse and acknowledges messages afterward. When RabbitMQ is unavailable or the in-memory queue fills, some click events are lost; redirects still succeed.

## Required Frontend Environment

```env
VITE_API_BASE_URL=https://your-api.example
VITE_CLERK_PUBLISHABLE_KEY=pk_live_...
VITE_SHORT_DOMAIN=https://your-short-domain.example
```

## Deployment Steps

1. Provision PostgreSQL and run `server/scripts/migrate-up.sh` with `POSTGRES_CONNECTION_STRING` set and `psql` installed on the migration host. Run it again to confirm there are no pending migrations.
2. Provision Redis, ClickHouse, and RabbitMQ. Keep RabbitMQ on a private network with a persistent volume.
3. Run `./clickhouse-migrate up` from the backend image before deploying this API version or starting the click worker. The new migration adds `event_id` for duplicate-safe analytics queries.
4. Create a Clerk production app and configure OAuth callback URLs for the frontend domain.
5. Deploy the backend container from `server/Dockerfile` and run a second instance of the image with command `./click-worker`.
6. Deploy the frontend container from `client/Dockerfile`, or run `pnpm build` and serve `client/dist`.
7. Configure DNS:
   - App domain to frontend host.
   - API domain to backend host.
   - Short-link domain to backend host if separate from the API domain.
8. Set `BASE_URL` and `VITE_SHORT_DOMAIN` to the public HTTPS short-link origin. Both should point to the backend redirect endpoint, including when the frontend uses a different domain.
9. Set `CORS_ALLOWED_ORIGINS` to the exact frontend origin.

## Coolify / Dokploy Notes

Deploy the backend, click worker, and frontend as **separate applications**. Add private RabbitMQ and ClickHouse services with persistent volumes. Do not use `docker-compose.yml` for production; it is a local development stack with development credentials and bundled databases.

For both applications:

- Docker build context: repository root
- Dockerfile path:
  - Backend: `server/Dockerfile`
  - Frontend: `client/Dockerfile`

Backend runtime environment:

- Set the backend variables from "Required Backend Environment".
- Expose port `8080`.
- Use `/api/v1/health/ready` as the health check path.

Click worker runtime environment:

- Set `RABBITMQ_URL`, `CLICKHOUSE_URL`, `CLICKHOUSE_USERNAME`, `CLICKHOUSE_PASSWORD`, and `CLICKHOUSE_TABLE_NAME`.
- Override the container command with `./click-worker` and do not expose a public port.
- Monitor the `link4it.clicks` backlog and the `link4it.clicks.dead` dead-letter queue. Quorum queues on one Dokploy host are durable across process restarts but do not survive loss of that host or its volume.

Frontend build environment:

The frontend is a Vite static build, so public frontend variables are compiled into the image at build time. Configure these as build arguments in Coolify/Dokploy:

```env
VITE_API_BASE_URL=https://your-api.example
VITE_CLERK_PUBLISHABLE_KEY=pk_live_...
VITE_SHORT_DOMAIN=https://your-short-domain.example
```

Expose frontend port `80`. The frontend image includes an Nginx SPA fallback so client-side routes work after refresh.

## Local Development Compose

`docker-compose.yml` is for local development only. It starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- ClickHouse on `localhost:9000` and `localhost:8123`
- RabbitMQ management on `localhost:15672`
- A click worker that consumes RabbitMQ
- Backend on `localhost:8080`
- Frontend on `localhost:5173`

The dev compose disables API rate limiting with `RATE_LIMIT_REQUESTS=0` to avoid interrupting local testing.

## Health Checks

Use these for uptime checks and platform health probes:

- `GET /api/v1/health/live` checks that the API process is running.
- `GET /api/v1/health/ready` checks required dependencies.

Readiness behavior:

- PostgreSQL unavailable or required tables missing: returns `503`.
- Redis unavailable: returns `200` with `degraded: true` because redirects still work through Postgres.
- Analytics is `disabled` when neither ClickHouse nor RabbitMQ is configured, `ok` when click delivery has not failed, or `degraded` after a startup connection failure or click loss. Analytics degradation keeps HTTP readiness at `200` while PostgreSQL is healthy. Check structured logs for cumulative lost-click counts. RabbitMQ publisher confirms happen in the background; a click lost before RabbitMQ accepts it cannot be recovered.

## Post-Deploy Smoke Test

```bash
curl -i https://your-api.example/api/v1/health/live
curl -i https://your-api.example/api/v1/health/ready
curl -i https://your-api.example/api/v1/reference
curl -i https://your-api.example/api/v1/openapi.yaml
```

Then use the UI:

1. Sign in with Clerk.
2. Create a link with a tag.
3. Copy the short URL from the UI. Confirm its host is the backend redirect domain, then visit it and check the destination and click count.
4. Update the link expiration and clear it again.
5. Try a nonexistent shortcode and an expired link; expect `404` and `410` respectively.
6. Confirm rate-limit and request ID headers are present:

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
- OpenAPI schema validation for `docs/openapi/openapi.yaml`
