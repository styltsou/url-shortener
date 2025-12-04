# Handler Layer Patterns

## Overview

Handlers are the **HTTP layer** - they handle HTTP requests and responses, but delegate business logic to services. This guide explains how to create and structure handlers.

## Handler Responsibilities

A handler should:
- ✅ Parse HTTP requests (JSON, query params, path params)
- ✅ Validate request format (syntax, not business rules)
- ✅ Convert between DTOs and domain models
- ✅ Call services for business logic
- ✅ Handle errors directly and write HTTP responses
- ✅ Render HTTP responses

A handler should NOT:
- ❌ Contain business logic (that's in services)
- ❌ Access the database directly (use services)
- ❌ Know about database models (use DTOs)

## Handler Structure

### Basic Pattern

```go
// pkg/handlers/link.go

// LinkServiceInterface defines what handlers need from services
type LinkServiceInterface interface {
    CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error)
    ListAllLinks(ctx context.Context, userID string) ([]db.Link, error)
    // ... other methods
}

type LinkHandler struct {
    LinkService LinkServiceInterface  // Interface, not concrete type!
    logger      logger.Logger
}

func NewLinkHandler(linkService *service.LinkService, logger logger.Logger) *LinkHandler {
    return &LinkHandler{
        LinkService: linkService,  // service.LinkService implements LinkServiceInterface
        logger:      logger,
    }
}
```

### Why Interfaces?

- **Testability**: Mock the service in tests
- **Flexibility**: Swap implementations (caching, different services)
- **Dependency Inversion**: Handler depends on abstraction

See `INTERFACES_GUIDE.md` for details.

## Creating a New Handler

### Step 1: Define the Service Interface

```go
// pkg/handlers/user.go

type UserServiceInterface interface {
    GetUserProfile(ctx context.Context, userID string) (db.User, error)
    UpdateUserProfile(ctx context.Context, userID string, updates UpdateUserParams) (db.User, error)
}

type UserHandler struct {
    UserService UserServiceInterface
    logger      logger.Logger
}

func NewUserHandler(userService *service.UserService, logger logger.Logger) *UserHandler {
    return &UserHandler{
        UserService: userService,
        logger:      logger,
    }
}
```

### Step 2: Implement Handler Methods

```go
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    // 1. Extract userID from context (set by RequireAuth middleware)
    userID := middleware.GetUserID(r.Context())

    // 2. Call service
    user, err := h.UserService.GetUserProfile(r.Context(), userID)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 3. Render response
    render.Status(r, http.StatusOK)
    render.JSON(w, r, &dto.SuccessResponse[db.User]{
        Data:    user,
        Message: "User profile retrieved successfully",
    })
}
```

## Common Patterns

### Pattern 1: Create (POST)

```go
// POST /api/v1/links
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request body
    var reqBody dto.CreateLinkRequest
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

    // 2. Extract userID from context
    userID := middleware.GetUserID(r.Context())

    // 3. Call service
    createdLink, err := h.LinkService.CreateShortLink(r.Context(), userID, reqBody.URL)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 4. Log success
    h.logger.Info("Short link created successfully",
        zap.String("user_id", userID),
        zap.String("link_id", createdLink.ID.String()),
        zap.String("short_code", createdLink.Shortcode),
    )

    // 5. Render response
    render.Status(r, http.StatusCreated)
    render.JSON(w, r, &dto.SuccessResponse[db.Link]{
        Data:    createdLink,
        Message: "Short Link created successfully",
    })
}
```

### Pattern 2: List (GET)

```go
// GET /api/v1/links
func (h *LinkHandler) ListLinks(w http.ResponseWriter, r *http.Request) {
    // 1. Extract userID
    userID := middleware.GetUserID(r.Context())

    // 2. Call service
    links, err := h.LinkService.ListAllLinks(r.Context(), userID)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 3. Empty slice is valid - not an error
    if len(links) == 0 {
        h.logger.Info("User has no links", zap.String("user_id", userID))
    } else {
        h.logger.Info("User links retrieved successfully",
            zap.String("user_id", userID),
            zap.Int("link_count", len(links)),
        )
    }

    // 4. Render response
    render.Status(r, http.StatusOK)
    render.JSON(w, r, &dto.SuccessResponse[[]db.Link]{
        Data:    links,
        Message: "User links retrieved successfully",
    })
}
```

### Pattern 3: Get by ID (GET with path param)

```go
// GET /api/v1/links/{id}
func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
    // 1. Extract path parameter
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        h.logger.Warn("Invalid link ID format",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Message: "Invalid link ID format",
            },
        })
        return
    }

    // 2. Extract userID
    userID := middleware.GetUserID(r.Context())

    // 3. Call service
    link, err := h.LinkService.GetLinkByID(r.Context(), id, userID)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 4. Render response
    render.Status(r, http.StatusOK)
    render.JSON(w, r, &dto.SuccessResponse[db.Link]{
        Data:    link,
        Message: "Link retrieved successfully",
    })
}
```

### Pattern 4: Update (PATCH)

```go
// PATCH /api/v1/links/{id}
func (h *LinkHandler) UpdateLink(w http.ResponseWriter, r *http.Request) {
    // 1. Extract path parameter
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        h.logger.Warn("Invalid link ID format",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Message: "Invalid link ID format",
            },
        })
        return
    }

    // 2. Decode request body
    var reqBody dto.UpdateLinkRequest
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

    // 3. Extract userID
    userID := middleware.GetUserID(r.Context())

    // 4. Convert DTO to service params
    updates := service.UpdateLinkParams{
        Shortcode: reqBody.Shortcode,
        ExpiresAt: reqBody.ExpiresAt,
    }

    // 5. Call service
    updatedLink, err := h.LinkService.UpdateLink(r.Context(), id, userID, updates)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 6. Render response
    render.Status(r, http.StatusOK)
    render.JSON(w, r, &dto.SuccessResponse[db.Link]{
        Data:    updatedLink,
        Message: "Link updated successfully",
    })
}
```

### Pattern 5: Delete (DELETE)

```go
// DELETE /api/v1/links/{id}
func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
    // 1. Extract path parameter
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        h.logger.Warn("Invalid link ID format",
            zap.Error(err),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Message: "Invalid link ID format",
            },
        })
        return
    }

    // 2. Extract userID
    userID := middleware.GetUserID(r.Context())

    // 3. Call service
    err = h.LinkService.DeleteLink(r.Context(), id, userID)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 4. Render response (no body for DELETE)
    render.Status(r, http.StatusNoContent)
}
```

### Pattern 6: Public Endpoint (No Auth)

```go
// GET /{code} - Public redirect
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
    // 1. Extract path parameter
    code := chi.URLParam(r, "code")
    if code == "" {
        h.logger.Warn("Short code is required",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Message: "Short code is required",
            },
        })
        return
    }

    // 2. Call service (no userID needed)
    link, err := h.LinkService.GetOriginalURL(r.Context(), code)
    if err != nil {
        h.handleError(w, r, err)
        return
    }

    // 3. Redirect
    http.Redirect(w, r, link.OriginalUrl, http.StatusFound)
}
```

## Error Handling

### Pattern: Handle Errors Directly

Handlers check errors with `errors.Is()` and write HTTP responses directly using the same pattern as success responses.

```go
// ✅ GOOD: Handle error directly where it occurs
if err != nil {
    h.handleError(w, r, err)
    return
}

// For handler-level errors (JSON decode, validation)
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
```

### Pattern: Use handleError Helper

For service errors, use a `handleError()` helper to avoid repetition:

```go
// handleError maps errors to HTTP responses and writes them directly
func (h *LinkHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
    switch {
    case errors.Is(err, apperrors.LinkNotFound):
        h.logger.Warn("Link not found", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
        render.Status(r, http.StatusNotFound)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Code:    apperrors.CodeLinkNotFound,
                Message: "Link not found",
            },
        })

    case errors.Is(err, apperrors.InvalidURL):
        h.logger.Warn("Invalid URL", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Code:    apperrors.CodeInvalidURL,
                Message: "Invalid URL format",
            },
        })

    case errors.Is(err, sql.ErrNoRows):
        h.logger.Warn("Resource not found", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
        render.Status(r, http.StatusNotFound)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Code:    apperrors.CodeLinkNotFound,
                Message: "Resource not found",
            },
        })

    default:
        h.logger.Error("Internal server error", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
        render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, apperrors.ErrorResponse{
            Error: apperrors.ErrorDetail{
                Code:    apperrors.CodeInternalError,
                Message: "An unexpected error occurred",
            },
        })
    }
}
```

See `ERROR_HANDLING_GUIDE.md` for more details.

## Request Validation

### JSON Body Validation

```go
// Decode and validate
var reqBody dto.CreateLinkRequest
if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
    // Handle decode error
}

// Validate required fields
if reqBody.URL == "" {
    h.logger.Warn("URL is required",
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )
    render.Status(r, http.StatusBadRequest)
    render.JSON(w, r, apperrors.ErrorResponse{
        Error: apperrors.ErrorDetail{
            Message: "URL is required",
        },
    })
    return
}
```

### Path Parameter Validation

```go
// Extract and validate UUID
idStr := chi.URLParam(r, "id")
id, err := uuid.Parse(idStr)
if err != nil {
    h.logger.Warn("Invalid ID format",
        zap.Error(err),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )
    render.Status(r, http.StatusBadRequest)
    render.JSON(w, r, apperrors.ErrorResponse{
        Error: apperrors.ErrorDetail{
            Message: "Invalid ID format",
        },
    })
    return
}
```

### Query Parameter Validation

```go
// Extract query params
pageStr := r.URL.Query().Get("page")
if pageStr == "" {
    pageStr = "1"  // Default
}

page, err := strconv.Atoi(pageStr)
if err != nil || page < 1 {
    h.logger.Warn("Invalid page parameter",
        zap.Error(err),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )
    render.Status(r, http.StatusBadRequest)
    render.JSON(w, r, apperrors.ErrorResponse{
        Error: apperrors.ErrorDetail{
            Message: "Invalid page parameter",
        },
    })
    return
}
```

## Response Patterns

### Success Response

```go
render.Status(r, http.StatusOK)
render.JSON(w, r, &dto.SuccessResponse[db.Link]{
    Data:    link,
    Message: "Link retrieved successfully",
})
```

### Created Response

```go
render.Status(r, http.StatusCreated)
render.JSON(w, r, &dto.SuccessResponse[db.Link]{
    Data:    createdLink,
    Message: "Link created successfully",
})
```

### No Content Response

```go
render.Status(r, http.StatusNoContent)
// No body for DELETE
```

### Redirect Response

```go
http.Redirect(w, r, url, http.StatusFound)
```

## Testing Handlers

### Mock the Service Interface

```go
// pkg/handlers/link_test.go

type mockLinkService struct {
    CreateShortLinkFunc func(ctx context.Context, userID string, url string) (db.Link, error)
}

func (m *mockLinkService) CreateShortLink(ctx context.Context, userID string, url string) (db.Link, error) {
    if m.CreateShortLinkFunc != nil {
        return m.CreateShortLinkFunc(ctx, userID, url)
    }
    return db.Link{}, errors.New("not implemented")
}

func TestLinkHandler_CreateLink(t *testing.T) {
    mockService := &mockLinkService{
        CreateShortLinkFunc: func(ctx context.Context, userID string, url string) (db.Link, error) {
            return createTestLink(...), nil
        },
    }

    handler := &LinkHandler{
        LinkService: mockService,
        logger:      createTestLogger(),
    }

    // Create request
    reqBody := dto.CreateLinkRequest{URL: "https://example.com"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user_123")
    req = req.WithContext(ctx)

    // Execute
    w := httptest.NewRecorder()
    handler.CreateLink(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    // ... more assertions
}
```

## When to Create a New Handler

Create a new handler when:
- ✅ You have a new resource/domain entity
- ✅ The resource has different concerns than existing handlers
- ✅ You want to separate concerns

Don't create a new handler when:
- ❌ It's just another endpoint for the same resource (add method to existing handler)
- ❌ It's a simple operation (can add to existing handler)

## Handler Composition

Handlers can use multiple services:

```go
type DashboardHandler struct {
    LinkService      LinkServiceInterface
    AnalyticsService AnalyticsServiceInterface
    UserService      UserServiceInterface
    logger           logger.Logger
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())

    // Use multiple services
    links, _ := h.LinkService.ListAllLinks(r.Context(), userID)
    stats, _ := h.AnalyticsService.GetUserStats(r.Context(), userID)
    user, _ := h.UserService.GetUserProfile(r.Context(), userID)

    // Combine and render
    render.JSON(w, r, &dto.DashboardResponse{
        Links: links,
        Stats: stats,
        User:  user,
    })
}
```

## Summary

- **Handlers handle HTTP concerns only**
- **Use interfaces for services** (testability)
- **Handle errors directly** where they occur
- **Delegate business logic to services**
- **Use DTOs for request/response**
- **Keep business logic out**

## Further Reading

- `INTERFACES_GUIDE.md` - Why we use interfaces
- `ERROR_HANDLING_GUIDE.md` - Error handling patterns
- `ARCHITECTURE.md` - Overall architecture
- `SERVICE_PATTERNS.md` - Service layer patterns
- `pkg/handlers/link.go` - Example implementation

