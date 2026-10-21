# link4.it — URL Shortener

A URL shortener MVP with analytics, tags, and QR code generation. Built with Go and React.

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

## Quick Start

```bash
# Clone and enter the project
git clone <repo-url> && cd url-shortener

# Set Clerk API keys (get them from https://dashboard.clerk.com)
export CLERK_SECRET_KEY=sk_test_...
export VITE_CLERK_PUBLISHABLE_KEY=pk_test_...

# Start the local development stack
docker compose up --build
```

Open http://localhost:5173, sign in with Google/GitHub, and start shortening URLs.

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
| `GET` | `/api/v1/links/{shortcode}/analytics` | Analytics |
| `GET` | `/api/v1/links/{shortcode}/qrcode` | QR code |

The API reference page is at http://localhost:8080/api/v1/reference; loading its OpenAPI document in a running stack still needs verification.

Operational headers:

- `X-Request-ID` is returned on every response and included in production logs.
- `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After` describe rate-limit state.

Shortcodes use 1–20 ASCII letters, digits, hyphens or underscores. `api` is reserved. A generated shortcode begins with a deterministic 7-character code; a collision (including a repeated URL for the same user) retries with a random code. Soft-deleted shortcodes can be reused.

Redis redirect entries live for at most one minute and never past a link's expiration. Updates and deletes invalidate them on a best-effort basis. If Redis is unavailable during invalidation, a cached redirect may remain usable until its TTL ends; immediate revocation is not guaranteed. Redirects fall back to PostgreSQL when Redis is unavailable. ClickHouse analytics are best-effort and queued events may be lost if shutdown exceeds five seconds.

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

# Run server (with hot reload)
cd server && cp .env.example .env
task run

# Run client
cd client && pnpm install && pnpm dev
```

## CI

GitHub Actions runs backend tests, frontend checks, tests and builds, and OpenAPI YAML validation on pushes to `main` and pull requests.

## License

MIT
