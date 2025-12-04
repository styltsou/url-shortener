# Go Pointers Guide - Explained with Your Codebase

This guide explains pointers in Go using examples from your actual codebase.

## Table of Contents
1. [The Core Concept](#the-core-concept)
2. [Why `logger` is a Pointer but `queries` is Not](#why-logger-is-a-pointer-but-queries-is-not)
3. [When to Use `&` When Passing Arguments](#when-to-use--when-passing-arguments)
4. [When to Return Pointers vs Values](#when-to-return-pointers-vs-values)
5. [Pointers in Struct Fields](#pointers-in-struct-fields)
6. [Common Patterns in Your Codebase](#common-patterns-in-your-codebase)
7. [Quick Reference](#quick-reference)

---

## The Core Concept

**A pointer holds the memory address of a value.** Think of it like a house address vs the house itself.

```go
var x int = 42      // x is the value (the house)
var p *int = &x     // p is a pointer to x (the address of the house)
```

**Key Rules:**
- `&` gets the address of a value (creates a pointer)
- `*` dereferences a pointer (gets the value it points to)
- `nil` is the zero value for pointers (no address)

---

## Why Both `queries` and `logger` Are Interfaces (Not Pointers)

This is the most confusing part! Let's look at your actual code:

### Your Code:

```go
// In service/link.go
type LinkService struct {
	queries LinkQueries  // NOT a pointer - it's an INTERFACE
	logger  logger.Logger  // NOT a pointer - it's also an INTERFACE
}

func NewLinkService(queries LinkQueries, logger logger.Logger) *LinkService {
	return &LinkService{
		queries: queries,  // No & needed!
		logger:  logger,   // No & needed!
	}
}
```

### Why This Happens:

**1. `queries` is an INTERFACE:**

```go
type LinkQueries interface {
	TryCreateLink(ctx context.Context, arg db.TryCreateLinkParams) (db.Link, error)
	// ... other methods
}
```

**2. `logger` is also an INTERFACE:**

```go
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	// ... other methods
}
```

**Interfaces in Go are already "pointer-like"** - they're a two-word structure:
- Word 1: Pointer to the actual value
- Word 2: Pointer to the type information

When you do:
```go
queries := db.New(pool)  // Returns *db.Queries (a pointer)
log, _ := logger.New("dev")  // Returns *logger.ZapLogger (a pointer)
linkSvc := service.NewLinkService(queries, log)
```

Even though both `db.New()` and `logger.New()` return pointers to concrete types, when you assign them to interfaces, Go automatically handles the conversion. The interface **already contains a pointer internally**, so you don't need to store a pointer to the interface itself.

### Visual Comparison:

```go
// Interface (queries) - already "pointer-like" internally
LinkQueries {
    data: *db.Queries  ← pointer inside interface
    type: *QueriesType
}

// Interface (logger) - also "pointer-like" internally
logger.Logger {
    data: *logger.ZapLogger  ← pointer inside interface
    type: *ZapLoggerType
}
```

**Key Point**: Both are interfaces, so neither needs a pointer. The interface itself already contains a pointer to the concrete implementation.

---

## When to Use `&` When Passing Arguments

### Rule: Use `&` when the function expects a pointer

Look at your `CreateLink` handler:

```go
func (h *LinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var reqBody dto.CreateLinkRequest  // Value, not pointer
	
	// json.Decoder.Decode expects a pointer!
	json.NewDecoder(r.Body).Decode(&reqBody)  // ← & here!
	
	// render.JSON can take either, but pointer is more efficient
	render.JSON(w, r, &dto.SuccessResponse[db.Link]{  // ← & here!
		Data:    createdLink,
		Message: "Short Link created successfully",
	})
}
```

**Why `&reqBody`?**
- `Decode` needs to **modify** the value (fill it with JSON data)
- Functions that modify values must take pointers
- Without `&`, Go would try to pass a copy, and modifications would be lost

**Why `&dto.SuccessReponse`?**
- Not strictly necessary (render.JSON accepts values too)
- But more efficient - avoids copying the struct
- Common Go idiom for structs

### When NOT to use `&`:

```go
// Strings, ints, slices - usually pass by value
func validateURL(rawURL string) error {  // No pointer needed
	// ...
}

// Interfaces - already "pointer-like"
func NewLinkService(queries LinkQueries, logger logger.Logger) *LinkService {
	// Both are interfaces - no & needed when assigning
	return &LinkService{
		queries: queries,  // No & here!
		logger:  logger,   // No & here! (interface, not pointer)
	}
}
```

---

## When to Return Pointers vs Values

### Return Pointers When:

**1. The struct is large or you want to share the same instance:**

```go
// Your code - returns pointer
func NewLinkService(queries LinkQueries, logger logger.Logger) *LinkService {
	return &LinkService{  // ← Returns pointer
		queries: queries,
		logger:  logger,
	}
}
```

**Why?** Multiple parts of your code can share the same service instance. If you returned a value, each assignment would create a copy.

**2. The value might be nil (optional):**

```go
// In your DTO (future)
type CreateLinkRequest struct {
	URL       string
	Shortcode *string  // ← Pointer = optional field (can be nil)
	ExpiresAt *string  // ← Pointer = optional field (can be nil)
}
```

**Why?** `nil` means "not provided". With a value, you'd need a sentinel like empty string `""`, which is ambiguous.

### Return Values When:

**1. Small, immutable types:**

```go
func generateRandomCode(n int) (string, error) {  // Returns value, not *string
	// ...
	return string(b), nil
}
```

**2. The caller doesn't need to modify it:**

```go
func (s *LinkService) GetLinkByID(...) (db.Link, error) {  // Returns value
	return link, nil
}
```

**Why?** `db.Link` is returned from database - caller gets a copy, which is fine. They're not modifying the original.

---

## Pointers in Struct Fields

### When to Use Pointers in Struct Fields:

**1. Optional fields (can be nil):**

```go
type Link struct {
	ID          uuid.UUID
	Shortcode   string
	Clicks      *int32           // ← Pointer = can be nil (NULL in DB)
	ExpiresAt   pgtype.Timestamp // ← Not a pointer, but pgtype handles NULL
}
```

**Why `*int32` for Clicks?**
- Database can return `NULL`
- Go's `int32` can't represent "not set"
- `*int32` can be `nil` = NULL, or point to a value = actual number

**2. Large structs you don't want to copy:**

```go
type LinkService struct {
	queries LinkQueries      // Interface (already efficient)
	logger  logger.Logger   // ← Interface (no pointer needed)
}
```

**3. Fields that need to be shared/mutated:**

```go
type Server struct {
	Logger logger.Logger  // ← Interface (no pointer needed)
	Pool   *pgxpool.Pool   // ← Shared connection pool
}
```

### When NOT to Use Pointers:

**1. Small, simple types:**

```go
type CreateLinkRequest struct {
	URL string  // ← No pointer needed (string is already a reference type internally)
}
```

**2. Slices and maps (already reference types):**

```go
type Response struct {
	Links []db.Link  // ← No pointer needed
	Meta  map[string]string  // ← No pointer needed
}
```

---

## Common Patterns in Your Codebase

### Pattern 1: Constructor Functions Return Pointers

```go
// ✅ GOOD - Returns pointer
func NewLinkService(...) *LinkService {
	return &LinkService{...}
}

// Usage:
linkSvc := service.NewLinkService(queries, logger)
// linkSvc is *LinkService (pointer)
```

**Why?** Services are typically shared, singleton-like objects.

### Pattern 2: Methods on Pointer Receivers

```go
// ✅ GOOD - Pointer receiver
func (s *LinkService) CreateShortLink(...) (db.Link, error) {
	s.logger.Debug(...)  // Can access fields
	// ...
}

// ❌ BAD - Value receiver (would copy entire service)
func (s LinkService) CreateShortLink(...) {
	// This copies the entire LinkService struct!
}
```

**Rule:** If your struct has any pointer fields or is larger than a few words, use pointer receivers.

### Pattern 3: Interface Assignment

```go
// db.New returns *db.Queries (pointer)
queries := db.New(pool)  // queries is *db.Queries

// But LinkQueries interface accepts it directly
linkSvc := service.NewLinkService(queries, logger)
// No & needed! Interface handles it.
```

**Why?** Go automatically converts `*db.Queries` to `LinkQueries` interface. The interface internally stores a pointer.

### Pattern 4: JSON Decoding

```go
var reqBody dto.CreateLinkRequest  // Value
json.NewDecoder(r.Body).Decode(&reqBody)  // Must use &
```

**Why?** `Decode` needs to modify the struct, so it requires a pointer.

---

## Quick Reference

### The Golden Rules

#### 1. **Interfaces are already "pointer-like"**
```go
type Service struct {
    queries LinkQueries  // ✅ NO pointer - interface handles it internally
    logger  *Logger      // ✅ YES pointer - concrete struct
}
```

**Why?** Interfaces store a pointer internally. You don't need `*LinkQueries`.

#### 2. **Use `&` when function needs to MODIFY**
```go
var req CreateLinkRequest
json.Decode(&req)  // ✅ Needs & because it modifies req
```

#### 3. **Use `&` when creating optional fields**
```go
code := "my-link"
req := CreateLinkRequest{
    Shortcode: &code,  // ✅ & makes it optional (can be nil)
}
```

#### 4. **Return pointers from constructors**
```go
func NewService() *Service {  // ✅ Returns pointer
    return &Service{}
}
```

#### 5. **Use pointer receivers for methods**
```go
func (s *Service) Method() {  // ✅ Pointer receiver
    // Can modify s
}
```

### Visual Guide

```
┌─────────────────────────────────────────────────────────┐
│ INTERFACE (LinkQueries)                                 │
│ ┌─────────────┐  ┌──────────────┐                       │
│ │ data: *ptr  │  │ type: *type  │  ← Already has ptr!   │
│ └─────────────┘  └──────────────┘                       │
└─────────────────────────────────────────────────────────┘
         ↓
    queries LinkQueries  // NO * needed!


┌─────────────────────────────────────────────────────────┐
│ CONCRETE STRUCT (Logger)                                │
│ ┌─────────────┐  ┌──────────────┐                       │
│ │ logger: ... │  │ isDev: bool  │  ← Real data           │
│ └─────────────┘  └──────────────┘                       │
└─────────────────────────────────────────────────────────┘
         ↓
    logger *Logger  // YES * needed!
```

### Decision Tree

```
Need to store in struct field?
│
├─ Is it an INTERFACE?
│  └─ NO pointer needed ✅
│     queries LinkQueries
│
└─ Is it a CONCRETE TYPE?
   │
   ├─ Is it LARGE or needs to be SHARED?
   │  └─ YES → Use pointer ✅
   │     logger *Logger
   │
   └─ Is it SMALL and simple?
      └─ NO pointer needed ✅
         id uuid.UUID
```

### Quick Reference Table

| Situation | Use Pointer? | Example |
|-----------|-------------|---------|
| Interface type | ❌ No | `logger logger.Logger` |
| Interface field | ❌ No | `queries LinkQueries` |
| Optional field | ✅ Yes | `Clicks *int32` |
| Small value field | ❌ No | `ID uuid.UUID` |
| Function that modifies | ✅ Yes | `Decode(&reqBody)` |
| Function that reads | ❌ Usually no | `validateURL(url string)` |
| Constructor return | ✅ Usually yes | `func New() *Service` |
| Method receiver | ✅ Usually yes | `func (s *Service) Method()` |
| Return small value | ❌ No | `func Get() (string, error)` |
| Return large struct | ✅ Usually yes | `func Get() (*Config, error)` |

### Common Patterns

| Pattern | Code | Why |
|---------|------|-----|
| Interface field | `queries LinkQueries` | Interface already has pointer |
| Large struct field | `logger *Logger` | Avoid copying |
| Optional field | `Shortcode *string` | Can be `nil` |
| Function modifies | `Decode(&req)` | Needs pointer to modify |
| Constructor | `func New() *Service` | Returns pointer |
| Method receiver | `func (s *Service) Method()` | Pointer receiver |

### Your Codebase Examples

#### ✅ CORRECT - Your actual code
```go
type LinkService struct {
    queries LinkQueries      // Interface - no pointer
    logger  logger.Logger   // Interface - no pointer
}

func NewLinkService(queries LinkQueries, logger logger.Logger) *LinkService {
    return &LinkService{
        queries: queries,  // No & needed
        logger:  logger,   // Already pointer, no & needed
    }
}
```

#### ✅ CORRECT - JSON decoding
```go
var reqBody dto.CreateLinkRequest
json.NewDecoder(r.Body).Decode(&reqBody)  // & needed!
```

#### ✅ CORRECT - Optional fields
```go
code := "my-link"
req := CreateLinkRequest{
    URL:       "https://example.com",
    Shortcode: &code,  // & makes it optional
}
```

### When in Doubt

1. **Is it an interface?** → No pointer needed
2. **Does function modify it?** → Use `&` when calling
3. **Is it large or shared?** → Use pointer
4. **Can it be nil/optional?** → Use pointer
5. **Is it small and simple?** → No pointer needed

---

## Debugging Tips

### Check if something is a pointer:

```go
var x logger.Logger  // Interface type
fmt.Printf("%T\n", x)  // logger.Logger (interface)
// Note: Interfaces can be nil, but you can't dereference them like pointers
```

### Check if interface is nil:

```go
if logger == nil {
	// Handle nil case
}
```

### Common Mistake:

```go
// ❌ WRONG
var logger logger.Logger  // nil interface
logger.Info("test")  // PANIC! nil interface call

// ✅ CORRECT
log, _ := logger.New("dev")  // Returns *logger.ZapLogger, assigned to logger.Logger interface
log.Info("test")  // Works!
```

---

## Summary

1. **Interfaces are already "pointer-like"** - don't need `*` when storing in structs
2. **Concrete structs** - use pointers if large or need to share
3. **Use `&`** when function needs to modify the value
4. **Return pointers** for constructors and large structs
5. **Use pointer fields** for optional values (can be nil) or large structs
6. **Use value fields** for small, simple types

The key insight: **Interfaces in Go already contain pointers internally**, so you don't need to make them pointers. Concrete structs are different - they're copied by value unless you use pointers.
