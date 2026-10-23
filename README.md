# link4.it — URL Shortener

A Go and React URL shortener MVP with tags, QR codes, and optional click analytics. This repository is being prepared for its first release; no public deployment is listed yet.

## Stack

**Backend:** Go 1.25, chi, sqlc, PostgreSQL, ClickHouse, Redis, Clerk  
**Frontend:** React 19, TypeScript, Vite, TanStack Router, TanStack Query, Tailwind CSS 4, Radix UI  
**Infra:** Docker Compose

## Features

- Create short links with auto-generated or custom shortcodes
- Tag management — create, assign, filter links by tags
- Click analytics — clicks over time, top referrers, user agents (ClickHouse)
- QR code generation per link
- Redis caching for fast redirects
- Clerk authentication (Google/GitHub OAuth)
- Request IDs, structured production logging, security headers, rate limiting, health checks
- Dark/light mode, responsive UI

## Architecture

```mermaid
flowchart LR
    B[Browser / React] -->|Clerk token| A[Go API]
    V[Short URL visitor] -->|HTTP redirect| A
    A --> P[(PostgreSQL)]
    A -.->|redirect cache| R[(Redis)]
    A -.->|click events| C[(ClickHouse)]
```

PostgreSQL stores links and tags. Redis speeds up redirects; the API uses PostgreSQL if Redis is unavailable. ClickHouse powers click analytics; links and redirects still work if it is unavailable. Clerk is required for signed-in management routes.

## Quick Start

Prerequisites: Docker with Compose, the PostgreSQL `psql` CLI, and a Clerk development application with its secret and publishable keys. Add `http://localhost:5173` to the Clerk application's allowed local origins and configure a sign-in method.

```bash
# From the repository root

# Set Clerk API keys (get them from https://dashboard.clerk.com)
export CLERK_SECRET_KEY=sk_test_...
export VITE_CLERK_PUBLISHABLE_KEY=pk_test_...

# Start PostgreSQL, apply migrations, then start the app
docker compose up postgres -d
export POSTGRES_CONNECTION_STRING='postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable'
(cd server && ./scripts/migrate-up.sh)
docker compose up --build
```

Open http://localhost:5173, sign in with Google/GitHub, and start shortening URLs.
The app supports whichever sign-in methods you enabled in Clerk. The local redirect endpoint is http://localhost:8080.

## Project Structure

```
server/        # Go backend
client/        # React frontend
docs/          # Architecture, onboarding, code reviews
docker-compose.yml
```

See [docs/WELCOME.md](docs/WELCOME.md) for a full developer onboarding guide.
See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for MVP deployment instructions.

## API Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/{shortcode}` | Redirect |
| `GET` | `/api/v1/health/live` | Liveness check |
| `GET` | `/api/v1/health/ready` | Dependency readiness check |
| `POST` | `/api/v1/links` | Create link |
| `GET` | `/api/v1/links` | List links |
| `GET` | `/api/v1/shortcodes/availability?shortcode={code}` | Check custom alias availability |
| `GET` | `/api/v1/links/{shortcode}/analytics` | Analytics |
| `GET` | `/api/v1/links/{shortcode}/qrcode` | QR code |

The API reference is at http://localhost:8080/api/v1/reference and its OpenAPI document is at http://localhost:8080/api/v1/openapi.yaml.

Operational headers:

- `X-Request-ID` is returned on every response and included in production logs.
- `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` describe rate-limit state.

Shortcodes use 1–20 ASCII letters, digits, hyphens or underscores. `api` is reserved. A generated shortcode begins with a deterministic 7-character code; an occupied code retries with a random code. Creating the same exact URL again returns an existing active link for that user (`200`, `existing: true`) without changing its settings or tags. Supply `force_new: true` or a custom shortcode to create a separate link (`201`, `existing: false`). Inactive, expired, and deleted links are not offered. Soft-deleted shortcodes can be reused.

Redis redirect entries live for at most one minute and never past a link's expiration. Updates and deletes invalidate them on a best-effort basis. If Redis is unavailable during invalidation, a cached redirect may remain usable until its TTL ends; immediate revocation is not guaranteed. Redirects fall back to PostgreSQL when Redis is unavailable. ClickHouse analytics are best-effort: click events can be lost if ClickHouse is unavailable, the memory queue fills, or shutdown exceeds five seconds. `/api/v1/health/ready` reports the analytics delivery state; logs record cumulative loss counts. Analytics default to the last 30 days, with selectable 7-day, 90-day, and 12-month ranges, or an inclusive custom UTC date range of up to 366 days. Choosing the same start and end date shows one day. Charts support daily, weekly, or monthly buckets. ClickHouse click events currently have no retention rule.

Rate limiting is in memory per API instance and resets on restart. Public redirects and `/api/v1` requests have separate buckets. Set `TRUSTED_PROXY_CIDRS` to the CIDR ranges of proxies that directly connect to the API before forwarded client IP headers are used; leave it empty when clients connect directly.

## Docker

`docker-compose.yml` is intentionally a **development compose**. It starts Postgres, Redis, ClickHouse, the API, and the web client for local work.

For self-hosting on a VPS with Coolify or Dokploy, deploy the app as two separate services:

- Backend: build from `server/Dockerfile`
- Frontend: build from `client/Dockerfile`

Use the repository root as the Docker build context for both services.

## Development

```bash
# Start dependencies
docker compose up postgres redis clickhouse -d

# Apply database migrations, then run the server
export POSTGRES_CONNECTION_STRING='postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable'
export REDIS_URL='localhost:6379'
export CLERK_SECRET_KEY='sk_test_...'
cd server && ./scripts/migrate-up.sh && go run ./cmd/main.go

# In another terminal, run the client
cd client && cp .env.example .env && pnpm install && pnpm dev
```

## CI

GitHub Actions runs backend tests, frontend checks, tests and builds, and OpenAPI schema validation on pushes to `main` and pull requests. PostgreSQL and Redis integration tests are opt-in locally through `TEST_DATABASE_URL` and `TEST_REDIS_ADDR`.

## License

MIT
