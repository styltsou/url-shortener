# Formatting & Linting Guide

## Why golangci-lint? (It's Actually ONE Tool)

**golangci-lint is a single tool** that runs many linters for you. Think of it like:
- **Without it**: Install and run `errcheck`, `staticcheck`, `govet`, `unused`, etc. separately
- **With it**: Install one tool, it runs all of them

The YAML file is just configuration - you can use defaults or simplify it (like we did).

### Alternatives (Simpler Options):

1. **Just `go vet`** (built-in, zero config):
   ```bash
   go vet ./...
   ```
   - Basic checks only
   - No installation needed

2. **Just `staticcheck`** (one extra tool):
   ```bash
   go install honnef.co/go/tools/cmd/staticcheck@latest
   staticcheck ./...
   ```
   - Catches more bugs than `go vet`
   - Still simple

3. **golangci-lint** (what we have):
   - Runs many linters in one command
   - Configurable (but you can use defaults)
   - Industry standard

**Recommendation**: Start with just `go vet` if you want simplicity. Add `staticcheck` or `golangci-lint` when you want more checks.

## Automatic Line Wrapping? **No, Go Doesn't Support This**

### The Reality:
- ❌ `gofmt` does NOT wrap long lines automatically
- ❌ There's no standard Go tool that wraps lines
- ❌ Go's philosophy: "You decide what's readable"

### Why No Auto-Wrapping?
Go's designers believe line length is a **readability judgment**, not something to automate. They want you to decide where to break lines based on context.

### Your Options:

1. **Editor Soft Wrap** (visual only):
   - VS Code: `"editor.wordWrap": "on"`
   - Just displays wrapped, doesn't change the file
   - Useful for reading, but not for formatting

2. **Manual Breaking** (what most Go devs do):
   ```go
   // Break function signatures
   func (s *LinkService) CreateShortLink(
       ctx context.Context,
       userID string,
       originalURL string,
   ) (db.Link, error) {
   
   // Break long strings
   panic(
       "request body not found in context: " +
       "RequestValidator middleware must be applied",
   )
   ```

3. **Editor Ruler** (visual guide):
   - Shows a line at 100/120 chars
   - Helps you decide when to break manually
   - VS Code: `"editor.rulers": [100, 120]`

4. **golangci-lint `lll` linter** (warns, doesn't fix):
   - Warns on long lines
   - You still break them manually
   - Already configured in `.golangci.yml` (line-length: 120)

### The Go Way:
Most Go developers:
1. Write code naturally
2. Break lines manually when they feel it's too long
3. Don't stress about exact character counts
4. Prioritize clarity over strict limits

## Simplified Setup

### Minimal (Just Editor):
- Your editor's Go plugin handles `gofmt`/`goimports`
- That's it!

### Recommended (Editor + Basic Checks):
```bash
# Just run go vet (built-in)
go vet ./...

# Or install staticcheck for more checks
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

### Current Setup (Optional):
- `.golangci.yml` - Simplified config (can delete if you prefer)
- Run: `golangci-lint run` (only if you installed it)

**Bottom line**: Your editor is enough. The linter is optional for catching bugs.
