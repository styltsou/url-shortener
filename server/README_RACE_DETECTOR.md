# Go Race Detector

## What is the Race Detector?

The **race detector** (`-race` flag) is a tool built into Go that detects **data races** in concurrent programs.

## What is a Data Race?

A data race occurs when:
- Two or more goroutines access the same memory location **concurrently**
- At least one access is a **write**
- The accesses are **not synchronized** (no mutex, channel, etc.)

### Example of a Data Race:

```go
var counter int

// Goroutine 1
go func() {
    counter++  // Write
}()

// Goroutine 2
go func() {
    counter++  // Write
}()

// Both goroutines write to 'counter' without synchronization
// This is a data race - undefined behavior!
```

## How the Race Detector Works

When you run tests with `-race`:
- Go instruments your code to track memory accesses
- It monitors for concurrent unsynchronized access to the same variable
- If a race is detected, it reports it and the test fails

## Usage

```bash
# Run tests with race detector
go test -race ./...

# Or in Makefile
make test
```

## Important Notes

### Performance Impact:
- **Slower**: Tests run 2-10x slower with race detector
- **More memory**: Uses 5-10x more memory
- **Only for testing**: Don't use `-race` in production builds

### When to Use:
- ✅ **Always in tests** - Catches concurrency bugs early
- ✅ **During development** - Especially when working with goroutines
- ✅ **In CI/CD** - Automated race detection

### When NOT to Use:
- ❌ **Production builds** - Too slow and uses too much memory
- ❌ **Performance benchmarks** - Skews results

## Example Output

If a race is detected, you'll see:

```
WARNING: DATA RACE
Read at 0x00c00001a0a8 by goroutine 7:
  main.increment()
      /path/to/file.go:10 +0x3a

Previous write at 0x00c00001a0a8 by goroutine 6:
  main.increment()
      /path/to/file.go:10 +0x56

Goroutine 7 (running) created at:
  main.main()
      /path/to/file.go:15 +0x5a

Goroutine 6 (finished) created at:
  main.main()
      /path/to/file.go:14 +0x5a
```

## Fixing Data Races

Use synchronization primitives:

```go
// ❌ Race condition
var counter int
go func() { counter++ }()

// ✅ Fixed with mutex
var (
    counter int
    mu      sync.Mutex
)
go func() {
    mu.Lock()
    counter++
    mu.Unlock()
}()

// ✅ Fixed with atomic
var counter int64
go func() {
    atomic.AddInt64(&counter, 1)
}()
```

## In Your Project

Your Makefile runs tests with `-race` by default:
```makefile
test: ## Run tests with race detector
	@go test -v -race ./...
```

This is **good practice** - it will catch any concurrency bugs in your handlers, services, or middleware that use goroutines.

