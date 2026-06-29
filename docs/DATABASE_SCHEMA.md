# Database Schema Reference

This document provides a complete reference for the database schema. It should be kept up to date with migrations.

**Last Updated:** Based on migrations 000001-000004

---

## Table of Contents

1. [Overview](#overview)
2. [Tables](#tables)
   - [links](#links)
   - [tags](#tags)
   - [link_tags](#link_tags)
3. [Relationships](#relationships)
4. [Indexes](#indexes)
5. [Query Patterns](#query-patterns)

---

## Overview

The database uses:
- **PostgreSQL** as the database engine
- **UUID** for primary keys (generated with `gen_random_uuid()`)
- **Soft deletes** (`deleted_at` column) for data retention
- **Timestamps** (`created_at`, `updated_at`) for audit trails

---

## Tables

### links

Stores shortened URL links created by users.

**Columns:**

| Column | Type | Constraints | Default | Description |
|--------|------|-------------|---------|-------------|
| `id` | UUID | PRIMARY KEY | `gen_random_uuid()` | Unique identifier |
| `shortcode` | VARCHAR(20) | NOT NULL | - | Short code for the URL (e.g., "abc123") |
| `original_url` | TEXT | NOT NULL | - | The original long URL |
| `user_id` | TEXT | NOT NULL | - | ID of the user who created the link |
| `expires_at` | TIMESTAMP | - | `NULL` | Optional expiration date/time |
| `is_active` | BOOLEAN | NOT NULL | `true` | Whether the link is active |
| `created_at` | TIMESTAMP | NOT NULL | `NOW()` | When the link was created |
| `updated_at` | TIMESTAMP | - | `NULL` | When the link was last updated |
| `deleted_at` | TIMESTAMP | - | `NULL` | Soft delete timestamp (NULL = not deleted) |

**Indexes:**
- `idx_links_shortcode` - Partial unique index on `shortcode` WHERE `deleted_at IS NULL`
- `idx_links_user_id` - Index on `user_id` for efficient user queries
- `idx_links_deleted_at` - Partial index on `deleted_at` WHERE `deleted_at IS NOT NULL`
- `idx_links_is_active` - Partial index on `is_active` WHERE `is_active = true`

**Notes:**
- `shortcode` must be unique among non-deleted links (allows reuse after deletion)
- `deleted_at` is used for soft deletes - never returned in API responses
- `is_active` allows users to temporarily disable links without deleting them

---

### tags

Stores user-created tags for categorizing links.

**Columns:**

| Column | Type | Constraints | Default | Description |
|--------|------|-------------|---------|-------------|
| `id` | UUID | PRIMARY KEY | `gen_random_uuid()` | Unique identifier |
| `name` | VARCHAR(30) | NOT NULL | - | Tag name (max 30 characters) |
| `user_id` | TEXT | NOT NULL | - | ID of the user who created the tag |
| `created_at` | TIMESTAMP | NOT NULL | `NOW()` | When the tag was created |
| `updated_at` | TIMESTAMP | - | `NULL` | When the tag was last updated |

**Indexes:**
- `index_tags_user_id_name` - Unique index on `(user_id, name)`

**Notes:**
- Tag names must be unique per user
- Uses hard delete (no `deleted_at` column) - deleted tags are permanently removed
- When a tag is deleted, all `link_tags` relationships are automatically removed via CASCADE
- Maximum tag name length is 30 characters

---

### link_tags

Junction table for the many-to-many relationship between links and tags.

**Columns:**

| Column | Type | Constraints | Default | Description |
|--------|------|-------------|---------|-------------|
| `link_id` | UUID | NOT NULL, FK → links.id | - | Reference to links table |
| `tag_id` | UUID | NOT NULL, FK → tags.id | - | Reference to tags table |

**Primary Key:**
- Composite primary key: `(link_id, tag_id)`

**Foreign Keys:**
- `link_id` → `links(id)` ON DELETE CASCADE
- `tag_id` → `tags(id)` ON DELETE CASCADE

**Indexes:**
- `idx_link_tags_link_id` - Index on `link_id` for "get all tags for a link" queries
- `idx_link_tags_tag_id` - Index on `tag_id` for "get all links with a tag" queries

**Notes:**
- No timestamps (not needed for junction tables)
- No soft deletes (relationships are removed when link or tag is deleted via CASCADE)
- Composite primary key prevents duplicate tag assignments to the same link
- CASCADE deletion: if a link is deleted, all its tag relationships are automatically removed

---

## Relationships

### Entity Relationship Diagram

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│    links    │         │  link_tags   │         │    tags     │
├─────────────┤         ├──────────────┤         ├─────────────┤
│ id (PK)     │◄──┐     │ link_id (FK) │     ┌──►│ id (PK)     │
│ shortcode   │   │     │ tag_id (FK)  │     │   │ name        │
│ original_url│   │     └──────────────┘     │   │ user_id     │
│ user_id     │   │                           │   │ created_at  │
│ expires_at  │   │                           │   │ updated_at  │
│ is_active   │   │                           │   └─────────────┘
│ created_at  │   │                           │
│ updated_at  │   │                           │
│ deleted_at  │   │                           │
└─────────────┘   │                           │
                  │                           │
                  └───────────────────────────┘
```

**Relationship Types:**
- **links ↔ tags**: Many-to-many (via `link_tags` junction table)
  - One link can have many tags
  - One tag can be assigned to many links

---

## Indexes

### Summary

| Table | Index Name | Columns | Type | Partial? | Purpose |
|-------|------------|---------|------|----------|---------|
| `links` | `idx_links_shortcode` | `shortcode` | UNIQUE | Yes (`deleted_at IS NULL`) | Enforce unique shortcodes for active links |
| `links` | `idx_links_user_id` | `user_id` | Regular | No | Speed up user queries |
| `links` | `idx_links_deleted_at` | `deleted_at` | Regular | Yes (`deleted_at IS NOT NULL`) | Speed up cleanup queries |
| `links` | `idx_links_is_active` | `is_active` | Regular | Yes (`is_active = true`) | Speed up active link queries |
| `tags` | `index_tags_user_id_name` | `(user_id, name)` | UNIQUE | No | Enforce unique tag names per user |
| `link_tags` | `idx_link_tags_link_id` | `link_id` | Regular | No | Speed up "get tags for link" queries |
| `link_tags` | `idx_link_tags_tag_id` | `tag_id` | Regular | No | Speed up "get links for tag" queries |

### Partial Indexes Explained

Partial indexes only index rows that match a condition, making them smaller and faster:

- **`idx_links_shortcode`**: Only indexes non-deleted links, allowing shortcode reuse after deletion
- **`idx_links_deleted_at`**: Only indexes deleted links, useful for cleanup operations
- **`idx_links_is_active`**: Only indexes active links, speeding up queries for active links only
- **`index_tags_user_id_name`**: Only indexes non-deleted tags, allowing tag name reuse after deletion

---

## Query Patterns

### Common Query Patterns

#### Get Link with Tags

```sql
SELECT 
    l.*,
    COALESCE(
        json_agg(
            json_build_object('id', t.id, 'name', t.name, 'created_at', t.created_at)
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'::json
    ) as tags
FROM links l
LEFT JOIN link_tags lt ON l.id = lt.link_id
LEFT JOIN tags t ON lt.tag_id = t.id
WHERE l.shortcode = $1 AND l.user_id = $2 AND l.deleted_at IS NULL
GROUP BY l.id;
```

#### List User Links with Tags

```sql
SELECT 
    l.*,
    COALESCE(
        json_agg(
            json_build_object('id', t.id, 'name', t.name, 'created_at', t.created_at)
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'::json
    ) as tags
FROM links l
LEFT JOIN link_tags lt ON l.id = lt.link_id
LEFT JOIN tags t ON lt.tag_id = t.id
WHERE l.user_id = $1 AND l.deleted_at IS NULL
GROUP BY l.id
ORDER BY l.created_at DESC;
```

#### Filter Links by Tag

```sql
SELECT DISTINCT l.*
FROM links l
INNER JOIN link_tags lt ON l.id = lt.link_id
INNER JOIN tags t ON lt.tag_id = t.id
WHERE l.user_id = $1 
  AND l.deleted_at IS NULL
  AND t.name = $2
ORDER BY l.created_at DESC;
```

#### Get All Tags for a User

```sql
SELECT id, name, created_at
FROM tags
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name;
```

---

## Important Notes

### Soft Deletes

- **Links table** uses `deleted_at` for soft deletes
- **Tags table** uses hard delete (no `deleted_at` column) - tags are permanently removed when deleted
- **Never return `deleted_at` in API responses** (security/privacy)
- Always filter links with `WHERE deleted_at IS NULL` in queries
- Partial unique indexes on links allow reuse of shortcodes after deletion

### Foreign Key Behavior

- `link_tags` uses `ON DELETE CASCADE`:
  - Deleting a link automatically removes all its tag relationships
  - Deleting a tag automatically removes it from all links

### Column Visibility

- `deleted_at` is **never returned** in SELECT queries (excluded from column lists)
- This is enforced at the database query level, not in application code
- See `server/queries/links.sql` for examples

### Timestamps

- `created_at`: Set automatically on INSERT (default: `NOW()`)
- `updated_at`: Must be set manually on UPDATE (typically `NOW()`)
- `deleted_at`: Set to `NOW()` when soft deleting

---

## Migration History

| Migration | Description |
|-----------|-------------|
| `000001` | Create `links` table |
| `000002` | Add `is_active` column to `links` |
| `000003` | Create `tags` table |
| `000004` | Create `link_tags` junction table |
| `000005` | Remove `clicks` column from `links` table |

---

## Future Considerations

When adding new tables or columns:

1. **Always include**:
   - `id` (UUID PRIMARY KEY)
   - `created_at` (TIMESTAMP NOT NULL DEFAULT NOW())
   - `updated_at` (TIMESTAMP)
   - `deleted_at` (TIMESTAMP) for soft deletes

2. **Consider indexes for**:
   - Foreign keys
   - Columns used in WHERE clauses frequently
   - Unique constraints (use partial indexes with soft deletes)

3. **Update this document** when schema changes

