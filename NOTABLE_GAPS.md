# Notable Gaps

This file documents known gaps and deferred work discovered during codebase exploration.

## Stub Pages

- **Dashboard** (`/`) — implemented (aggregate stats, chart, recent links)
- **Account** (`/account`) — implemented (Clerk profile, email, auth providers)
- **Billing** (`/billing`) — placeholder content
- **Settings** (`/settings`) — placeholder content
- **Feedback** (`/feedback`) — removed; feedback collected externally

## Missing Production Features

- **Rate limiting** — no request throttling anywhere
- **Background job queue** — no async job infrastructure for tasks like link expiration, email reports
- **Public API** — no public-facing API with API keys/rate limiting
- **Analytics data export** — no CSV/JSON export of analytics data
- **Custom link previews** — Open Graph tag support not started
- **Custom domains** — not started
- **Team workspaces** — not started
- **Password-protected links** — not started

## Infrastructure & Quality

- **No integration/E2E tests** — all tests are unit tests with mocks; no real DB/Redis/ClickHouse tests
- **No ClickHouse mock in unit tests** — analytics methods are skipped when `analyticsClient == nil`
- **No proper DB migration tool** — migrations run via handwritten shell scripts instead of a framework (golang-migrate, goose, etc.)
- **`server/main` binary** — was committed, now removed from git tracking
- **No CI/CD** — no deployment scripts or pipeline configuration
- **No monitoring/alerting** — no structured metrics, health check webhooks, or error alerting

## Code Quality / Technical Debt

- **Mock data still used** — `mock-data.ts` is imported by some frontend components, suggesting some areas may still use mock data in edge cases
- **QR code endpoint behind auth** — the QR code endpoint (`{shortcode}/qrcode`) is mounted inside the auth-required `/api/v1` group, while the redirect endpoint is public. QR codes for public sharing require auth.
- **TODOs in production code** — `cmd/main.go:58` and `pkg/server.go:70` have developer learning notes
- **ClickHouse database hardcoded** — `CREATE TABLE IF NOT EXISTS link4it.click_events` in `clickhouse.go:79` hardcodes the database name
