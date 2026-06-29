# Welcome to URL Shortener (link4.it)

This doc helps you get familiar with the codebase quickly.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, chi router, sqlc, pgx |
| Database | PostgreSQL (primary), ClickHouse (analytics) |
| Cache | Redis |
| Auth | Clerk (external OAuth/SSO) |
| Frontend | React 19, TypeScript, Vite, TanStack Router, TanStack Query |
| UI | Tailwind CSS 4, Radix UI, shadcn-style components |
| Charts | Recharts |

## Project Layout

```
url-shortener/
  server/           # Go backend
    cmd/main.go     # Entry point
    pkg/
      server.go     # Dependency wiring
      config/       # Viper-based config
      db/           # sqlc-generated code
      dto/          # Request/response types
      errors/       # Sentinel errors
      handlers/     # HTTP handlers (link, tag)
      logger/       # Zap logger
      middleware/   # Auth, validation, logging
      router/       # Route definitions
      service/      # Business logic
      analytics/    # ClickHouse client
    migrations/     # SQL migrations
    queries/        # SQL query files (sqlc input)
  client/           # React frontend
    src/
      components/   # UI components
      hooks/        # React Query hooks
      lib/          # API client, env, constants
      routes/       # TanStack Router routes
      types/        # TypeScript types
  docs/             # Documentation
  docker-compose.yml
```

## Key Architecture Decisions

- **Chi router** with middleware pattern: CORS -> RequestID -> RequestLogger -> Recoverer -> Auth -> Handlers
- **sqlc** generates Go code from SQL queries in `server/queries/`. Run `sqlc generate` to regenerate.
- **Service layer** contains all business logic, handlers are thin wrappers.
- **Sentinel errors** (`pkg/errors`) for consistent HTTP error mapping.
- **Redis cache-aside** pattern for redirect lookups, graceful degradation.
- **ClickHouse** for click analytics (optional, degrades gracefully).

## How to Run

### Quick start (Docker Compose)
```bash
# Set Clerk keys (get them from https://dashboard.clerk.com)
export CLERK_SECRET_KEY=sk_test_...
export VITE_CLERK_PUBLISHABLE_KEY=pk_test_...

docker compose up --build
```

### Manual (for development)
```bash
# Start dependencies
docker compose up postgres redis clickhouse -d

# Server
cd server
cp .env.example .env  # Fill in values
go run ./cmd/main.go

# Client
cd client
pnpm install
pnpm dev
```

## Key Features

- **Create short links** with auto-generated 7-char shortcodes (SHA-256 + base62, deterministic per URL+user) or custom shortcodes
- **Tag management** - create, assign, filter links by tags
- **Click analytics** via ClickHouse (clicks over time, top referrers, user agents)
- **QR code generation** per link
- **Redis caching** for fast redirects
- **Clerk authentication** (Google/GitHub OAuth)
- **Responsive UI** with dark/light mode

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/{shortcode}` | Redirect to original URL |
| POST | `/api/v1/links` | Create short link |
| GET | `/api/v1/links` | List links (paginated, filterable) |
| GET | `/api/v1/links/{shortcode}` | Get link details |
| PATCH | `/api/v1/links/{id}` | Update link |
| DELETE | `/api/v1/links/{id}` | Delete link |
| GET | `/api/v1/links/{shortcode}/analytics` | Get click analytics |
| GET | `/api/v1/links/{shortcode}/qrcode` | Get QR code PNG |
| POST | `/api/v1/links/{id}/tags` | Add tags to link |
| POST | `/api/v1/links/{id}/tags/remove` | Remove tags from link |
| GET/POST/PATCH/DELETE | `/api/v1/tags` | Tag CRUD |
| POST | `/api/v1/tags/bulk-delete` | Bulk delete tags |

## Key Details

### Shortcode Generation
Auto-generated shortcodes are **deterministic** — derived from SHA-256 of the URL + user ID, then base62-encoded to 7 characters. Same URL by the same user always gets the same shortcode (deduplication). Length 7 gives 62^7 ≈ 3.5 trillion combinations. Custom shortcodes are user-provided and go through the same uniqueness constraint.

## Development Patterns

### Adding a new endpoint
1. Add SQL query in `server/queries/`
2. Run `sqlc generate`
3. Add service method in `server/pkg/service/`
4. Add handler in `server/pkg/handlers/`
5. Register route in `server/pkg/router/router.go`

### Error handling
- Define sentinel error in `server/pkg/errors/errors.go`
- Return from service using `fmt.Errorf("%w: ...", apperrors.SomeError)`
- Add `errors.Is()` case in handler's `handleError()`

### Testing
- `task test` runs all Go tests
- Service tests use `mockQueries` struct
- Handler tests use `mockLinkService` struct
