# Service Layer Patterns

## Overview

Services contain **business logic** - the "what" and "why" of your application, not the "how" of HTTP or database access. This guide explains how to create and structure services.

## Service Responsibilities

A service should:
- ✅ Contain business logic and domain rules
- ✅ Validate business rules (beyond input validation)
- ✅ Coordinate between multiple repositories if needed
- ✅ Return errors (sentinel errors for known cases)
- ✅ Be independent of HTTP concerns

A service should NOT:
- ❌ Know about HTTP requests/responses
- ❌ Know about DTOs (except database models)
- ❌ Write to HTTP response writers
- ❌ Handle HTTP status codes
- ❌ Know about routing

## Service Structure

### Basic Pattern

```go
// pkg/service/link.go

// LinkQueries defines database operations needed by LinkService
type LinkQueries interface {
    TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
    ListUserLinks(ctx context.Context, userID string) ([]db.Link, error)
    // ... other methods
}

type LinkService struct {
    queries LinkQueries  // Interface, not concrete type!
    logger  logger.Logger
}

func NewLinkService(queries *db.Queries, logger logger.Logger) *LinkService {
    return &LinkService{
        queries: queries,  // db.Queries implements LinkQueries
        logger:  logger,
    }
}
```

### Why Interfaces?

- **Testability**: Mock the interface in tests
- **Flexibility**: Swap implementations (caching, different databases)
- **Dependency Inversion**: Service depends on abstraction, not concrete type

See `INTERFACES_GUIDE.md` for details.

## Creating a New Service

### Step 1: Define the Interface

```go
// pkg/service/user.go

type UserQueries interface {
    GetUser(ctx context.Context, userID string) (db.User, error)
    CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

type UserService struct {
    queries UserQueries
    logger  logger.Logger
}

func NewUserService(queries *db.Queries, logger logger.Logger) *UserService {
    return &UserService{
        queries: queries,
        logger:  logger,
    }
}
```

### Step 2: Implement Business Logic

```go
func (s *UserService) GetUserProfile(ctx context.Context, userID string) (db.User, error) {
    // Business logic: validate userID format
    if userID == "" {
        return db.User{}, fmt.Errorf("%w: userID is required", apperrors.ErrInvalidInput)
    }

    // Call repository
    user, err := s.queries.GetUser(ctx, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.User{}, fmt.Errorf("%w: user %s", apperrors.ErrUserNotFound, userID)
        }
        return db.User{}, fmt.Errorf("failed to get user: %w", err)
    }

    // Business logic: check if user is active
    if !user.IsActive {
        return db.User{}, fmt.Errorf("%w: user is inactive", apperrors.ErrUserInactive)
    }

    return user, nil
}
```

### Step 3: Error Handling

Services should return **sentinel errors** for known cases:

```go
// In pkg/errors/errors.go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserInactive = errors.New("user is inactive")
    ErrInvalidInput = errors.New("invalid input")
)

// In service
if userID == "" {
    return db.User{}, fmt.Errorf("%w: userID is required", apperrors.ErrInvalidInput)
}
```

Handlers will check these with `errors.Is()` and map them to HTTP responses directly.

## Common Patterns

### Pattern 1: Create with Validation

```go
func (s *LinkService) CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error) {
    // 1. Validate input
    if err := validateURL(originalURL); err != nil {
        return db.Link{}, err  // Returns sentinel error
    }

    // 2. Business logic: generate code
    code, err := generateRandomCode(9)
    if err != nil {
        return db.Link{}, fmt.Errorf("failed to generate code: %w", err)
    }

    // 3. Try to create (handle collisions)
    for range maxAttempts {
        link, err := s.queries.TryCreateLink(ctx, db.TryCreateLinkParams{
            Shortcode:   code,
            OriginalUrl: originalURL,
            UserID:      userID,
        })

        if err == nil {
            return link, nil
        }

        // Handle collision
        if errors.Is(err, sql.ErrNoRows) {
            // Collision, try again
            code, err = generateRandomCode(9)
            if err != nil {
                return db.Link{}, fmt.Errorf("failed to generate code: %w", err)
            }
            continue
        }

        // Other error
        return db.Link{}, fmt.Errorf("failed to create link: %w", err)
    }

    return db.Link{}, fmt.Errorf("failed to create link after %d attempts", maxAttempts)
}
```

### Pattern 2: Get with Not Found Handling

```go
func (s *LinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error) {
    link, err := s.queries.GetLinkByIdAndUser(ctx, db.GetLinkByIdAndUserParams{
        ID:     id,
        UserID: userID,
    })
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.Link{}, fmt.Errorf("%w: %v", apperrors.ErrLinkNotFound, err)
        }
        return db.Link{}, fmt.Errorf("failed to get link: %w", err)
    }
    
    return link, nil
}
```

### Pattern 3: List with Empty Result

```go
func (s *LinkService) ListAllLinks(ctx context.Context, userID string) ([]db.Link, error) {
    links, err := s.queries.ListUserLinks(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to list links: %w", err)
    }

    // Empty slice is valid - user just has no links
    // Don't return error for empty results
    return links, nil
}
```

### Pattern 4: Update with Validation

```go
func (s *LinkService) UpdateLink(ctx context.Context, id uuid.UUID, userID string, updates UpdateLinkParams) (db.Link, error) {
    // 1. Get existing link
    link, err := s.GetLinkByID(ctx, id, userID)
    if err != nil {
        return db.Link{}, err
    }

    // 2. Business logic: validate updates
    if updates.Shortcode != "" {
        if err := validateShortcode(updates.Shortcode); err != nil {
            return db.Link{}, err
        }
        // Check if code is available
        if exists, err := s.queries.ShortcodeExists(ctx, updates.Shortcode); err != nil {
            return db.Link{}, fmt.Errorf("failed to check shortcode: %w", err)
        } else if exists {
            return db.Link{}, fmt.Errorf("%w: shortcode already taken", apperrors.ErrCodeTaken)
        }
    }

    // 3. Update
    updated, err := s.queries.UpdateLink(ctx, db.UpdateLinkParams{
        ID:        id,
        UserID:    userID,
        Shortcode: updates.Shortcode,
        ExpiresAt: updates.ExpiresAt,
    })
    if err != nil {
        return db.Link{}, fmt.Errorf("failed to update link: %w", err)
    }

    return updated, nil
}
```

### Pattern 5: Delete with Authorization Check

```go
func (s *LinkService) DeleteLink(ctx context.Context, id uuid.UUID, userID string) error {
    // Verify link exists and belongs to user
    _, err := s.GetLinkByID(ctx, id, userID)
    if err != nil {
        return err  // Already wrapped with ErrLinkNotFound if not found
    }

    // Delete
    err = s.queries.DeleteLink(ctx, db.DeleteLinkParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return fmt.Errorf("%w: %v", apperrors.ErrLinkNotFound, err)
        }
        return fmt.Errorf("failed to delete link: %w", err)
    }

    return nil
}
```

## Error Handling Best Practices

### 1. Use Sentinel Errors

```go
// ✅ GOOD: Return sentinel error
if userID == "" {
    return db.User{}, fmt.Errorf("%w: userID is required", apperrors.ErrInvalidInput)
}

// ❌ BAD: Return generic error
if userID == "" {
    return db.User{}, errors.New("userID is required")
}
```

### 2. Wrap Errors with Context

```go
// ✅ GOOD: Wrap with context
if err != nil {
    return db.Link{}, fmt.Errorf("failed to create link: %w", err)
}

// ❌ BAD: Lose original error
if err != nil {
    return db.Link{}, errors.New("failed to create link")
}
```

### 3. Map Database Errors

```go
// ✅ GOOD: Map to domain error
if errors.Is(err, sql.ErrNoRows) {
    return db.Link{}, fmt.Errorf("%w: %v", apperrors.ErrLinkNotFound, err)
}

// ❌ BAD: Return database error directly
if err != nil {
    return db.Link{}, err  // Exposes database implementation
}
```

## Testing Services

### Mock the Interface

```go
// pkg/service/link_test.go

type mockQueries struct {
    TryCreateLinkFunc func(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
}

func (m *mockQueries) TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error) {
    if m.TryCreateLinkFunc != nil {
        return m.TryCreateLinkFunc(ctx, arg)
    }
    return db.Link{}, errors.New("not implemented")
}

func TestLinkService_CreateShortLink(t *testing.T) {
    mockQueries := &mockQueries{
        TryCreateLinkFunc: func(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error) {
            // Test behavior
            return createTestLink(...), nil
        },
    }

    service := &LinkService{
        queries: mockQueries,  // Use mock
        logger:  createTestLogger(),
    }

    // Test service
    link, err := service.CreateShortLink(ctx, "user_123", "https://example.com")
    // Assertions...
}
```

## When to Create a New Service

Create a new service when:
- ✅ You have a new resource/domain entity
- ✅ Business logic is complex enough to warrant separation
- ✅ You need to coordinate multiple repositories

Don't create a new service when:
- ❌ It's just a simple CRUD operation (can add to existing service)
- ❌ It's just a database query (use repository directly)
- ❌ It's HTTP-specific (belongs in handler)

## Service Composition

Services can call other services:

```go
type AnalyticsService struct {
    linkService *LinkService  // Can use other services
    queries     AnalyticsQueries
    logger      logger.Logger
}

func (s *AnalyticsService) GetLinkStats(ctx context.Context, linkID uuid.UUID) (Stats, error) {
    // Use LinkService to verify link exists
    _, err := s.linkService.GetLinkByID(ctx, linkID, userID)
    if err != nil {
        return Stats{}, err
    }

    // Get analytics
    return s.queries.GetStats(ctx, linkID)
}
```

## Summary

- **Services contain business logic**
- **Use interfaces for dependencies** (testability)
- **Return sentinel errors** (for known cases)
- **Wrap errors with context**
- **Map database errors to domain errors**
- **Keep HTTP concerns out**

## Further Reading

- `INTERFACES_GUIDE.md` - Why we use interfaces
- `ERROR_HANDLING_GUIDE.md` - Error handling patterns
- `ARCHITECTURE.md` - Overall architecture
- `pkg/service/link.go` - Example implementation

