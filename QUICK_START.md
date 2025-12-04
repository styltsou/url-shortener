# Quick Start Guide for New Developers

## Welcome! 👋

This guide will help you get up to speed quickly. Follow these steps in order.

## Step 1: Read the Overview (15 minutes)

1. **`README.md`** - Project overview and setup
2. **`PROJECT_OVERVIEW.md`** - What the project does, features, tech stack
3. **`ARCHITECTURE.md`** - High-level architecture and request flow

**Goal**: Understand what we're building and how it's structured.

## Step 2: Understand the Structure (10 minutes)

1. **`PROJECT_STRUCTURE.md`** - Directory structure, what goes where

**Goal**: Know where to find and add code.

## Step 3: Learn the Patterns (30 minutes)

1. **`INTERFACES_GUIDE.md`** - Why we use interfaces (important!)
2. **`ERROR_HANDLING_GUIDE.md`** - How errors work
3. **`SERVICE_PATTERNS.md`** - How to create services
4. **`HANDLER_PATTERNS.md`** - How to create handlers

**Goal**: Understand the patterns we follow.

## Step 4: Set Up Your Environment (15 minutes)

1. **`DEVELOPMENT_GUIDE.md`** - Setup instructions
2. Follow the setup steps
3. Run the server and verify it works

**Goal**: Get the code running locally.

## Step 5: Trace a Request (20 minutes)

Follow a real request through the codebase:

1. Start at `cmd/main.go` - see how server starts
2. Go to `pkg/router/router.go` - see route definitions
3. Go to `pkg/handlers/link.go` - see `CreateLink` handler
4. Go to `pkg/service/link.go` - see `CreateShortLink` service
5. Go to `pkg/db/links.sql.go` - see database query

**Goal**: Understand how code flows.

## Step 6: Read Existing Code (30 minutes)

1. Read `pkg/service/link.go` - understand service pattern
2. Read `pkg/handlers/link.go` - understand handler pattern
3. Read `pkg/middleware/error_handler.go` - understand error handling
4. Read test files - understand testing patterns

**Goal**: See patterns in action.

## Step 7: Make Your First Change

Try adding a simple feature:

1. Add a new endpoint (follow `HANDLER_PATTERNS.md`)
2. Add tests (follow existing test patterns)
3. Run tests and verify

**Goal**: Apply what you learned.

## Common Tasks Quick Reference

### Adding a New Endpoint

1. Add route in `pkg/router/router.go`
2. Add handler method in `pkg/handlers/`
3. Add service method if needed in `pkg/service/`
4. Add DTOs in `pkg/dto/`
5. Add tests

**See**: `HANDLER_PATTERNS.md` for details

### Adding a New Service

1. Create `pkg/service/new_service.go`
2. Define interface for dependencies
3. Implement business logic
4. Add tests

**See**: `SERVICE_PATTERNS.md` for details

### Adding New Error Types

1. Add error code in `pkg/errors/errors.go`
2. Add sentinel error (if needed)
3. Update handler's `handleError()` method to handle the new error

**See**: `ERROR_HANDLING_GUIDE.md` for details

### Adding Database Changes

1. Create migration in `migrations/`
2. Add queries in `queries/`
3. Run `sqlc generate`
4. Update service/handler

**See**: `DEVELOPMENT_GUIDE.md` for details

## Documentation Index

### Architecture & Design
- **`ARCHITECTURE.md`** - Overall architecture, layers, request flow
- **`PROJECT_STRUCTURE.md`** - Directory structure, what goes where
- **`INTERFACES_GUIDE.md`** - Why and how we use interfaces

### Patterns & Practices
- **`SERVICE_PATTERNS.md`** - How to create and structure services
- **`HANDLER_PATTERNS.md`** - How to create and structure handlers
- **`ERROR_HANDLING_GUIDE.md`** - Error handling architecture

### Development
- **`DEVELOPMENT_GUIDE.md`** - Setup, workflow, testing, debugging
- **`QUICK_START.md`** - This file - getting started guide

### Project Info
- **`PROJECT_OVERVIEW.md`** - Project scope, features, tech stack
- **`TECH_ROADMAP.md`** - Future plans and evolution
- **`LEARNING_PATH.md`** - Learning progression guide

### Other
- **`ERROR_HANDLING_GUIDE.md`** - Detailed error handling guide
- **`LOGGING_GUIDE.md`** - Logging patterns and practices

## Key Concepts to Remember

### 1. **Layered Architecture**
- Router → Handler → Service → Repository
- Dependencies flow downward only

### 2. **Interface-Based Design**
- Services use interfaces for database access
- Handlers use interfaces for services
- Enables testing with mocks

### 3. **Error Handling**
- Services return errors (sentinel errors)
- Handlers set errors in context
- Middleware handles errors

### 4. **Single Responsibility**
- Each layer has one clear purpose
- Handlers = HTTP concerns
- Services = Business logic
- Repository = Data access

## Getting Help

1. **Check documentation** - Most questions are answered in the docs
2. **Look at existing code** - Follow established patterns
3. **Read tests** - Tests show how code is used
4. **Ask the team** - Don't hesitate to ask!

## Next Steps

After completing the quick start:

1. **Read `LEARNING_PATH.md`** - Deep dive into specific areas
2. **Explore the codebase** - Read through different files
3. **Make changes** - Start contributing!
4. **Update docs** - If you find gaps, document them

## Checklist for New Features

When adding a new feature, make sure you:

- [ ] Follow existing patterns
- [ ] Add tests
- [ ] Handle errors correctly
- [ ] Use interfaces for dependencies
- [ ] Update documentation if needed
- [ ] Run linter/formatter
- [ ] Test locally

## Common Mistakes to Avoid

- ❌ Don't put business logic in handlers
- ❌ Don't access database directly from handlers
- ❌ Don't handle errors by writing responses in handlers
- ❌ Don't edit generated files (`pkg/db/*.sql.go`)
- ❌ Don't create circular dependencies
- ❌ Don't skip tests

## Questions?

If you have questions:

1. Check the relevant documentation file
2. Look at similar code in the codebase
3. Ask the team

Welcome to the team! 🚀

