# Go Interfaces: A Practical Guide for Our Codebase

## What Are Go Interfaces?

In Go, an **interface** is a type that defines a set of method signatures. Unlike many other languages, Go interfaces are **implicitly satisfied** - you don't need to explicitly declare that a type implements an interface. If a type has all the methods an interface requires, it automatically implements that interface.

### Basic Example

```go
// Define an interface
type Writer interface {
    Write([]byte) (int, error)
}

// Any type with a Write method automatically implements Writer
type File struct {}
func (f *File) Write(data []byte) (int, error) { /* ... */ }

type Buffer struct {}
func (b *Buffer) Write(data []byte) (int, error) { /* ... */ }

// Both File and Buffer can be used wherever Writer is expected
func saveData(w Writer, data []byte) {
    w.Write(data)  // Works with File, Buffer, or any Writer
}
```

## Why Use Interfaces?

### 1. **Dependency Injection & Testing**

Interfaces allow you to swap implementations, making code:
- **Testable**: Replace real dependencies with mocks
- **Flexible**: Change implementations without changing calling code
- **Modular**: Components depend on behavior, not concrete types

### 2. **Decoupling**

Interfaces create a contract between components. The caller doesn't need to know the implementation details - only what methods are available.

### 3. **Polymorphism**

Different types can be used interchangeably if they implement the same interface, enabling flexible, reusable code.

## How We Used Interfaces in Our Codebase

### The Problem We Solved

Initially, our code had **tight coupling**:

```go
// ❌ BEFORE: Tight coupling
type LinkService struct {
    queries *db.Queries  // Concrete type - hard to test
    logger  logger.Logger
}

type LinkHandler struct {
    LinkService *service.LinkService  // Concrete type - hard to test
    logger      logger.Logger
}
```

**Issues:**
- Can't test `LinkService` without a real database
- Can't test `LinkHandler` without a real `LinkService`
- Hard to swap implementations (e.g., different database, caching layer)

### The Solution: Interfaces

We introduced interfaces to create **loose coupling**:

```go
// ✅ AFTER: Loose coupling with interfaces

// In service/link.go
type LinkQueries interface {
    TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
    ListUserLinks(ctx context.Context, userID string) ([]db.Link, error)
    GetLinkByIdAndUser(ctx context.Context, arg db.GetLinkByIdAndUserParams) (db.Link, error)
    GetLinkForRedirect(ctx context.Context, shortcode string) (db.GetLinkForRedirectRow, error)
    DeleteLink(ctx context.Context, arg db.DeleteLinkParams) error
}

type LinkService struct {
    queries LinkQueries  // Interface - easy to mock!
    logger  logger.Logger
}

// In handlers/link.go
type LinkServiceInterface interface {
    CreateShortLink(ctx context.Context, userID string, originalURL string) (db.Link, error)
    ListAllLinks(ctx context.Context, userID string) ([]db.Link, error)
    GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error)
    GetOriginalURL(ctx context.Context, code string) (db.GetLinkForRedirectRow, error)
    DeleteLink(ctx context.Context, id uuid.UUID, userID string) error
}

type LinkHandler struct {
    LinkService LinkServiceInterface  // Interface - easy to mock!
    logger      logger.Logger
}
```

**Benefits:**
- ✅ Can test `LinkService` with mock queries (no database needed)
- ✅ Can test `LinkHandler` with mock service (no service logic needed)
- ✅ Fast, isolated unit tests
- ✅ Easy to add new implementations (caching, different databases, etc.)

## Scaling to Multiple Services and Handlers

### Current State: One Service, One Handler

Right now we have:
- `LinkService` - handles link operations
- `LinkHandler` - handles HTTP requests for links

### Future State: Multiple Services and Handlers

As the codebase grows, you'll likely add:
- `UserService` - user management
- `AnalyticsService` - click tracking, statistics
- `AuthService` - authentication logic
- `UserHandler`, `AnalyticsHandler`, etc.

### How Interfaces Help Scale

#### 1. **Consistent Testing Pattern**

Every service and handler follows the same pattern:

```go
// Future: AnalyticsService
type AnalyticsQueries interface {
    GetClickStats(ctx context.Context, linkID uuid.UUID) ([]ClickStat, error)
    RecordClick(ctx context.Context, click Click) error
}

type AnalyticsService struct {
    queries AnalyticsQueries  // Same pattern as LinkService
    logger  logger.Logger
}

// Future: AnalyticsHandler
type AnalyticsServiceInterface interface {
    GetClickStats(ctx context.Context, linkID uuid.UUID) ([]ClickStat, error)
    RecordClick(ctx context.Context, click Click) error
}

type AnalyticsHandler struct {
    AnalyticsService AnalyticsServiceInterface  // Same pattern as LinkHandler
    logger           logger.Logger
}
```

**Result:** Every new service/handler is testable from day one using the same mocking approach.

#### 2. **Easy to Add Cross-Cutting Concerns**

Want to add caching? Just create a wrapper:

```go
// CachedLinkService wraps LinkService and adds caching
type CachedLinkService struct {
    service LinkServiceInterface
    cache   CacheInterface
}

func (c *CachedLinkService) GetLinkByID(ctx context.Context, id uuid.UUID, userID string) (db.Link, error) {
    // Check cache first
    if link, found := c.cache.Get(id.String()); found {
        return link, nil
    }
    
    // Fall back to service
    link, err := c.service.GetLinkByID(ctx, id, userID)
    if err == nil {
        c.cache.Set(id.String(), link)
    }
    return link, err
}
```

The handler doesn't need to change - it still uses `LinkServiceInterface`, but now gets caching for free!

#### 3. **Multiple Implementations**

You can have different implementations for different environments:

```go
// Production: Real database
prodService := service.NewLinkService(realQueries, logger)

// Testing: Mock database
testService := service.NewLinkService(mockQueries, logger)

// Development: In-memory database
devService := service.NewLinkService(memoryQueries, logger)
```

All three implement `LinkServiceInterface`, so handlers work with any of them.

#### 4. **Service Composition**

Handlers can depend on multiple services without tight coupling:

```go
// Future: A handler that needs multiple services
type DashboardHandler struct {
    LinkService      LinkServiceInterface
    AnalyticsService AnalyticsServiceInterface
    UserService      UserServiceInterface
    logger           logger.Logger
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())
    
    // Use multiple services - all via interfaces
    links, _ := h.LinkService.ListAllLinks(r.Context(), userID)
    stats, _ := h.AnalyticsService.GetUserStats(r.Context(), userID)
    user, _ := h.UserService.GetUser(r.Context(), userID)
    
    // Combine and return
    // ...
}
```

Each service is independently testable and swappable.

## Real Example: Our Test Mocks

Here's how we use interfaces in tests:

```go
// In service/link_test.go
type mockQueries struct {
    TryCreateLinkFunc func(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
    // ... other methods
}

func (m *mockQueries) TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error) {
    if m.TryCreateLinkFunc != nil {
        return m.TryCreateLinkFunc(ctx, arg)
    }
    return db.Link{}, errors.New("not implemented")
}

// mockQueries implements LinkQueries interface automatically!
// No need to declare "mockQueries implements LinkQueries"

func TestLinkService_CreateShortLink(t *testing.T) {
    mockQueries := &mockQueries{
        TryCreateLinkFunc: func(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error) {
            // Custom test behavior
            return createTestLink(...), nil
        },
    }
    
    // Can use mockQueries anywhere LinkQueries is expected
    service := &LinkService{
        queries: mockQueries,  // ✅ Works because mockQueries implements LinkQueries
        logger:  createTestLogger(),
    }
    
    // Test without database!
}
```

## Interface Design Principles

### 1. **Keep Interfaces Small**

Prefer small, focused interfaces:

```go
// ✅ GOOD: Small, focused interface
type Reader interface {
    Read([]byte) (int, error)
}

// ❌ BAD: Too many responsibilities
type FileOperations interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Delete() error
    Rename(string) error
    // ... 20 more methods
}
```

**Why?** Small interfaces are easier to implement and more flexible.

### 2. **Accept Interfaces, Return Structs**

```go
// ✅ GOOD: Function accepts interface
func ProcessData(r Reader) error {
    // Works with any Reader implementation
}

// ❌ BAD: Function accepts concrete type
func ProcessData(f *File) error {
    // Only works with File
}
```

### 3. **Define Interfaces Where They're Used**

We defined `LinkServiceInterface` in the `handlers` package because that's where it's used:

```go
// handlers/link.go
type LinkServiceInterface interface {
    // Methods that handlers actually need
}

type LinkHandler struct {
    LinkService LinkServiceInterface  // Used here
}
```

**Why?** The handler defines what it needs from the service, not the other way around. This follows the **Interface Segregation Principle**.

## Common Patterns in Our Codebase

### Pattern 1: Service Layer Interface

```go
// Service defines what it needs from data layer
type LinkQueries interface {
    // Database operations
}

type LinkService struct {
    queries LinkQueries  // Depends on interface
}
```

### Pattern 2: Handler Layer Interface

```go
// Handler defines what it needs from service layer
type LinkServiceInterface interface {
    // Business logic operations
}

type LinkHandler struct {
    LinkService LinkServiceInterface  // Depends on interface
}
```

### Pattern 3: Mock Implementation

```go
// Test file defines mock that implements interface
type mockLinkService struct {
    CreateShortLinkFunc func(...) (db.Link, error)
    // ...
}

func (m *mockLinkService) CreateShortLink(...) (db.Link, error) {
    // Mock implementation
}
```

## When NOT to Use Interfaces

Interfaces aren't always the answer:

### ❌ Don't Use Interfaces For:
- **Simple data structures** (structs with no methods)
- **Over-abstracting** (if you only have one implementation, interfaces add complexity)
- **Premature optimization** (add interfaces when you need them, not "just in case")

### ✅ Do Use Interfaces For:
- **Testing** (mocking dependencies)
- **Multiple implementations** (different databases, caching layers)
- **Cross-cutting concerns** (logging, metrics, middleware)
- **Plugin architectures** (swappable components)

## Summary

We introduced interfaces to:

1. **Enable Testing**: Mock dependencies without real databases/services
2. **Reduce Coupling**: Components depend on behavior, not implementations
3. **Scale Gracefully**: New services/handlers follow the same testable pattern
4. **Enable Flexibility**: Swap implementations (caching, different databases) without changing callers

The refactoring we did might seem like "extra work" for one service and one handler, but it establishes a **scalable pattern** that will pay dividends as your codebase grows. Every new service and handler you add will be testable from day one, following the same clean architecture.

## Further Reading

- [Go Interfaces Tutorial](https://go.dev/tour/methods/9)
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Code Review Comments: Interfaces](https://github.com/golang/go/wiki/CodeReviewComments#interfaces)

