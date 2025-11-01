# 🚨 Error Handling Guide - Production Grade

A comprehensive guide for implementing production-grade error handling in the URL Shortener project.

---

## 📋 Table of Contents

1. [Current State (MVP)](#current-state-mvp)
2. [When to Add Custom Errors](#when-to-add-custom-errors)
3. [Custom Error Patterns](#custom-error-patterns)
4. [Migration Strategy](#migration-strategy)
5. [Best Practices](#best-practices)
6. [Error Catalog](#error-catalog)

---

## Current State (MVP)

**What we have now:**

```go
// Service layer
func (s *LinkService) CreateShortLink(...) (db.Link, error) {
    // ...
    if err != nil {
        return db.Link{}, fmt.Errorf("failed to create link: %w", err)
    }
}
```

**Pros:**

- Simple, quick to implement
- Standard library only
- Good enough for MVP

**Cons:**

- Handlers must inspect wrapped errors (e.g., `errors.Is(err, sql.ErrNoRows)`)
- No consistent error response format
- Hard to add metadata (error codes, user messages, etc.)

---

## When to Add Custom Errors

Add custom errors when you need:

✅ **Consistent API error responses**

- Standardized error codes (e.g., `LINK_NOT_FOUND`, `INVALID_URL`)
- User-friendly messages separate from internal details

✅ **Rich error metadata**

- HTTP status codes
- Validation field names
- Retry information

✅ **Better handler logic**

- Stop checking `errors.Is(err, sql.ErrNoRows)` everywhere
- One place to map errors → HTTP responses

✅ **Client SDK generation**

- Typed error responses for API clients

---

## Custom Error Patterns

### Pattern 1: Sentinel Errors (Simplest)

**When to use:** Simple, well-known errors without metadata.

```go
// pkg/errors/errors.go
package errors

import "errors"

var (
    ErrLinkNotFound     = errors.New("link not found")
    ErrInvalidURL       = errors.New("invalid URL")
    ErrCodeTaken        = errors.New("code already taken")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrLinkExpired      = errors.New("link expired")
)
```

**Usage:**

```go
// Service
func (s *LinkService) GetLink(ctx context.Context, code string) (db.Link, error) {
    link, err := s.queries.GetLinkForRedirect(ctx, code)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.Link{}, apperrors.ErrLinkNotFound
        }
        return db.Link{}, fmt.Errorf("failed to get link: %w", err)
    }

    if link.ExpiresAt.Valid && link.ExpiresAt.Time.Before(time.Now()) {
        return db.Link{}, apperrors.ErrLinkExpired
    }

    return link, nil
}

// Handler
func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
    link, err := h.service.GetLink(r.Context(), code)
    if err != nil {
        switch {
        case errors.Is(err, apperrors.ErrLinkNotFound):
            http.Error(w, "Link not found", http.StatusNotFound)
        case errors.Is(err, apperrors.ErrLinkExpired):
            http.Error(w, "Link expired", http.StatusGone)
        default:
            h.logger.Error("get link failed", "error", err)
            http.Error(w, "Internal error", http.StatusInternalServerError)
        }
        return
    }
    json.NewEncoder(w).Encode(link)
}
```

---

### Pattern 2: Error Types with Metadata (Recommended)

**When to use:** When you need HTTP status, error codes, or validation details.

```go
// pkg/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

// AppError represents an application-level error with metadata
type AppError struct {
    Code       string // Machine-readable code (LINK_NOT_FOUND)
    Message    string // User-friendly message
    StatusCode int    // HTTP status code
    Internal   error  // Original error (not exposed to clients)
}

func (e *AppError) Error() string {
    if e.Internal != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Internal)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Internal
}

// Constructor functions
func NewLinkNotFound(code string) *AppError {
    return &AppError{
        Code:       "LINK_NOT_FOUND",
        Message:    fmt.Sprintf("Link with code '%s' not found", code),
        StatusCode: http.StatusNotFound,
    }
}

func NewInvalidURL(url string) *AppError {
    return &AppError{
        Code:       "INVALID_URL",
        Message:    fmt.Sprintf("URL '%s' is invalid", url),
        StatusCode: http.StatusBadRequest,
    }
}

func NewInternalError(err error) *AppError {
    return &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "An internal error occurred",
        StatusCode: http.StatusInternalServerError,
        Internal:   err,
    }
}

func NewUnauthorized() *AppError {
    return &AppError{
        Code:       "UNAUTHORIZED",
        Message:    "Authentication required",
        StatusCode: http.StatusUnauthorized,
    }
}

func NewLinkExpired() *AppError {
    return &AppError{
        Code:       "LINK_EXPIRED",
        Message:    "This link has expired",
        StatusCode: http.StatusGone,
    }
}
```

**Usage:**

```go
// Service
func (s *LinkService) GetLink(ctx context.Context, code string) (db.Link, error) {
    link, err := s.queries.GetLinkForRedirect(ctx, code)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.Link{}, apperrors.NewLinkNotFound(code)
        }
        return db.Link{}, apperrors.NewInternalError(err)
    }

    if link.ExpiresAt.Valid && link.ExpiresAt.Time.Before(time.Now()) {
        return db.Link{}, apperrors.NewLinkExpired()
    }

    return link, nil
}

// Handler (simple!)
func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
    link, err := h.service.GetLink(r.Context(), code)
    if err != nil {
        h.handleError(w, err)
        return
    }
    json.NewEncoder(w).Encode(link)
}

// Error handler middleware
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        // Application error - return structured response
        h.logger.Error("request failed",
            "code", appErr.Code,
            "message", appErr.Message,
            "internal", appErr.Internal,
        )

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(appErr.StatusCode)
        json.NewEncoder(w).Encode(map[string]string{
            "error": appErr.Code,
            "message": appErr.Message,
        })
        return
    }

    // Unknown error - return generic 500
    h.logger.Error("unexpected error", "error", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
}
```

---

### Pattern 3: Validation Errors (Advanced)

**When to use:** Form/input validation with field-level errors.

```go
// pkg/errors/validation.go
package errors

import (
    "fmt"
    "net/http"
    "strings"
)

type ValidationError struct {
    Fields map[string]string // field -> error message
}

func NewValidationError() *ValidationError {
    return &ValidationError{
        Fields: make(map[string]string),
    }
}

func (e *ValidationError) Add(field, message string) {
    e.Fields[field] = message
}

func (e *ValidationError) HasErrors() bool {
    return len(e.Fields) > 0
}

func (e *ValidationError) Error() string {
    var msgs []string
    for field, msg := range e.Fields {
        msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
    }
    return strings.Join(msgs, "; ")
}

func (e *ValidationError) ToAppError() *AppError {
    return &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        StatusCode: http.StatusBadRequest,
        Internal:   e,
    }
}
```

**Usage:**

```go
// Service
func (s *LinkService) ValidateCreateLink(originalURL string) error {
    valErr := apperrors.NewValidationError()

    if originalURL == "" {
        valErr.Add("original_url", "URL is required")
    }

    if !isValidURL(originalURL) {
        valErr.Add("original_url", "URL format is invalid")
    }

    if len(originalURL) > 2048 {
        valErr.Add("original_url", "URL is too long (max 2048 characters)")
    }

    if valErr.HasErrors() {
        return valErr.ToAppError()
    }

    return nil
}

// Handler response
{
    "error": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
        "original_url": "URL format is invalid"
    }
}
```

---

## Migration Strategy

### Phase 1: Add Error Package (Now)

```bash
mkdir -p server/pkg/errors
touch server/pkg/errors/errors.go
```

Add basic sentinel errors or AppError type.

### Phase 2: Update Services Gradually

Start with the most common errors:

1. `ErrLinkNotFound` - Used in redirect endpoint
2. `ErrUnauthorized` - Used in authenticated endpoints
3. `ErrInvalidURL` - Used in create link

### Phase 3: Update Handlers

Add centralized error handler:

```go
func (h *Handler) handleError(w http.ResponseWriter, err error)
```

### Phase 4: Add Middleware

```go
func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Catch panics and convert to errors
    })
}
```

---

## Best Practices

### ✅ DO

**1. Keep errors package-level**

```go
// pkg/errors/errors.go - All error definitions in one place
```

**2. Use descriptive error codes**

```go
"LINK_NOT_FOUND"        // ✅ Clear
"ERR_001"               // ❌ Cryptic
```

**3. Separate user messages from internal details**

```go
AppError{
    Message:  "Link not found",           // User sees this
    Internal: fmt.Errorf("query failed"), // Logs only
}
```

**4. Log internal errors, hide from users**

```go
h.logger.Error("db error", "internal", appErr.Internal)
// Don't send database details to client!
```

**5. Use consistent error response format**

```json
{
	"error": "LINK_NOT_FOUND",
	"message": "Link with code 'abc123' not found"
}
```

### ❌ DON'T

**1. Don't expose internal errors to clients**

```go
// ❌ Bad
http.Error(w, err.Error(), 500)  // Might leak "connection to postgres failed"

// ✅ Good
http.Error(w, "Internal server error", 500)
```

**2. Don't create too many error types**

```go
// ❌ Bad - too granular
ErrLinkNotFoundInDatabase
ErrLinkNotFoundInCache
ErrLinkNotFoundAfterRetry

// ✅ Good
ErrLinkNotFound  // Callers don't care where it failed
```

**3. Don't ignore error context**

```go
// ❌ Bad
return ErrLinkNotFound

// ✅ Good
return fmt.Errorf("failed to get link %s: %w", code, ErrLinkNotFound)
```

---

## Error Catalog

### Planned Error Codes

| Code             | Status | Message                  | When                  |
| ---------------- | ------ | ------------------------ | --------------------- |
| `LINK_NOT_FOUND` | 404    | Link not found           | Code doesn't exist    |
| `LINK_EXPIRED`   | 410    | Link has expired         | Expired timestamp     |
| `INVALID_URL`    | 400    | Invalid URL format       | URL validation fails  |
| `CODE_TAKEN`     | 409    | Short code already taken | Custom code conflict  |
| `UNAUTHORIZED`   | 401    | Authentication required  | Missing/invalid auth  |
| `FORBIDDEN`      | 403    | Access denied            | User doesn't own link |
| `RATE_LIMITED`   | 429    | Too many requests        | Rate limit exceeded   |
| `INTERNAL_ERROR` | 500    | Internal server error    | Unexpected errors     |

---

## Example: Full Implementation

```go
// pkg/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

type AppError struct {
    Code       string
    Message    string
    StatusCode int
    Internal   error
}

func (e *AppError) Error() string {
    if e.Internal != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Internal)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Internal
}

// Constructors
func NewLinkNotFound(code string) *AppError {
    return &AppError{
        Code:       "LINK_NOT_FOUND",
        Message:    fmt.Sprintf("Link '%s' not found", code),
        StatusCode: http.StatusNotFound,
    }
}

func NewInternalError(err error) *AppError {
    return &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "An internal error occurred",
        StatusCode: http.StatusInternalServerError,
        Internal:   err,
    }
}

// pkg/service/link.go
func (s *LinkService) GetOriginalURL(ctx context.Context, code string) (string, error) {
    link, err := s.queries.GetLinkForRedirect(ctx, code)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", apperrors.NewLinkNotFound(code)
        }
        return "", apperrors.NewInternalError(err)
    }
    return link.OriginalUrl, nil
}

// pkg/handlers/link.go
func (h *Handler) RedirectLink(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")

    url, err := h.service.GetOriginalURL(r.Context(), code)
    if err != nil {
        h.handleError(w, err)
        return
    }

    http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(appErr.StatusCode)
        json.NewEncoder(w).Encode(map[string]string{
            "error":   appErr.Code,
            "message": appErr.Message,
        })
        return
    }

    http.Error(w, "Internal server error", http.StatusInternalServerError)
}
```

---

## References

- [Go Blog: Error Handling](https://go.dev/blog/error-handling-and-go)
- [Go Blog: Working with Errors](https://go.dev/blog/go1.13-errors)
- [Dave Cheney: Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)

---

## Next Steps

1. ✅ Keep using `fmt.Errorf` for MVP
2. ⏳ Create `pkg/errors` when adding handlers
3. ⏳ Migrate to AppError when adding API endpoints
4. ⏳ Add validation errors when building forms

**Remember:** Start simple, add complexity when needed!
