# 🚨 Error Handling Guide

A guide for implementing error handling in the URL Shortener project, following Go best practices and industry patterns.

---

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Sentinel Errors](#sentinel-errors)
3. [Error Codes](#error-codes)
4. [Handler Patterns](#handler-patterns)
5. [When to Use errors Package vs fmt.Errorf()](#when-to-use-errors-package-vs-fmterrorf)
6. [Best Practices](#best-practices)
7. [Error Catalog](#error-catalog)

---

## Architecture Overview

We use a **simple, direct approach**:

1. **Service Layer**: Returns standard Go errors (sentinel errors + wrapped errors)
2. **Handler Layer**: Checks errors with `errors.Is()` and writes HTTP responses directly

This follows Go idioms: services don't know about HTTP, handlers map errors to HTTP responses directly where they occur.

### Flow

```
Service Layer          Handler Layer          HTTP Response
─────────────          ─────────────          ────────────
LinkNotFound  →  errors.Is()  →  ErrorResponse  →  JSON Response
(sentinel error)     (check)         (DTO)            (with code)
```

**Key principle**: Handlers handle errors directly where they occur. No context passing, no middleware magic.

---

## Sentinel Errors

**Sentinel errors** are predefined error values in Go that you check for using `errors.Is()`. They're called "sentinels" because they act as markers or flags that indicate a specific error condition. Think of them as **error constants** - you define them once, and then check for them throughout your code.

### Definition

**Location**: `pkg/errors/errors.go`

```go
// Sentinel errors - use these in services, check with errors.Is()
var (
    LinkNotFound = errors.New("Link not found")
    InvalidURL   = errors.New("Invalid URL")
    InternalError = errors.New("Internal server error")
)
```

This creates **single error values** that you can check for later. The string message is just for debugging - what matters is the **identity** of the error.

### Usage in Services

In your service layer (`pkg/service/link.go`):

```go
func (s *LinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error) {
    link, err := s.queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // Wrap the sentinel error with context
            return db.Link{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
        }
        // Wrap other errors with context
        return db.Link{}, fmt.Errorf("failed to get link: %w", err)
    }
    return link, nil
}
```

**Key points:**
- Service layer doesn't know about HTTP
- Service returns Go errors (sentinel errors)
- Uses `fmt.Errorf("%w", ...)` to wrap errors with context

### Checking for Sentinel Errors

You use `errors.Is()` to check if an error is (or wraps) a sentinel error:

```go
err := service.GetLinkByID(ctx, id, userID)
if errors.Is(err, apperrors.LinkNotFound) {
    // Handle "link not found" case
}
```

**Why `errors.Is()`?**
- Works even when errors are wrapped: `fmt.Errorf("context: %w", LinkNotFound)`
- Checks the entire error chain
- Standard Go idiom

### Why Use Sentinel Errors?

#### 1. Type-Safe Error Checking

```go
// ✅ Good - type-safe, checked at compile time
if errors.Is(err, LinkNotFound) {
    // ...
}

// ❌ Bad - string comparison, error-prone
if err.Error() == "link not found" {
    // What if message changes?
}
```

#### 2. Works with Error Wrapping

Go's error wrapping (`fmt.Errorf("%w", err)`) preserves the sentinel error:

```go
// Service wraps error with context
return fmt.Errorf("failed to get link %s: %w", id, LinkNotFound)

// Handler can still check for it
if errors.Is(err, LinkNotFound) {  // ✅ Still works!
    // ...
}
```

#### 3. Separation of Concerns

- **Service layer**: Returns Go errors (sentinel errors)
- **Handler layer**: Maps Go errors to HTTP responses (error codes)

Services don't know about HTTP, handlers don't know about business logic.

#### 4. Testable

Easy to test:

```go
// In tests
mockService := &mockLinkService{
    GetLinkByIDFunc: func(...) (db.Link, error) {
        return db.Link{}, apperrors.LinkNotFound
    },
}

// Test can check
if !errors.Is(err, apperrors.LinkNotFound) {
    t.Errorf("Expected LinkNotFound")
}
```

### When to Create Sentinel Errors

Create a sentinel error when:

✅ **The error is checked in multiple places**
```go
// Used in GetLinkByID, DeleteLink, GetOriginalURL
var LinkNotFound = errors.New("Link not found")
```

✅ **The error needs special handling**
```go
// Frontend needs to show URL validation error differently
var InvalidURL = errors.New("Invalid URL")
```

✅ **The error represents a domain concept**
```go
// "Link not found" is a business concept, not just a database error
var LinkNotFound = errors.New("Link not found")
```

❌ **Don't create for one-off errors**
```go
// ❌ Bad - only used once
var ErrFailedToGenerateCode = errors.New("failed to generate code")

// ✅ Good - just return wrapped error
return fmt.Errorf("failed to generate code: %w", err)
```

### Common Patterns

#### Pattern 1: Wrapping with Context

```go
// Add context while preserving sentinel error
return fmt.Errorf("failed to get link %s: %w", id, LinkNotFound)
```

#### Pattern 2: Checking Multiple Sentinels

```go
switch {
case errors.Is(err, LinkNotFound):
    return handleNotFound()
case errors.Is(err, InvalidURL):
    return handleInvalidInput()
default:
    return handleUnknown()
}
```

#### Pattern 3: Standard Library Sentinels

```go
// Go standard library provides sentinel errors
if errors.Is(err, sql.ErrNoRows) {
    // Database "not found"
}
if errors.Is(err, os.ErrNotExist) {
    // File not found
}
```

### Sentinel Errors vs Error Codes

**Sentinel Errors (Service Layer):**
- **Purpose**: Internal Go error checking
- **Used in**: Service layer, business logic
- **Checked with**: `errors.Is(err, LinkNotFound)`
- **Example**: `LinkNotFound`, `InvalidURL`, `InternalError`
- **Type**: Go `error` interface
- **Scope**: Internal to your Go code

**Error Codes (HTTP Layer):**
- **Purpose**: Machine-readable codes for API clients (React frontend)
- **Used in**: HTTP responses, API contracts
- **Checked with**: `error.code === "link_not_found"` (in TypeScript)
- **Example**: `"link_not_found"`, `"invalid_url"`
- **Type**: `ErrorCode string` (JSON field)
- **Scope**: External API contract

**The Relationship:**

```
Service Layer (Go)              HTTP Layer (JSON)
─────────────────               ────────────────
LinkNotFound      ──────→    "link_not_found"
(sentinel error)               (error code)
     │                                │
     │                                │
  errors.Is()                    Frontend checks
  (internal)                    error.code === "..."
```

**The Flow:**
```
Service → Sentinel Error → Handler checks with errors.Is() → Error Code → Frontend
```

Both serve different purposes and work together!

---

## Error Codes

**Error codes** are machine-readable strings for API clients (React frontend). They're only defined for errors that need special frontend handling.

### Definition

**Location**: `pkg/errors/errors.go`

```go
// Error codes - only for errors that need special frontend handling
const (
    CodeLinkNotFound   ErrorCode = "link_not_found"
    CodeInvalidURL     ErrorCode = "invalid_url"
    CodeLinkExpired    ErrorCode = "link_expired"     // Future
    CodeCodeTaken      ErrorCode = "code_taken"       // Future
    CodeInternalError  ErrorCode = "internal_server_error"    // Only for unknown errors
)
```

### Design Principles

- **Only define codes for errors that need special frontend handling**
- **Generic HTTP errors (400, 401, 403, 404, 500) don't have codes** - use HTTP status directly
- **Codes should be more specific than HTTP status** (e.g., `link_not_found` vs just 404)

### JSON Response Format

**Generic HTTP error (no code):**
```json
{
  "error": {
    "message": "Invalid request body"
  }
}
```
Frontend checks: `status === 400`

**Specific error (has code):**
```json
{
  "error": {
    "code": "invalid_url",
    "message": "Invalid URL format"
  }
}
```
Frontend can handle: `if (error.code === "invalid_url") { /* show validation error */ }`

---

## Handler Patterns

Handlers check errors directly with `errors.Is()` and write HTTP responses using the same pattern as success responses.

### Pattern 1: Handler-Level Errors (JSON Decode, etc.)

For errors that occur in handlers (before calling services):

```go
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
    var reqBody dto.CreateLink
    
    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        h.logger.Warn("Invalid request body",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Message: "Invalid request body",
            },
        })
        return
    }

    // ... rest of handler
}
```

### Pattern 2: Service Errors

For errors from services, check with `errors.Is()` and handle directly:

```go
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
    // ... decode request body ...

    createdLink, err := h.LinkService.CreateShortLink(r.Context(), userID, reqBody.URL)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // Success response
    render.Status(r, http.StatusCreated)
    render.JSON(w, r, &dto.SuccessResponse[db.Link]{
        Data:    createdLink,
        Message: "Short Link created successfully",
    })
}

// handleError maps errors to HTTP responses and writes them directly
func (h *LinkHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
    switch {
    case errors.Is(err, apperrors.LinkNotFound):
        h.logger.Warn("Link not found",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusNotFound)
        render.JSON(w, r, dto.ErrorResponse{
            Error: dto.ErrorObject{
                Code:   apperrors.CodeLinkNotFound,
                Title:  apperrors.LinkNotFound.Error(),
                Detail: "Unable to find link with shortcode",
            },
        })

    case errors.Is(err, apperrors.InvalidURL):
        h.logger.Warn("Invalid URL",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, dto.ErrorResponse{
            Error: dto.ErrorObject{
                Code:   apperrors.CodeInvalidURL,
                Title:  apperrors.InvalidURL.Error(),
                Detail: "",
            },
        })

    case errors.Is(err, sql.ErrNoRows):
        h.logger.Warn("Resource not found",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusNotFound)
        render.JSON(w, r, dto.ErrorResponse{
            Error: dto.ErrorObject{
                Code:   apperrors.CodeLinkNotFound,
                Title:  "Resource not found",
                Detail: "",
            },
        })

    default:
        h.logger.Error("Internal server error",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, dto.ErrorResponse{
            Error: dto.ErrorObject{
                Code:   apperrors.CodeInternalError,
                Title:  apperrors.InternalError.Error(),
                Detail: "",
            },
        })
    }
}
```

### Pattern 3: Middleware Errors

For errors in middleware (like authentication):

```go
func RequireAuth(log logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := clerk.SessionClaimsFromContext(r.Context())

            if !ok || claims == nil {
                log.Warn("Missing or invalid session claims",
                    zap.String("method", r.Method),
                    zap.String("path", r.URL.Path),
                )
                render.Status(r, http.StatusUnauthorized)
                render.JSON(w, r, apperrors.ErrorResponse{
                    Error: apperrors.ErrorDetail{
                        Message: "Authentication required",
                    },
                })
                return
            }

            // ... continue
        })
    }
}
```

**Key points:**
- Handlers write error responses directly using `render.Status()` and `render.JSON()`
- Same pattern as success responses - no special abstractions
- Log errors appropriately (Warn for expected errors, Error for unexpected)
- Use `handleError()` helper to avoid repetition if you have multiple handlers

---

## When to Use errors Package vs fmt.Errorf()

This is a common question: when should you use `errors.New()` from the `errors` package vs `fmt.Errorf()`?

### Use `errors.New()` for Sentinel Errors

**When**: You need a predefined error value that will be checked with `errors.Is()` in multiple places.

```go
// In pkg/errors/errors.go
var (
    LinkNotFound = errors.New("Link not found")  // ✅ Sentinel error
    InvalidURL   = errors.New("Invalid URL")     // ✅ Sentinel error
)

// In service
if errors.Is(err, sql.ErrNoRows) {
    return db.Link{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
}

// In handler
if errors.Is(err, apperrors.LinkNotFound) {
    // Handle not found
}
```

**Why**: Sentinel errors are **identity-based** - you check for the specific error value, not the message. The message is just for debugging.

### Use `fmt.Errorf()` for Wrapping and Context

**When**: You need to add context to an error or wrap another error.

```go
// ✅ Good - wrap sentinel error with context
return fmt.Errorf("failed to get link %s: %w", id, apperrors.LinkNotFound)

// ✅ Good - wrap generic error with context
return fmt.Errorf("failed to create link: %w", err)

// ✅ Good - create one-off error with context
return fmt.Errorf("failed to generate code: %w", err)
```

**Why**: `fmt.Errorf()` with `%w` preserves the original error in the chain, allowing `errors.Is()` to work correctly.

### Decision Tree

```
Is this error checked in multiple places?
├─ YES → Use errors.New() (sentinel error)
│        └─ Check with errors.Is() in handlers/services
│
└─ NO → Use fmt.Errorf() (one-off error)
         └─ Just return it, no need to check for it
```

### Examples

#### ✅ Sentinel Error (errors.New)

```go
// pkg/errors/errors.go
var LinkNotFound = errors.New("Link not found")

// pkg/service/link.go
if errors.Is(err, sql.ErrNoRows) {
    return db.Link{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
}

// pkg/handlers/link.go
if errors.Is(err, apperrors.LinkNotFound) {
    // Handle not found
}
```

#### ✅ One-Off Error (fmt.Errorf)

```go
// pkg/service/link.go
code, err := generateRandomCode(9)
if err != nil {
    // Only used once, no need for sentinel
    return db.Link{}, fmt.Errorf("failed to generate code: %w", err)
}
```

#### ✅ Wrapping Standard Library Errors

```go
// pkg/service/link.go
link, err := s.queries.GetLinkByIdAndUser(ctx, params)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        // Wrap standard library error with sentinel
        return db.Link{}, fmt.Errorf("%w: %v", apperrors.LinkNotFound, err)
    }
    // Wrap other errors with context
    return db.Link{}, fmt.Errorf("failed to get link: %w", err)
}
```

### Best Practices

1. **Create sentinel errors for domain concepts** that are checked in multiple places
2. **Use `fmt.Errorf("%w", ...)` to wrap errors** - preserves error chain
3. **Don't create sentinels for one-off errors** - just use `fmt.Errorf()`
4. **Always wrap with context** - `fmt.Errorf("context: %w", err)` not just `err`

---

## Best Practices

### ✅ DO

**1. Keep errors package-level**
```go
// pkg/errors/errors.go - All error definitions in one place
```

**2. Use descriptive error codes**
```go
"link_not_found"  // ✅ Clear
"ERR_001"         // ❌ Cryptic
```

**3. Log errors appropriately**
```go
// Warn for expected errors (not found, validation errors)
        h.logger.Warn("Link not found", zap.Error(err))

// Error for unexpected errors (database failures, etc.)
h.logger.Error("Database query failed", zap.Error(err))
```

**4. Don't expose internal errors to clients**
```go
// ✅ Good - generic message
render.JSON(w, r, apperrors.ErrorResponse{
    Error: apperrors.ErrorDetail{
        Message: "An unexpected error occurred",
    },
})

// ❌ Bad - might leak database details
render.JSON(w, r, apperrors.ErrorResponse{
    Error: apperrors.ErrorDetail{
        Message: err.Error(),  // "connection to postgres failed"
    },
})
```

**5. Use consistent error response format**
```json
{
  "error": {
    "code": "link_not_found",
    "message": "Link not found"
  }
}
```

**6. Wrap errors with context**
```go
// ✅ Good
return fmt.Errorf("failed to get link %s: %w", id, LinkNotFound)

// ❌ Bad
return LinkNotFound
```

**7. Use the same pattern for errors as success responses**
```go
// Success
render.Status(r, http.StatusOK)
render.JSON(w, r, &dto.SuccessResponse[...]{...})

// Error
render.Status(r, http.StatusNotFound)
render.JSON(w, r, apperrors.ErrorResponse{...})
```

### ❌ DON'T

**1. Don't expose internal errors to clients**
```go
// ❌ Bad
http.Error(w, err.Error(), 500)  // Might leak "connection to postgres failed"

// ✅ Good
render.Status(r, http.StatusInternalServerError)
render.JSON(w, r, apperrors.ErrorResponse{
    Error: apperrors.ErrorDetail{
        Message: "An unexpected error occurred",
    },
})
```

**2. Don't create too many error types**
```go
// ❌ Bad - too granular
LinkNotFoundInDatabase
LinkNotFoundInCache
LinkNotFoundAfterRetry

// ✅ Good
LinkNotFound  // Callers don't care where it failed
```

**3. Don't create error codes for generic HTTP errors**
```go
// ❌ Bad - redundant
ErrorBadRequest ErrorCode = "BAD_REQUEST"  // HTTP 400 is enough

// ✅ Good - just use HTTP status, no code field
render.Status(r, http.StatusBadRequest)
render.JSON(w, r, apperrors.ErrorResponse{
    Error: apperrors.ErrorDetail{
        Message: "Invalid request body",
    },
})
```

**4. Don't use context to pass errors**
```go
// ❌ Bad - indirect, requires special middleware
r = apperrors.SetError(r, err)

// ✅ Good - handle directly
if err != nil {
    h.handleError(w, r, err)
    return
}
```

---

## Error Catalog

### Current Error Codes

**Note:** We only define error codes for errors that need special frontend handling. Generic HTTP errors (400, 401, 403, 404, 500) use HTTP status codes directly - no code field in JSON response.

| Code             | Status | Message                  | When                  | Notes                    |
| ---------------- | ------ | ------------------------ | --------------------- | ------------------------ |
| `link_not_found` | 404    | Link not found           | Code doesn't exist    | Specific - frontend handles differently |
| `invalid_url`    | 400    | Invalid URL format       | URL validation fails  | Specific - frontend shows validation error |
| `link_expired`   | 410    | Link has expired         | Expired timestamp     | Future: different from not found |
| `code_taken`     | 409    | Short code already taken | Custom code conflict  | Future: custom code conflicts |
| `internal_error` | 500    | Internal server error    | Unexpected errors     | Only for unknown errors |

**Generic HTTP Errors (no code field):**
- `400 Bad Request` - Invalid request body, malformed JSON, etc.
- `401 Unauthorized` - Missing/invalid authentication
- `403 Forbidden` - Access denied
- `404 Not Found` - Generic resource not found (when not link-specific)
- `500 Internal Server Error` - Server errors (when not using `internal_error` code)

---

## References

- [Go Blog: Error Handling](https://go.dev/blog/error-handling-and-go)
- [Go Blog: Working with Errors](https://go.dev/blog/go1.13-errors)
- [Dave Cheney: Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
