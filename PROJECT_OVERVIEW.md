# 🧩 URL Shortener — Project Overview

A production-ready, extensible URL shortener built in Go.
The project starts simple — a minimal, robust MVP — and can later expand into a complex, distributed system that demonstrates intermediate-to-advanced backend and system design skills.

---

## 🚀 MVP Scope

### 🎯 **Core Goal**

Implement a backend service that allows:

- Users to shorten long URLs into unique short codes.
- Anyone to access a short URL and be redirected to the original link.
- Authenticated users to manage (list, delete) their shortened links.

---

### 🧱 **Core Features**

| Feature             | Description                                                                                                                        |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| **Shorten URL**     | Authenticated users can submit a long URL and receive a unique short code (e.g., `short.ly/abc123`).                               |
| **Redirect**        | Visiting `/{code}` redirects to the original long URL.                                                                             |
| **User Management** | Integrated with Clerk authentication. Each link belongs to a specific user (identified by Clerk ID). No local user storage needed. |
| **List User Links** | Authenticated users can list and manage their own shortened links.                                                                 |
| **Basic Analytics** | Track total click count per link.                                                                                                  |
| **Error Handling**  | Handle invalid codes, expired links, and validation errors gracefully.                                                             |

---

## 🧰 Tech Stack

| Layer                | Tool / Library  | Purpose                                      |
| -------------------- | --------------- | -------------------------------------------- |
| **Language**         | Go              | Backend core logic                           |
| **Framework**        | Chi             | Lightweight HTTP router                      |
| **Database**         | PostgreSQL      | Core data storage                            |
| **ORM/DB Layer**     | `pgx` or `sqlc` | Type-safe query management                   |
| **Auth**             | Clerk           | User authentication and session management   |
| **Cache (later)**    | Redis           | Speed up redirects and analytics aggregation |
| **Containerization** | Docker          | Local development and deployment consistency |

---

## 🧩 Database Models

> **Note:** Following Clerk's best practices, we don't store a users table. Clerk IDs are stored directly in the links table. This is recommended when you don't need extra user metadata or user-to-user relationships.

---

### 1. **Link**

Core table for shortened links.

| Field          | Type                 | Notes                        |
| -------------- | -------------------- | ---------------------------- |
| `id`           | UUID                 | Primary key                  |
| `code`         | varchar(10)          | Unique short identifier      |
| `original_url` | text                 | The long destination URL     |
| `user_id`      | varchar(255)         | Clerk user ID (no FK needed) |
| `clicks`       | int                  | Total clicks                 |
| `created_at`   | timestamp            | Creation time                |
| `expires_at`   | timestamp (nullable) | Optional expiration date     |
| `updated_at`   | timestamp (nullable) | Last update time             |

---

### 2. **ClickEvent** _(optional in MVP, for future analytics)_

Stores each click for analytics and aggregation.

| Field        | Type          | Notes                   |
| ------------ | ------------- | ----------------------- |
| `id`         | serial        | Primary key             |
| `link_id`    | FK → links.id | Related Link            |
| `timestamp`  | timestamp     | When the click occurred |
| `ip_address` | text          | Request origin          |
| `user_agent` | text          | Browser info            |
| `referer`    | text          | Source of click         |

> For MVP, you can start with just a `clicks` counter in the `links` table.

---

## 🧭 API Overview

| Endpoint               | Method | Auth | Description                       |
| ---------------------- | ------ | ---- | --------------------------------- |
| `/{code}`              | GET    | ❌   | Redirect to original URL          |
| `/api/v1/links`        | GET    | ✅   | List all user's links             |
| `/api/v1/links`        | POST   | ✅   | Create a new short link           |
| `/api/v1/links/{code}` | GET    | ✅   | Get link details and stats        |
| `/api/v1/links/{code}` | PATCH  | ✅   | Update link (custom code, expiry) |
| `/api/v1/links/{code}` | DELETE | ✅   | Delete a link                     |

---

## 🧮 MVP Architecture

A simple **monolithic Go service**:

```
┌──────────────────────────────┐
│          API Layer           │
│    (Chi Router + Handlers)   │
└──────────────┬───────────────┘
               │
┌──────────────▼──────────────┐
│      Service Layer           │
│  (Business / domain logic)   │
└──────────────┬──────────────┘
               │
┌──────────────▼──────────────┐
│   Repository Layer           │
│  (PostgreSQL CRUD ops)       │
└──────────────┬──────────────┘
               │
┌──────────────▼──────────────┐
│ PostgreSQL Database          │
└──────────────────────────────┘
```

---

## 🔮 Future Expansion & System Design Evolution

Once the MVP is solid, you can evolve the project into a **real-world distributed system**.
Here’s a roadmap to make it more complex and educational:

---

### 1. **Caching Layer (Redis)**

- Cache short-code lookups (`code → URL`).
- Reduce database read pressure.
- Implement a TTL-based invalidation policy.

**Pattern:** Read-through cache
**Benefit:** Sub-millisecond redirects at scale.

---

### 2. **Event-Driven Analytics**

Move from incrementing `clicks` synchronously → to **asynchronous event processing**.

**Architecture:**

```
Client → Go API → Kafka (or RabbitMQ) → Analytics Consumer → ClickHouse
```

- Each redirect produces a `ClickEvent` published to a queue.
- A background consumer stores detailed events in **ClickHouse**.
- Enables rich, queryable analytics.

---

### 3. **Analytics Platform (ClickHouse)**

ClickHouse is ideal for large-scale, write-heavy analytics.

Possible metrics:

- Click counts per link
- Clicks by time period
- Clicks by country / referrer / device

You can later expose an API like:

```bash
GET /analytics/{code}?range=7d
```

that queries ClickHouse directly for insights.

---

### 4. **Rate Limiting and Abuse Protection**

- Prevent malicious mass URL submissions.
- Use Redis for request counting per user.
- Enforce per-minute/hour limits.

---

### 5. **Link Customization and Management**

- Custom aliases (`/my-brand-link`).
- Expiring links.
- Link tags or folders for organization.
- Web dashboard integration.

---

### 6. **Scalable Architecture**

When load grows:

- Split services: `api`, `analytics`.
- Use **message queues (Kafka / NATS)** between services.
- Deploy via Docker + Kubernetes.
- Add **Prometheus + Grafana** for metrics.

---

## 🧠 Long-Term Learning Value

Building this project helps you learn:

| Skill Area                  | Concepts Learned                                               |
| --------------------------- | -------------------------------------------------------------- |
| **Go backend fundamentals** | Routing, handlers, dependency injection, testing               |
| **Persistence**             | SQL schema design, migrations, indexing                        |
| **Auth integration**        | Clerk tokens, middleware, direct Clerk ID usage (no user sync) |
| **Caching**                 | Redis for high-speed lookups                                   |
| **Event-driven design**     | Kafka or NATS for analytics                                    |
| **Observability**           | Logging, metrics, tracing                                      |
| **System design**           | Scalability, separation of concerns, service boundaries        |

---

## ✅ Summary

| Stage       | Focus                      | Key Tech                       |
| ----------- | -------------------------- | ------------------------------ |
| **MVP**     | Shorten + Redirect + Auth  | Go, Chi, PostgreSQL, Clerk     |
| **Phase 2** | Caching + Analytics events | Redis, background jobs         |
| **Phase 3** | Full analytics platform    | ClickHouse, Kafka              |
| **Phase 4** | Scaling & observability    | Docker, Prometheus, Kubernetes |

---

### 🧭 Guiding Philosophy

Start small, design cleanly, and evolve the architecture _organically_ as new requirements emerge.
Each layer (auth, persistence, analytics, caching) can be developed independently without rewriting the core service.
