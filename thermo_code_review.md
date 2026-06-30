# Thermo-Nuclear Code Quality Review

**Date:** 2026-06-30
**Scope:** Full codebase audit (server + client)

---

## Overall: ⚠️ Solid architecture, but serious complexity debt on the client

The backend has been meaningfully cleaned up (good job on the last commit `57b2fff`). The frontend needs the most attention — complex optimistic update patterns, oversized components, and type-system workarounds are dragging maintainability down.

---

## 🔴 Critical

### 1. `useUpdateLink` optimistic update is dangerously over-engineered

**Location:** `client/src/hooks/use-links.ts:131-210`

This single mutation handler snapshots ALL cached list queries, manually patches every one, patches the detail query, then on success **patches again with server data AND refetches**. The `onError` rolls back every snapshot. That's ~80 lines of optimistic update logic for a simple `is_active` toggle.

**Code-judo move:** Delete the entire `onMutate`/`onError` optimistic update block. Just use `onSuccess` with `queryClient.invalidateQueries`. The flash of stale data is imperceptible for an `is_active` toggle. If you want it snappy, set `placeholderData` instead of manually sharding cache state.

### 2. `RecordClick` goroutine is unbounded

**Location:** `server/pkg/handlers/link.go:89`

```go
go h.LinkService.RecordClick(context.WithoutCancel(r.Context()), link.ID, ip, r.UserAgent(), r.Referer())
```

Every redirect spawns a goroutine with no cap. If ClickHouse gets slow (or goes down with the 3s timeout), goroutines pile up. Add a bounded worker pool or channel-based queue with a sane buffer.

### 3. `sanitizeDSN` is fragile and has dead code

**Location:** `server/pkg/server.go:28-42`

The `!found` branch after `strings.Cut` is dead (you already checked `strings.Contains`). More importantly, it breaks on:
- `postgres://user@host/db` (no password — no colon, no redaction)
- `redis://:password@host:6379` (empty username, starts with colon)

Redis URLs with passwords are logged unsanitized.

### 4. `useLinks` type system is lying

**Location:** `client/src/hooks/use-links.ts:70`

```typescript
const response = await apiClient.get<SuccessResponse<Link[]>>(url, token);
const successData = response as unknown as SuccessResponse<Link[]>;
```

The cast `as unknown as SuccessResponse<Link[]>` is a full type-safety bypass. Either `apiClient.get` returns the right shape or it doesn't. The null checks downstream (`!successData || !Array.isArray(successData.data)`) confirm the types don't match reality. Fix the API client generics.

---

## 🟡 High

### 5. `link-details-card.tsx` is 583 lines with 6+ responsibilities

**Location:** `client/src/components/link/link-details-card.tsx`

This component manages:
- Destination display + copy-to-clipboard
- Expiration date/time picker (with edit/save/cancel)
- Tag editing with combobox (with add/remove/create)
- Navigation blocking (via hook)
- Two separate ESC keydown listeners
- Two state-initialization effects

Extract the expiration editor and tag editor into their own components. This file is one feature away from spaghetti territory.

### 6. `use-block-navigation.tsx` has convoluted state/ref interplay

**Location:** `client/src/hooks/use-block-navigation.tsx`

Three pieces of state (`showDialog`, `currentBlocker`, `shouldGoBackRef`) interact in ways that are hard to trace. The `shouldGoBackRef` flag is set in `handlePopState` and read in `handleConfirm` — but `currentBlocker` is also set in `handlePopState` and read in `handleConfirm`. If a second popstate fires between setting and confirming, the closure captures stale references.

### 7. Two identical ESC keydown effects

**Location:** `client/src/components/link/link-details-card.tsx:289-317`

```typescript
useEffect(() => { ... handleCancelExpiration ... }, [isEditingExpiration, ...]);
useEffect(() => { ... handleCancelTags ... }, [isEditingTags, ...]);
```

These should be a single effect: if `isEditingExpiration || isEditingTags`, register one listener that dispatches to the right handler.

### 8. `mock-data.ts` bundles utility functions with fake data

**Location:** `client/src/lib/mock-data.ts`

`formatDate`, `formatDateTime`, `getTimePeriod` are imported across the app from a file called `mock-data`. If mock data is ever removed, the entire app breaks. Extract these to `lib/date-utils.ts` or similar.

### 9. ClickHouse connection uses nil-conditional checks instead of a null object

**Location:** `server/pkg/analytics/clickhouse.go`

Every public method starts with `if c.conn == nil { return ... }`. This is repeated 7+ times. Better: return a `nopClient` from `New` when URL is empty that implements the same interface with no-op methods.

### 10. `handleShorten` creates link then adds tags — non-atomic partial-update

**Location:** `client/src/routes/links/index.tsx:71-91`

```typescript
const createdLink = await createLink.mutateAsync({...});
if (tagIds && tagIds.length > 0) {
    await addTagsToLink.mutateAsync({ linkId: createdLink.id, tagIds });
}
```

If the second mutation fails, the user has an orphan link with no tags and no rollback.

---

## 🟠 Medium

### 11. `env.ts` and `constants.ts` overlap

**Location:** `client/src/lib/env.ts:38` vs `client/src/lib/constants.ts:8`

`env.ts` exports `getApiBaseUrl()`, `constants.ts` exports `API_BASE_URL = getApiBaseUrl()`. Pick one indirection layer.

### 12. `normalizeSlice` generic exists for one call site

**Location:** `server/pkg/handlers/handler_helpers.go:39`

Generics add compilation overhead for a single caller. A direct nil check at the call site would be simpler.

### 13. Error sentinels mix casing conventions

**Location:** `server/pkg/errors/errors.go`

`LinkNotFound`, `TagNotFound` (PascalCase) vs standard Go convention of lowercase `errLinkNotFound`. Consistency matters in a shared package.

### 14. Docker Compose ClickHouse image still pinned to `latest`

**Location:** `docker-compose.yml`

Not addressed in the latest commit. Pin to a specific version.

### 15. No `.dockerignore` in `server/` or `client/`

Not addressed.

### 16. ClickHouse table name `link4it.click_events` is hardcoded

**Location:** `server/pkg/analytics/clickhouse.go:81`

Should be configurable or derive from config.

---

## 🟢 What's Working Well

- Clean handler → service → database layering on backend
- Excellent last commit — fixed 10+ security/correctness issues in one pass
- The `runInTx` functional approach to transactions is clean (just make it a method, not a field)
- Sentinel errors with consistent HTTP mapping via `handleError`
- Graceful degradation for Redis and ClickHouse
- `parseUUIDParam` extraction eliminated 5 copy-pasted blocks
- `normalizeSlice` handles the `null` vs `[]` JSON issue
- Client uses react-query's cache invalidation pattern well overall
- Good separation of component files on the client

---

## 🎯 Top 5 Action Items

1. **Delete the `useUpdateLink` optimistic update complexity** — replace with `invalidateQueries` on success
2. **Split `link-details-card.tsx`** — extract expiration-editor and tag-editor subcomponents
3. **Bound the ClickHouse recording goroutines** — channel-based worker pool
4. **Fix `sanitizeDSN` for edge cases** — handle empty usernames, missing passwords
5. **Move utilities out of `mock-data.ts`** — break the dependency on a mock file
