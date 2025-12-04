# Logging Guide: From Development to Production

A comprehensive guide to understanding logging, structured logging, and how logging works in development vs production environments.

---

## Table of Contents

1. [What is Logging?](#what-is-logging)
2. [Why Do We Log?](#why-do-we-log)
3. [Types of Logging](#types-of-logging)
4. [Log Levels](#log-levels)
5. [Structured Logging](#structured-logging)
6. [Development vs Production](#development-vs-production)
7. [Use Cases and When to Log](#use-cases-and-when-to-log)
8. [Log Aggregation and Analysis](#log-aggregation-and-analysis)
9. [Best Practices](#best-practices)
10. [Common Patterns](#common-patterns)

---

## What is Logging?

**Logging** is the practice of recording events, messages, and data that occur during application execution. These records (called "logs") are written to various outputs like:

- Console/terminal (development)
- Files (local or remote)
- Databases
- Log aggregation services (cloud platforms)
- Standard output streams (stdout/stderr)

Think of logging as a **black box recorder** for your application - it captures what happened, when it happened, and in what context.

---

## Why Do We Log?

### Primary Purposes

1. **Debugging**: Understand what went wrong when things break
2. **Monitoring**: Track application health and performance
3. **Auditing**: Record important business events and user actions
4. **Troubleshooting**: Diagnose issues in production without direct access
5. **Analytics**: Understand user behavior and system usage patterns
6. **Compliance**: Meet regulatory requirements (GDPR, HIPAA, etc.)

### The Problem Without Logging

Without logging, when something breaks in production:
- ❌ You have no idea what happened
- ❌ You can't reproduce the issue
- ❌ You can't see the sequence of events
- ❌ Debugging becomes guesswork
- ❌ You can't track down the root cause

### The Solution With Logging

With proper logging:
- ✅ You can trace exactly what happened
- ✅ You can see the sequence of events leading to the error
- ✅ You can identify patterns and trends
- ✅ You can debug issues without reproducing them
- ✅ You have a historical record of system behavior

---

## Types of Logging

### 1. Unstructured Logging (Traditional)

**What it is:**
- Plain text messages
- Human-readable format
- Free-form text

**Example:**
```
2024-01-15 10:30:45 User john@example.com logged in successfully
2024-01-15 10:31:12 Error: Database connection failed
2024-01-15 10:32:00 Processing payment for order #12345
```

**Pros:**
- Easy to read for humans
- Simple to implement
- Good for development

**Cons:**
- Hard to parse programmatically
- Difficult to search and filter
- Inconsistent format
- Not scalable for large systems

### 2. Structured Logging (Modern Standard)

**What it is:**
- Machine-readable format (usually JSON)
- Consistent key-value pairs
- Easy to parse and query

**Example:**
```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "User logged in",
  "user_id": "12345",
  "email": "john@example.com",
  "ip_address": "192.168.1.1",
  "request_id": "req-abc-123"
}
```

**Pros:**
- ✅ Easy to parse programmatically
- ✅ Powerful search and filtering
- ✅ Consistent format
- ✅ Scalable for large systems
- ✅ Works with log aggregation tools
- ✅ Can add context easily

**Cons:**
- Less human-readable (but tools fix this)
- Slightly more verbose

### 3. Binary Logging

**What it is:**
- Compact binary format
- Maximum performance
- Used in high-throughput systems

**Example:**
- Protocol Buffers (protobuf)
- Apache Avro
- Custom binary formats

**When to use:**
- Extremely high-volume logging
- Performance-critical systems
- When storage/bandwidth is a concern

---

## Log Levels

Log levels indicate the **severity** or **importance** of a log message. They help filter and prioritize logs.

### Standard Log Levels (from most to least severe)

1. **FATAL / CRITICAL**
   - Application cannot continue
   - Immediate attention required
   - Usually triggers alerts
   - Example: Database connection completely lost, cannot recover

2. **ERROR**
   - Something went wrong, but application continues
   - Needs investigation
   - May trigger alerts
   - Example: Failed to process a request, but other requests work

3. **WARN / WARNING**
   - Something unusual happened, but not necessarily wrong
   - Should be monitored
   - Example: Slow database query, high memory usage

4. **INFO**
   - Normal application flow
   - Important business events
   - Example: User logged in, order created, payment processed

5. **DEBUG**
   - Detailed information for debugging
   - Usually disabled in production
   - Example: Function entry/exit, variable values, detailed flow

6. **TRACE** (optional)
   - Very detailed, low-level information
   - Usually only for development
   - Example: Every function call, every network packet

### When to Use Each Level

| Level | When to Use | Production? |
|-------|-------------|------------|
| FATAL | Application cannot continue | ✅ Always log |
| ERROR | Errors that need attention | ✅ Always log |
| WARN | Unusual but not critical | ✅ Usually log |
| INFO | Important business events | ✅ Always log |
| DEBUG | Development debugging | ❌ Usually disabled |
| TRACE | Very detailed debugging | ❌ Never in production |

### Log Level Best Practices

- **Don't log everything at ERROR** - Reserve for actual errors
- **Use INFO for important events** - Not for every function call
- **Use DEBUG for development** - Disable in production for performance
- **Be consistent** - Use the same level for similar events

---

## Structured Logging

### What Makes Logging "Structured"?

Structured logging means logs are in a **consistent, machine-readable format** (usually JSON) with **key-value pairs**.

### Key Characteristics

1. **Consistent Format**: All logs follow the same structure
2. **Key-Value Pairs**: Data is organized as fields
3. **Machine-Readable**: Easy to parse programmatically
4. **Queryable**: Can search and filter by any field
5. **Extensible**: Easy to add new fields

### Structured Log Example

```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "message": "User logged in successfully",
  "user_id": "12345",
  "email": "john@example.com",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "request_id": "req-abc-123",
  "duration_ms": 45,
  "service": "auth-service",
  "environment": "production"
}
```

### Benefits of Structured Logging

1. **Searchability**
   ```
   Find all logs where user_id = "12345"
   Find all errors in the last hour
   Find all slow requests (>1000ms)
   ```

2. **Aggregation**
   ```
   Count errors per service
   Average response time per endpoint
   Most common error types
   ```

3. **Context**
   - Every log entry can include relevant context
   - Request IDs, user IDs, transaction IDs
   - Makes tracing issues much easier

4. **Integration**
   - Works seamlessly with log aggregation tools
   - Can be parsed and analyzed automatically
   - Can trigger alerts based on patterns

### Structured Logging Fields

Common fields to include:

- **Timestamp**: When the event occurred
- **Level**: Log severity (info, error, etc.)
- **Message**: Human-readable description
- **Request ID**: Trace requests across services
- **User ID**: Identify which user triggered the event
- **Service/Component**: Which part of the system
- **Environment**: dev, staging, production
- **Duration**: How long something took
- **Error details**: Stack traces, error codes

---

## Development vs Production

### Development Logging

**Goals:**
- Help developers debug quickly
- Be human-readable
- Show detailed information
- Fast iteration

**Characteristics:**
- **Format**: Pretty-printed, colored output
- **Level**: DEBUG and TRACE enabled
- **Output**: Console/terminal
- **Detail**: Very verbose, includes variable values
- **Performance**: Less concern about log volume

**Example (Development):**
```
2024-01-15 10:30:45.123 [INFO]  User logged in successfully
  user_id: 12345
  email: john@example.com
  ip: 192.168.1.1
  duration: 45ms
```

### Production Logging

**Goals:**
- Monitor system health
- Debug issues without code changes
- Track business metrics
- Meet compliance requirements

**Characteristics:**
- **Format**: JSON (structured)
- **Level**: INFO, WARN, ERROR, FATAL (no DEBUG)
- **Output**: Log aggregation service (not console)
- **Detail**: Balanced - enough info but not too verbose
- **Performance**: Optimized, sampled for high volume

**Example (Production):**
```json
{"timestamp":"2024-01-15T10:30:45.123Z","level":"info","message":"User logged in","user_id":"12345","email":"john@example.com","ip":"192.168.1.1","duration_ms":45,"request_id":"req-abc-123"}
```

### Key Differences

| Aspect | Development | Production |
|--------|-------------|------------|
| **Format** | Pretty text | JSON |
| **Readability** | Human-friendly | Machine-friendly |
| **Level** | DEBUG enabled | DEBUG disabled |
| **Volume** | High (detailed) | Moderate (important events) |
| **Output** | Console | Log aggregation service |
| **Performance** | Less critical | Critical |
| **Sampling** | None | May use sampling |

---

## Use Cases and When to Log

### 1. Error Logging

**When:** Every time an error occurs

**What to log:**
- Error message
- Stack trace
- Request context (user, request ID)
- Input parameters (sanitized)
- Timestamp

**Example:**
```json
{
  "level": "error",
  "message": "Failed to process payment",
  "error": "Insufficient funds",
  "user_id": "12345",
  "order_id": "67890",
  "amount": 99.99,
  "request_id": "req-abc-123",
  "stack_trace": "..."
}
```

### 2. Request/Response Logging

**When:** Important API requests (not every request - too verbose)

**What to log:**
- Request method and path
- Response status code
- Duration
- Request ID (for tracing)
- User ID (if authenticated)

**Example:**
```json
{
  "level": "info",
  "message": "API request completed",
  "method": "POST",
  "path": "/api/v1/orders",
  "status_code": 201,
  "duration_ms": 145,
  "request_id": "req-abc-123",
  "user_id": "12345"
}
```

### 3. Business Event Logging

**When:** Important business events

**What to log:**
- Event type
- User/actor
- Relevant IDs
- Timestamp

**Example:**
```json
{
  "level": "info",
  "message": "Order created",
  "event": "order.created",
  "order_id": "67890",
  "user_id": "12345",
  "amount": 99.99,
  "items_count": 3
}
```

### 4. Performance Logging

**When:** Slow operations or performance-critical paths

**What to log:**
- Operation name
- Duration
- Resource usage (if relevant)
- Context

**Example:**
```json
{
  "level": "warn",
  "message": "Slow database query",
  "query": "SELECT * FROM orders WHERE user_id = ?",
  "duration_ms": 1250,
  "threshold_ms": 1000
}
```

### 5. Security Logging

**When:** Security-related events

**What to log:**
- Event type (login, logout, permission denied, etc.)
- User ID
- IP address
- Timestamp
- Success/failure

**Example:**
```json
{
  "level": "warn",
  "message": "Failed login attempt",
  "event": "auth.failed_login",
  "email": "user@example.com",
  "ip_address": "192.168.1.1",
  "reason": "Invalid password"
}
```

### 6. State Change Logging

**When:** Important state transitions

**What to log:**
- Previous state
- New state
- What triggered the change
- Who triggered it

**Example:**
```json
{
  "level": "info",
  "message": "Order status changed",
  "order_id": "67890",
  "previous_status": "pending",
  "new_status": "paid",
  "changed_by": "user_12345"
}
```

### When NOT to Log

- ❌ **Sensitive data**: Passwords, credit cards, tokens (log that it happened, not the value)
- ❌ **Every function call**: Too verbose, use DEBUG level
- ❌ **High-frequency events**: Use sampling or aggregation
- ❌ **Redundant information**: Don't log the same thing multiple times
- ❌ **Personal data**: Be GDPR/compliance aware

---

## Log Aggregation and Analysis

### What is Log Aggregation?

**Log aggregation** is the practice of collecting logs from multiple sources into a central location for analysis, search, and monitoring.

### Why Aggregate Logs?

1. **Centralized View**: All logs in one place
2. **Search Across Services**: Find related logs across multiple services
3. **Correlation**: Connect events across different parts of the system
4. **Analytics**: Analyze patterns and trends
5. **Alerting**: Set up alerts based on log patterns

### Common Log Aggregation Tools

1. **ELK Stack** (Elasticsearch, Logstash, Kibana)
   - Open source
   - Powerful search and visualization
   - Self-hosted or cloud

2. **Loki** (Grafana Labs)
   - Lightweight
   - Prometheus-inspired
   - Good for Kubernetes

3. **Cloud Services**
   - **AWS CloudWatch Logs**
   - **Google Cloud Logging**
   - **Azure Monitor**
   - **Datadog**
   - **New Relic**
   - **Splunk**

### How Log Aggregation Works

```
Application → Logs → Log Collector → Log Aggregation Service → Analysis/Visualization
```

1. **Application** writes structured logs (JSON)
2. **Log Collector** (agent) collects logs from files/stdout
3. **Log Aggregation Service** stores and indexes logs
4. **Analysis Tools** search, filter, visualize, and alert

### Log Analysis Use Cases

1. **Error Tracking**
   - Find all errors in the last hour
   - Group errors by type
   - Track error trends

2. **Performance Monitoring**
   - Average response times
   - Slowest endpoints
   - Resource usage patterns

3. **User Behavior**
   - Most used features
   - User journey tracking
   - Conversion funnels

4. **Security Monitoring**
   - Failed login attempts
   - Unusual access patterns
   - Potential attacks

5. **Business Metrics**
   - Orders per hour
   - Revenue tracking
   - Feature usage

---

## Best Practices

### 1. Use Structured Logging

✅ **DO:**
```json
{"level":"info","message":"User logged in","user_id":"12345","ip":"192.168.1.1"}
```

❌ **DON'T:**
```
User 12345 logged in from 192.168.1.1
```

### 2. Include Context

✅ **DO:** Include relevant context (request ID, user ID, etc.)
```json
{"level":"error","message":"Payment failed","order_id":"67890","user_id":"12345","error":"Insufficient funds"}
```

❌ **DON'T:** Log without context
```json
{"level":"error","message":"Payment failed"}
```

### 3. Use Appropriate Log Levels

✅ **DO:**
- ERROR for actual errors
- INFO for important events
- DEBUG for development only

❌ **DON'T:**
- Log everything as ERROR
- Use INFO for every function call
- Enable DEBUG in production

### 4. Don't Log Sensitive Data

✅ **DO:**
```json
{"level":"info","message":"User logged in","user_id":"12345"}
```

❌ **DON'T:**
```json
{"level":"info","message":"User logged in","password":"secret123","credit_card":"1234-5678-9012-3456"}
```

### 5. Use Request IDs for Tracing

✅ **DO:** Include request ID in all related logs
```json
{"request_id":"req-abc-123","message":"Processing order"}
{"request_id":"req-abc-123","message":"Order validated"}
{"request_id":"req-abc-123","message":"Order created"}
```

### 6. Log at Appropriate Boundaries

✅ **DO:** Log at service boundaries, important events
- API request/response
- Database operations (important ones)
- External service calls
- Business events

❌ **DON'T:** Log inside every small function

### 7. Make Logs Searchable

✅ **DO:** Use consistent field names
```json
{"user_id":"12345","order_id":"67890"}
```

❌ **DON'T:** Use inconsistent names
```json
{"userId":"12345","OrderID":"67890","user":"12345"}
```

### 8. Include Timestamps

✅ **DO:** Always include timestamps (usually automatic)
```json
{"timestamp":"2024-01-15T10:30:45Z","message":"..."}
```

### 9. Use Sampling for High-Volume Logs

✅ **DO:** Sample high-frequency events
- Log 1 in 100 requests instead of all
- Still log all errors

### 10. Monitor Your Logs

✅ **DO:**
- Set up alerts for errors
- Monitor log volume
- Review logs regularly
- Set up dashboards

---

## Common Patterns

### 1. Request Tracing

**Pattern:** Include request ID in all logs for a request

```json
{"request_id":"req-123","message":"Request started","method":"POST","path":"/api/orders"}
{"request_id":"req-123","message":"Validating input"}
{"request_id":"req-123","message":"Creating order","order_id":"67890"}
{"request_id":"req-123","message":"Request completed","status":201,"duration_ms":145}
```

### 2. Error Context

**Pattern:** Include full context when logging errors

```json
{
  "level":"error",
  "message":"Failed to process payment",
  "error":"Insufficient funds",
  "user_id":"12345",
  "order_id":"67890",
  "amount":99.99,
  "request_id":"req-123",
  "stack_trace":"..."
}
```

### 3. Performance Monitoring

**Pattern:** Log duration for important operations

```json
{"level":"info","message":"Database query completed","query":"SELECT...","duration_ms":45}
{"level":"warn","message":"Slow operation","operation":"process_payment","duration_ms":1250}
```

### 4. Business Events

**Pattern:** Log important business events with all relevant data

```json
{
  "level":"info",
  "message":"Order created",
  "event":"order.created",
  "order_id":"67890",
  "user_id":"12345",
  "amount":99.99,
  "items":3,
  "currency":"USD"
}
```

### 5. Security Events

**Pattern:** Log all security-related events

```json
{
  "level":"warn",
  "message":"Failed login attempt",
  "event":"auth.failed_login",
  "email":"user@example.com",
  "ip_address":"192.168.1.1",
  "user_agent":"Mozilla/5.0...",
  "reason":"Invalid password"
}
```

---

## Summary

### Key Takeaways

1. **Logging is essential** for debugging, monitoring, and understanding your application
2. **Structured logging** (JSON) is the modern standard for production
3. **Log levels** help prioritize and filter logs
4. **Development** uses pretty, verbose logs; **Production** uses structured, optimized logs
5. **Log aggregation** tools are essential for production systems
6. **Best practices** ensure logs are useful, searchable, and secure

### Remember

- ✅ Use structured logging (JSON) in production
- ✅ Include context (request IDs, user IDs)
- ✅ Use appropriate log levels
- ✅ Don't log sensitive data
- ✅ Aggregate logs for analysis
- ✅ Monitor and alert on errors

---

## Further Reading

- [The Twelve-Factor App: Logs](https://12factor.net/logs)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/best-practices-for-structured-logging/)
- [Logging Levels Explained](https://www.loggly.com/ultimate-guide/python-logging-basics/)

---

*This guide is framework and language agnostic. The concepts apply to any programming language and framework.*

