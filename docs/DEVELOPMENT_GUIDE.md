# Development Guide

## Getting Started

### Prerequisites

- **Go 1.25+** - [Install Go](https://go.dev/doc/install)
- **PostgreSQL 14+** - [Install PostgreSQL](https://www.postgresql.org/download/)
- **sqlc** - [Install sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- **Clerk Account** - For authentication (or use test keys)

### Initial Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd url-shortener
   ```

2. **Set up environment variables**
   ```bash
   cd server
   cp .env.example .env  # If exists, or create .env
   ```
   
   Required variables:
   ```env
   APP_ENV=dev
   PORT=8080
   POSTGRES_CONNECTION_STRING=postgres://user:password@localhost:5432/urlshortener?sslmode=disable
   CLERK_SECRET_KEY=sk_test_...
   ```

3. **Set up database**
   ```bash
   # Create database
   createdb urlshortener
   
   # Run migrations (if using migration tool)
   # Or manually run SQL from migrations/
   ```

4. **Generate database code**
   ```bash
   cd server
   sqlc generate
   ```

5. **Install dependencies**
   ```bash
   go mod download
   ```

6. **Run the server**
   ```bash
   go run cmd/main.go
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**
   - Follow patterns in existing code
   - See `SERVICE_PATTERNS.md` and `HANDLER_PATTERNS.md`

3. **Run tests**
   ```bash
   go test ./...
   ```

4. **Check for issues**
   ```bash
   go vet ./...
   go fmt ./...
   ```

5. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/my-feature
   ```

### Adding Database Changes

1. **Create migration**
   ```bash
   # Create migration file
   touch migrations/000002_my_change.up.sql
   touch migrations/000002_my_change.down.sql
   ```

2. **Write SQL**
   ```sql
   -- migrations/000002_my_change.up.sql
   ALTER TABLE links ADD COLUMN new_field TEXT;
   ```

   ```sql
   -- migrations/000002_my_change.down.sql
   ALTER TABLE links DROP COLUMN new_field;
   ```

3. **Add queries** (if needed)
   ```sql
   -- queries/links.sql
   -- name: GetLinkWithNewField :one
   SELECT * FROM links WHERE new_field = $1;
   ```

4. **Generate code**
   ```bash
   sqlc generate
   ```

5. **Update service/handler** as needed

### Adding New Endpoints

See `HANDLER_PATTERNS.md` for detailed guide.

Quick checklist:
- [ ] Add route in `pkg/router/router.go`
- [ ] Add handler method in `pkg/handlers/`
- [ ] Add service method if needed in `pkg/service/`
- [ ] Add DTOs in `pkg/dto/`
- [ ] Add tests
- [ ] Update OpenAPI spec if needed

## Testing

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/service/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Writing Tests

**Service Tests**:
- Mock the `LinkQueries` interface
- Test business logic
- Test error cases

**Handler Tests**:
- Mock the `LinkServiceInterface`
- Test HTTP request/response
- Test error handling

**See**: Test files in `pkg/service/link_test.go` and `pkg/handlers/link_test.go` for examples.

### Test Structure

```go
func TestService_Method(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() *mockQueries
        input   string
        want    db.Link
        wantErr bool
    }{
        {
            name: "success case",
            setup: func() *mockQueries {
                return &mockQueries{
                    GetLinkFunc: func(...) (db.Link, error) {
                        return testLink, nil
                    },
                }
            },
            input: "test",
            want: testLink,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Code Quality

### Linting

```bash
# Run linter (if configured)
golangci-lint run

# Or use go vet
go vet ./...
```

### Formatting

```bash
# Format code
go fmt ./...

# Or use goimports
goimports -w .
```

### Code Review Checklist

Before submitting PR:

- [ ] Code follows existing patterns
- [ ] Tests added/updated
- [ ] No linter errors
- [ ] Documentation updated if needed
- [ ] Error handling follows patterns
- [ ] Interfaces used for testability
- [ ] No hardcoded values (use config)

## Debugging

### Logging

The logger is configured in `cmd/main.go`:

```go
log, err := logger.New(cfg.AppEnv)
```

In development, logs are pretty-printed. In production, they're JSON.

### Adding Logs

Use zap fields for type-safe logging:

```go
import "go.uber.org/zap"

// Info level
h.logger.Info("Operation completed",
    zap.String("user_id", userID),
    zap.String("link_id", linkID),
)

// Error level
h.logger.Error("Operation failed",
    zap.Error(err),
    zap.String("user_id", userID),
)

// With multiple fields
h.logger.Info("Request processed",
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.Int("status", statusCode),
    zap.Duration("latency", elapsed),
)
```

### Debugging Tips

1. **Check logs**: Look at console output for errors
2. **Use debugger**: Set breakpoints in IDE
3. **Test endpoints**: Use curl or Postman
4. **Check database**: Verify data directly
5. **Check context**: Ensure userID is set correctly

## Common Tasks

### Adding a New Service

1. Create `pkg/service/new_service.go`
2. Define interface in `pkg/service/new_service.go` (for testing)
3. Implement service methods
4. Add tests in `pkg/service/new_service_test.go`
5. Wire up in `pkg/server.go`

See `SERVICE_PATTERNS.md` for details.

### Adding a New Handler

1. Create `pkg/handlers/new_handler.go`
2. Define service interface in handler file
3. Implement handler methods
4. Add tests
5. Add routes in `pkg/router/router.go`

See `HANDLER_PATTERNS.md` for details.

### Adding New Error Types

1. Add error code in `pkg/errors/errors.go`:
   ```go
   const ErrorNewError ErrorCode = "NEW_ERROR"
   ```

2. Add sentinel error (if needed):
   ```go
   var ErrNewError = errors.New("new error")
   ```

3. Update handler's `handleError()` method to handle the new error:
   ```go
   case errors.Is(err, apperrors.ErrNewError):
       h.logger.Warn("New error occurred", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
       render.Status(r, http.StatusBadRequest)
       render.JSON(w, r, apperrors.ErrorResponse{
           Error: apperrors.ErrorDetail{
               Code:    apperrors.ErrorCodeNewError,
               Message: "Message",
           },
       })
   ```

### Adding Middleware

1. Create `pkg/middleware/new_middleware.go`
2. Implement middleware function
3. Add to chain in `pkg/server.go` or `pkg/router/router.go`
4. Consider execution order

## Environment Setup

### Development

```env
APP_ENV=dev
PORT=8080
POSTGRES_CONNECTION_STRING=postgres://localhost/urlshortener
CLERK_SECRET_KEY=sk_test_...
```

### Production

```env
APP_ENV=production
PORT=8080
POSTGRES_CONNECTION_STRING=postgres://...
CLERK_SECRET_KEY=sk_live_...
```

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
pg_isready

# Test connection
psql $POSTGRES_CONNECTION_STRING
```

### sqlc Generation Issues

```bash
# Check sqlc config
cat sqlc.yaml

# Regenerate
sqlc generate

# Check for syntax errors in SQL
```

### Import Errors

```bash
# Download dependencies
go mod download

# Tidy modules
go mod tidy
```

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

## Useful Commands

```bash
# Run server
go run cmd/main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Generate database code
sqlc generate

# Download dependencies
go mod download

# Tidy modules
go mod tidy
```

## IDE Setup

### VS Code

Recommended extensions:
- Go extension
- SQLC extension (if available)
- Error Lens

### GoLand

- Configure Go SDK
- Enable Go modules
- Set up database connection for migrations

## Further Reading

- `ARCHITECTURE.md` - Overall architecture
- `PROJECT_STRUCTURE.md` - Directory structure

