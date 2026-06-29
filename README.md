# link4.it — URL Shortener

A production-grade URL shortener with analytics, tags, and QR code generation. Built with Go + React, designed for solo founders and SaaS teams.

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
- Dark/light mode, responsive UI

## Quick Start

```bash
# Clone and enter the project
git clone <repo-url> && cd url-shortener

# Set Clerk API keys (get them from https://dashboard.clerk.com)
export CLERK_SECRET_KEY=sk_test_...
export VITE_CLERK_PUBLISHABLE_KEY=pk_test_...

# Start everything with a single command
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

## API Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/{shortcode}` | Redirect |
| `POST` | `/api/v1/links` | Create link |
| `GET` | `/api/v1/links` | List links |
| `GET` | `/api/v1/links/{shortcode}/analytics` | Analytics |
| `GET` | `/api/v1/links/{shortcode}/qrcode` | QR code |

Full API reference at http://localhost:8080/api/v1/reference

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

## License

MIT
