# 📚 Learning Path - Understanding the Codebase

Follow this reading order to understand the codebase and established patterns before developing.

---

## 🎯 Phase 1: Foundation (30 minutes)

**Goal**: Understand what we're building and the high-level architecture.

### 1. **`PROJECT_OVERVIEW.md`** (15 min)
- What the project does
- Features and scope
- Tech stack overview
- Database models

**Why first**: Get context about the domain and goals.

### 2. **`ARCHITECTURE.md`** (15 min)
- High-level architecture
- Request flow (HTTP → Handler → Service → Database)
- Layer responsibilities
- Key design decisions

**Why second**: Understand how everything fits together.

---

## 🗂️ Phase 2: Structure (10 minutes)

**Goal**: Know where to find and add code.

### 3. **`PROJECT_STRUCTURE.md`** (10 min)
- Directory structure
- What goes where
- File naming conventions
- Package organization

**Why third**: Navigate the codebase efficiently.

---

## 🎨 Phase 3: Core Patterns (60 minutes)

**Goal**: Understand the established patterns you must follow.

### 4. **`INTERFACES_GUIDE.md`** (15 min) ⭐ **IMPORTANT**
- Why we use interfaces
- How to design interfaces
- Testing with interfaces
- Dependency injection

**Why important**: Interfaces are fundamental to our architecture.

### 5. **`ERROR_HANDLING_GUIDE.md`** (20 min) ⭐ **CRITICAL**
- Sentinel errors (service layer)
- Error codes (HTTP layer)
- Error mapping
- Handler patterns
- Best practices

**Why critical**: Error handling is used everywhere. Understanding this is essential.

### 6. **`SERVICE_PATTERNS.md`** (15 min)
- How to create services
- Service responsibilities
- Business logic patterns
- Database interaction

**Why**: Services contain your business logic.

### 7. **`HANDLER_PATTERNS.md`** (15 min)
- How to create handlers
- Handler responsibilities
- Request/response patterns
- Error handling in handlers

**Why**: Handlers are your HTTP layer.

---

## 🛠️ Phase 4: Setup & Trace (35 minutes)

**Goal**: Get the code running and trace a real request.

### 8. **`DEVELOPMENT_GUIDE.md`** (15 min)
- Setup instructions
- Environment variables
- Running the server
- Development workflow

**Why**: Get your environment ready.

### 9. **Trace a Request** (20 min)
Follow a real request through the codebase:

1. **`cmd/main.go`** - Server startup
   - How the server initializes
   - Dependency wiring
   - Middleware setup

2. **`pkg/router/router.go`** - Route definitions
   - How routes are defined
   - Middleware chain
   - Route grouping

3. **`pkg/handlers/link.go`** - `CreateLink` handler
   - Request parsing
   - Service calls
   - Error handling
   - Response rendering

4. **`pkg/service/link.go`** - `CreateShortLink` service
   - Business logic
   - Validation
   - Database calls
   - Error wrapping

5. **`pkg/db/links.sql.go`** - Database query
   - Generated code from sqlc
   - Query execution

**Why**: See how code flows in practice.

---

## 📖 Phase 5: Deep Dive (Optional but Recommended)

**Goal**: Understand advanced topics and best practices.

### 10. **`CODE_REVIEW.md`** (20 min)
- What's good in the codebase
- Areas for improvement
- Best practices
- Common pitfalls to avoid

**Why**: Learn from the review - what to do and what to avoid.

### 11. **Additional Guides** (as needed)
- **`LOGGING_GUIDE.md`** - Structured logging patterns
- **`TECH_ROADMAP.md`** - Future plans and features

---

## 🚀 Quick Reference

### When Adding a New Feature

1. **Read**: `SERVICE_PATTERNS.md` and `HANDLER_PATTERNS.md`
2. **Check**: `ERROR_HANDLING_GUIDE.md` for error patterns
3. **Follow**: Existing code in `pkg/service/` and `pkg/handlers/`
4. **Test**: Write tests following existing test patterns

### When Debugging

1. **Check**: `ERROR_HANDLING_GUIDE.md` - understand error flow
2. **Trace**: Follow the request through handlers → services → database
3. **Logs**: Check structured logs (see `LOGGING_GUIDE.md`)

### When Adding a New Resource

1. **Create**: SQL migration in `migrations/`
2. **Generate**: Run `sqlc generate`
3. **Create**: Service in `pkg/service/`
4. **Create**: Handler in `pkg/handlers/`
5. **Add**: Routes in `pkg/router/router.go`
6. **Test**: Write tests for service and handler

---

## ✅ Checklist

Before you start developing, make sure you:

- [ ] Understand the project overview and goals
- [ ] Know the architecture and request flow
- [ ] Can navigate the directory structure
- [ ] Understand interfaces and dependency injection
- [ ] Know how error handling works (sentinel errors, error codes, mapping)
- [ ] Understand service patterns
- [ ] Understand handler patterns
- [ ] Have the code running locally
- [ ] Can trace a request through the codebase

---

## 🎓 Key Concepts to Master

1. **Interfaces** - Everything is interface-based for testability
2. **Error Handling** - Sentinel errors → Error codes → HTTP responses
3. **Layered Architecture** - Handlers → Services → Database
4. **Context Usage** - Request context for cancellation and values
5. **Dependency Injection** - Dependencies passed via constructors
6. **Type Safety** - sqlc generates type-safe database code

---

## 💡 Pro Tips

- **Start small**: Understand one handler/service pair completely before moving on
- **Read tests**: Tests show how code is used and expected to behave
- **Follow patterns**: Don't invent new patterns - follow existing ones
- **Ask questions**: If something doesn't make sense, check the guides first
- **Trace errors**: When debugging, trace errors through the error handling flow

---

**Total Time**: ~2.5 hours for complete understanding

**Minimum Time**: 1 hour (Phases 1-3) to get started

**Ready to code?** Start with Phase 1! 🚀
