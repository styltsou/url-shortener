# Backend TODOs

## Infrastructure

### 1. Docker Compose Improvements
- Add external uptime monitoring against `/api/v1/health/ready`
- Add network isolation between services
- Fix empty `REDIS_USERNAME`/`REDIS_PASSWORD` env vars

### 2. Dockerfile Build Context
- Builder stage copies `docs/` unnecessarily — exclude from build context
