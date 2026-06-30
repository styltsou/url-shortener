# Backend TODOs

## Infrastructure

### 1. Docker Compose Improvements
- Pin ClickHouse image version (currently `latest`)
- Add health check for server dependency
- Add network isolation between services
- Fix empty `REDIS_USERNAME`/`REDIS_PASSWORD` env vars

### 2. Dockerfile Build Context
- Builder stage copies `docs/` unnecessarily — exclude from build context

### 3. Add `.dockerignore` Files
- Missing in both `server/` and `client/` directories
