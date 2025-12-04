# Technology Roadmap

This document tracks new technologies and patterns being introduced to the project for learning purposes.

## 🎯 Learning Goals

These technologies are being introduced to:
- Learn industry-standard Go patterns
- Improve codebase maintainability
- Prepare for production-scale applications
- Build job-ready skills

---

## 📋 Technologies to Implement

### 1. **Viper** - Configuration Management
**Status:** 🔄 In Progress  
**Priority:** High (Quick Win)  
**Estimated Time:** 30-60 minutes

**What it is:**
- Industry-standard configuration library for Go
- Supports multiple config sources (env vars, files, flags, etc.)
- Type-safe configuration with validation

**Why we're using it:**
- Replace manual env var handling
- Support multiple environments (dev/staging/prod)
- Better validation and error handling
- Industry standard pattern

**Learning outcomes:**
- Viper configuration patterns
- Struct-based config (common in production)
- Environment-based configuration
- Config validation

**Resources:**
- [Viper GitHub](https://github.com/spf13/viper)
- [Viper Documentation](https://github.com/spf13/viper#readme)

---

### 2. **Wire** - Dependency Injection
**Status:** ⏳ Planned  
**Priority:** High (Before ClickHouse)  
**Estimated Time:** 1-2 hours

**What it is:**
- Code generation tool for dependency injection
- Compile-time DI (no runtime overhead)
- Generates explicit initialization code

**Why we're using it:**
- Industry standard for production Go applications
- Clean dependency management
- Makes adding new dependencies (like ClickHouse) easier
- Type-safe dependency graphs

**Learning outcomes:**
- Wire provider functions
- Wire provider sets
- Understanding generated code patterns
- Dependency graph management

**Resources:**
- [Wire GitHub](https://github.com/google/wire)
- [Wire Tutorial](https://github.com/google/wire/blob/main/docs/guide.md)

---

### 3. **ClickHouse** - Analytics Database
**Status:** ⏳ Planned  
**Priority:** Medium (After Wire)  
**Estimated Time:** 2-4 hours

**What it is:**
- Column-oriented database optimized for analytics
- Fast aggregations and time-series data
- Perfect for click tracking and analytics

**Why we're using it:**
- Store click events and analytics
- Fast aggregations (clicks per link, time-series data)
- Separate analytics from transactional data (PostgreSQL)
- Industry standard for analytics workloads

**Learning outcomes:**
- Column-oriented databases
- Time-series data patterns
- Analytics query patterns
- Multi-database architecture

**Resources:**
- [ClickHouse Documentation](https://clickhouse.com/docs)
- [ClickHouse Go Driver](https://github.com/ClickHouse/clickhouse-go)

---

## 📊 Implementation Order

1. ✅ **Viper** - Quick win, improves config management
2. ⏳ **Wire** - Sets up DI before adding more dependencies
3. ⏳ **ClickHouse** - Feature addition, benefits from Wire setup

---

## 📝 Notes

- Each technology will be implemented with learning in mind
- Code will be reviewed for best practices
- Documentation will be updated as we progress
- Focus on understanding patterns, not just implementation


