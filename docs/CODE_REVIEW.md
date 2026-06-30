# Backend Code Review

**Date:** 2026-06-30  
**Scope:** Complete backend codebase review (fresh)

---

## Executive Summary

The codebase demonstrates solid architecture with clear separation of concerns. Authentication bypass issues from the previous review have been resolved. However, several critical security issues, correctness bugs, and code quality problems remain.

**Overall Assessment:** ⚠️ **Good foundation, but needs critical fixes before production**

---

## 🔴 Critical Issues

### 1. Host Header Injection in QR Code Generation

**Location:** `pkg/handlers/link.go:399`  
**Severity:** CRITICAL

```go
shortURL := fmt.Sprintf("%s/%s", r.Host, shortcode)
```

`r.Host` is client-controlled. A malicious user can set `Host: evil.com` and generate a QR code pointing to an attacker-controlled domain. This is a phishing vector.

**Fix:** Use a configured base URL from config instead of the request Host header.

---

### 2. ClickHouse Write Blocks Redirect Synchronously

**Location:** `pkg/handlers/link.go:85` → `pkg/service/link.go:371`  
**Severity:** CRITICAL

`RecordClick` performs a synchronous ClickHouse INSERT with a 3-second timeout. Every redirect request blocks on this write. This creates a DoS vector against the redirect endpoint.

**Fix:** Run `RecordClick` in a goroutine with a detached context.

---

### 3. Database Credentials Logged in Plaintext

**Location:** `pkg/server.go:55, 79, 84`  
**Severity:** CRITICAL

```go
zap.String("pg_connection_str", config.PostgresConnectionString)
zap.String("redis_url", config.RedisURL)
```

The full PostgreSQL connection string (containing username/password) and Redis URL (potentially `redis://user:pass@host`) are logged at Info level.

**Fix:** Sanitize connection strings before logging (redact credentials).

---

### 4. Analytics Error Silently Discarded

**Location:** `pkg/server.go:91`  
**Severity:** CRITICAL

```go
s.AnalyticsClient, _ = analytics.New(...)
```

If ClickHouse is configured but unreachable, the error is silently swallowed with `_`. The app continues with a nil analytics client without any warning.

**Fix:** Log the error as a warning at minimum.

---

### 5. Open Redirect via Inactive/Expired Links

**Location:** `pkg/handlers/link.go:61-88`  
**Severity:** HIGH

The `Redirect` handler calls `GetOriginalURL` which does NOT check `is_active` or `expires_at`. Expired or deactivated links still redirect. The error code `CodeLinkExpired` exists but is never used in this path.

**Fix:** The `GetLinkForRedirect` SQL query already filters for `is_active=true` and non-expired. Verify this query is the one truly being used (it is in `GetOriginalURL` → `GetLinkForRedirect`). If correct, mark as resolved. If not, add the check.

---

### 6. IP Address Includes Port in Analytics

**Location:** `pkg/service/link.go:372`  
**Severity:** HIGH

```go
s.analyticsClient.RecordClick(ctx, analytics.ClickEvent{
    IP: ip,  // r.RemoteAddr is "192.168.1.1:54321"
})
```

`r.RemoteAddr` includes the port. The full string is stored in ClickHouse, making IP-based analytics incorrect.

**Fix:** Use `net.SplitHostPort(r.RemoteAddr)` to extract the IP.

---

## 🟡 High Severity

### 7. Hardcoded PostgreSQL Error Code

**Location:** `pkg/service/tag.go:68, 93` and `pkg/service/link.go:454`  
**Severity:** HIGH

```go
pgErr.Code == "23505"
```

The magic string `"23505"` (unique_violation) is repeated in 3 places.

**Fix:** Define a package-level constant `const pgUniqueViolation = "23505"`.

---

### 8. Redis Credentials Marked `required` in Config

**Location:** `pkg/config/config.go:22-23`  
**Severity:** HIGH

```go
RedisUsername string `mapstructure:"REDIS_USERNAME" validate:"required"`
RedisPassword string `mapstructure:"REDIS_PASSWORD" validate:"required"`
```

Redis often runs without auth in development/Docker. The app refuses to start unnecessarily.

**Fix:** Change to `validate:"omitempty"`.

---

### 9. Minimal Linter Configuration

**Location:** `.golangci.yml`  
**Severity:** HIGH

Only 5 linters enabled. Critically missing:
- `gosec` — would catch Host header injection, credential logging
- `gosimple` — code simplification suggestions
- `gocritic` — performance and code smells
- `misspell` — catches typos
- `goconst` — finds repeated strings that should be constants

**Fix:** Enable additional linters.

---

### 10. UUID Parsing Duplicated Across 5+ Handlers

**Location:** `pkg/handlers/link.go` and `pkg/handlers/tag.go`  
**Severity:** MEDIUM

The 16-line UUID parsing/validation block is copy-pasted 5 times across handlers.

**Fix:** Extract to shared helper function.

---

### 11. Invalid Tag UUIDs Silently Dropped

**Location:** `pkg/handlers/link.go:150-157`  
**Severity:** MEDIUM

Invalid UUIDs in `?tags=` query param are logged and silently skipped. The user receives a 200 OK with the filter partially applied.

**Fix:** Return 400 Bad Request with details about invalid UUIDs.

---

### 12. Invalid Pagination Values Silently Clamped

**Location:** `pkg/handlers/link.go:165-174`  
**Severity:** MEDIUM

Non-numeric `page` and `limit` values silently default to 1 and 5 without warning.

**Fix:** Log a warning when defaults are used due to invalid input.

---

### 13. Tag Handler Missing `sql.ErrNoRows` Check

**Location:** `pkg/handlers/tag.go:172-218`  
**Severity:** MEDIUM

`TagHandler.handleError` does not check for `sql.ErrNoRows`. If the service layer fails to wrap it in a sentinel, the handler returns 500 instead of 404.

**Fix:** Add `sql.ErrNoRows` check to tag handler error handling.

---

### 14. Bulk Delete Returns `null` Instead of `[]`

**Location:** `pkg/handlers/tag.go:166`  
**Severity:** LOW

If the DB query returns nil, JSON output will be `"data": null` instead of `"data": []`.

**Fix:** Normalize nil to empty slice.

---

### 15. Duplicate Validation in DTO and Service

**Location:** `pkg/dto/link.go:30` and `pkg/service/link.go:117`  
**Severity:** MEDIUM

The future-date check for `expires_at` runs in both the DTO validation middleware and the service layer. Creates two error paths for the same input.

**Fix:** Keep validation in one layer only.

---

### 16. DTO Receiver Type Inconsistency

**Location:** `pkg/dto/link.go:25` vs `pkg/dto/tag.go:17`  
**Severity:** LOW

`UpdateLink.Validate()` uses value receiver, `CreateTag.Validate()` uses pointer receiver.

**Fix:** Standardize on pointer receivers for all `Validate()` methods.

---

### 17. `BypassAuth` Function Still Exists

**Location:** `pkg/middleware/auth.go`  
**Severity:** MEDIUM

The `BypassAuth` middleware function exists in the codebase even though it's not used in the router. Dead code that could be accidentally re-enabled.

**Fix:** Remove the function entirely.

---

### 18. TODO Comments in Production Code

**Locations:**
- `cmd/main.go:58` — "Need to get more comfortable with what this does"
- `pkg/server.go:70` — "Need to understand this better"

**Severity:** LOW

These should be resolved or replaced with proper documentation.

---

### 19. Missing Transaction Management

**Location:** `pkg/service/link.go` — `AddTagsToLink`, `RemoveTagsFromLink`  
**Severity:** MEDIUM

Multi-step operations (add tags → fetch link) don't use database transactions.

**Fix:** Use `queries.WithTx(tx)` for atomic operations.

---

### 20. No ClickHouse Timeout Configuration

**Location:** `pkg/analytics/clickhouse.go:57`  
**Severity:** LOW

DialTimeout, MaxOpenConns, etc. are hardcoded instead of configurable.

**Fix:** Add ClickHouse config fields to `Config` struct.

---

### 21. Goroutine Shutdown Race Condition

**Location:** `cmd/main.go:51-69`  
**Severity:** LOW

The shutdown goroutine starts before `ListenAndServe`. If the server fails to start, `Shutdown` is called on a server that never ran.

**Fix:** Start the shutdown goroutine after confirming `ListenAndServe` started.

---

### 22. Missing `ErrForbidden` Sentinel

**Location:** `pkg/errors/errors.go`  
**Severity:** LOW

No `ErrForbidden` or `CodeForbidden` for authorization failures (distinct from auth). If a user tries to access another user's resource, there's no appropriate error.

---

### 23. Exported Context Key Function

**Location:** `pkg/middleware/request_validator.go:27`  
**Severity:** LOW

```go
func ReqBodyKey() contextKey { return reqBodyKey }
```

Exposing the context key is unnecessary and couples tests to implementation.

---

## Infrastructure

### 24. Docker Compose Deficiency

**Location:** `docker-compose.yml`  
**Severity:** MEDIUM

- ClickHouse image uses `latest` tag instead of a pinned version
- Client depends on server but has no health check
- No network isolation between services
- Environment variables `REDIS_USERNAME` and `REDIS_PASSWORD` set to empty strings, causing config validation to reject them (`required` tag)

### 25. Dockerfile Multi-stage Optimization

**Location:** `server/Dockerfile`  
**Severity:** LOW

The builder stage copies `docs/` into the build context but the server binary doesn't need docs at build time. This invalidates Docker layer cache when docs change.

**Fix:** Separate the build context better or use multi-stage more carefully.

### 26. No `.dockerignore` Files

**Severity:** LOW

Neither the server nor client directories have `.dockerignore` files, which means the entire directory is sent to the Docker daemon as build context.

---

## 🟢 What's Working Well

- Clean handler → service → database layering
- Interface-based design enables testing
- Sentinel errors with consistent HTTP mapping
- Graceful shutdown with SIGINT/SIGTERM handling
- Graceful degradation for Redis and ClickHouse
- Structured logging with zap
- Request validation middleware
- CORS configuration
- SQLC-generated type-safe database code
- Migration scripts for schema changes
- Context propagation for cancellation/timeouts

---

## 📊 Summary

| Category | Status | Key Issues |
|----------|--------|------------|
| Security | ⚠️ Needs Fixes | Host header injection, credential logging, DoS via ClickHouse |
| Correctness | ⚠️ Issues Found | IP includes port, expired links redirect, UUID parsing duplication |
| Error Handling | ✅ Good | Missing sql.ErrNoRows in tag handler, bulk-delete returns null |
| Code Quality | ⚠️ Needs Cleanup | TODOs, magic strings, inconsistent receiver types |
| Performance | ⚠️ One Critical | Synchronous ClickHouse on redirect path |
| Infrastructure | ⚠️ Needs Work | Linter config minimal, Docker compose improvements |
| Architecture | ✅ Excellent | Proper layering, interface design, graceful degradation |

---

## 🎯 Recommended Action Plan

### Immediate (security/correctness)
1. Fix Host header injection in QR code generation
2. Make ClickHouse writes async in redirect handler
3. Sanitize connection strings before logging
4. Extract IP from RemoteAddr before storing
5. Log analytics connection errors

### Short Term
6. Extract UUID parsing helper
7. Add missing linters (gosec, gosimple, gocritic)
8. Make Redis creds optional in config
9. Add transaction support for multi-step operations
10. Remove `BypassAuth` dead code
11. Resolve TODO comments

### Medium Term
12. Add integration/E2E tests
13. Configure DB connection pool
14. Add Redis health checks
15. Pin Docker image versions
16. Add `.dockerignore` files
