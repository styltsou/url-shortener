# Frontend Code Review

**Date:** 2026-06-30  
**Scope:** Complete frontend codebase review (fresh)

---

## Executive Summary

The frontend uses modern patterns (TanStack Query, TanStack Router, shadcn/ui, Recharts) and is well-organized. However, there are several significant code quality issues including a 583-line monolithic component, duplicated patterns, error handling gaps, and accessibility concerns.

**Overall Assessment:** ⚠️ **Strong architecture but needs component refactoring and error hardening**

---

## 🔴 High Priority Issues

### 1. `link-details-card.tsx` is a 583-Line Monolith

**Location:** `src/components/link/link-details-card.tsx`  
**Severity:** HIGH

This component reimplements all functionality from 4 separate components that already exist:
- `destination-card.tsx` — destination URL display/copy
- `expiration-section.tsx` — expiration editing UI
- `shortcode-section.tsx` — shortcode display
- `tags-section.tsx` — tag management

It duplicates escape-key event listeners, copy-to-clipboard logic, tag conversion logic, and expiration validation.

**Fix:** Compose the individual section components instead of reimplementing inline.

---

### 2. Empty Catch Blocks Swallow Errors

**Locations:** Multiple files  
**Severity:** HIGH

```typescript
try {
  await mutation.mutateAsync(...);
} catch (error) {
  // Error is handled by the hook
}
```

Affected files: `url-card.tsx`, `link-actions.tsx`, `link-details-card.tsx`, `shortcode-section.tsx`, `tags-section.tsx`, `link-qrcode.tsx`

The comment claims errors are handled by the hook, but the catch block swallows the error silently. `mutateAsync` will throw, so the catch IS reached, but nothing is done with the error.

**Fix:** Remove the `try/catch` or handle the error locally with a toast.

---

### 3. `<a href>` Used for Client-Side Navigation

**Locations:** `src/components/dashboard/recent-links.tsx:28`, `src/routes/index.tsx:90`  

```typescript
<a href='/links'>Create your first link</a>
<a href={`/links/${link.shortcode}`}>
```

Using `<a href>` instead of TanStack Router's `<Link>` causes full-page navigation, losing client-side state and animations.

**Fix:** Use `<Link>` from `@tanstack/react-router`.

---

### 4. UTC/Local Timezone Mismatch in Expiration Validation

**Locations:** `new-url-form.tsx:76-78`, `link-details-card.tsx:142`, `expiration-section.tsx:70`  
**Severity:** HIGH

```typescript
const combinedDate = new Date(formState.expirationDate);
combinedDate.setUTCHours(hours, minutes, 0, 0);
if (combinedDate < new Date()) { ... }
```

The combined date uses UTC hours, but `new Date()` returns local time. On UTC+2, this causes off-by-hour validation errors.

**Fix:** Use consistent timezone handling throughout (either all UTC or all local).

---

### 5. `useEffect` with Unstable Callback Dependencies

**Location:** `src/components/link/link-details-card.tsx:289-317`  
**Severity:** HIGH

`handleCancelExpiration` and `handleCancelTags` are not wrapped in `useCallback` but appear in `useEffect` dependency arrays. The effect re-runs on every render, adding/removing event listeners continuously.

**Fix:** Wrap handlers in `useCallback`.

---

## 🟡 Medium Priority Issues

### 6. API Response Type Mismatch

**Location:** `src/lib/api-client.ts:3-15`, `src/hooks/use-links.ts:71-76`  
**Severity:** MEDIUM

`ApiSuccessResponse<T>` is `{data: T}`, but the actual API returns `SuccessResponse<T>` = `{data: T, pagination?: PaginationMeta}`. This requires double casts like `as unknown as SuccessResponse<Link[]>` in `use-links.ts`.

**Fix:** Align the API client response type with the actual server response shape.

---

### 7. No Global Error Boundary

**Location:** `src/main.tsx`  
**Severity:** MEDIUM

The app renders without a React Error Boundary. A runtime error in any component unmounts the entire UI tree.

**Fix:** Add an `<ErrorBoundary>` wrapping the router.

---

### 8. Copy-to-Clipboard Pattern Duplicated 4x

**Locations:** `url-card.tsx`, `link-header.tsx`, `link-details-card.tsx`, `destination-card.tsx`  
**Severity:** MEDIUM

The copy button + tooltip + state management pattern is duplicated verbatim in 4 places.

**Fix:** Extract into a reusable `CopyButton` component or `useCopyToClipboard` hook.

---

### 9. Expiration Validation Duplicated in 3 Places

**Locations:** `new-url-form.tsx:73-90`, `link-details-card.tsx:133-175`, `expiration-section.tsx:61-103`  
**Severity:** MEDIUM

The logic to combine date/time, validate future dates, and convert to ISO string is duplicated.

**Fix:** Extract to `lib/utils.ts` as a shared utility function.

---

### 10. Tag Conversion Logic Duplicated in 3 Places

**Locations:** `new-url-form.tsx:63-67`, `link-details-card.tsx:103-109`, `tags-section.tsx:39-45`  
**Severity:** MEDIUM

The pattern of mapping API tags (`{id, name, created_at, updated_at}`) to UI tags (`{id, name}`) is duplicated.

**Fix:** Create a helper function `toUiTags(tags: ApiTag[]): Tag[]`.

---

### 11. `useUpdateLink` Cache Strategy: Manual Update + `refetchQueries` is Redundant

**Location:** `src/hooks/use-links.ts:231-266`  
**Severity:** MEDIUM

On success, the handler manually updates ALL caches, then immediately calls `refetchQueries` which overwrites the manual data. This is redundant work.

**Fix:** Pick one strategy — either manual update OR refetch, not both.

---

### 12. Expiration Cannot Be Cleared Once Set

**Locations:** `link-details-card.tsx:134`, `expiration-section.tsx:62-65`  
**Severity:** MEDIUM

```typescript
if (!expirationDate) {
  setIsEditing(false);
  return;
}
```

If a user clears the date picker, saving silently cancels instead of removing the expiration.

**Fix:** Allow setting `expires_at` to null via the API to clear expiration.

---

### 13. `links-list-skeleton.tsx` is Dead Code

**Location:** `src/components/links/links-list-skeleton.tsx`  
**Severity:** MEDIUM

This component is not imported anywhere in the codebase but adds to bundle size.

**Fix:** Remove the file.

---

### 14. `LoadingState` Uses `min-h-screen` But Used Inline

**Location:** `loading-state.tsx:14`, used in `dashboard-chart.tsx:20`, `performance-chart-card.tsx:20`  
**Severity:** MEDIUM

`min-h-screen` causes inline loading states inside cards to expand to fill the viewport.

**Fix:** Make `min-h-screen` optional via prop, or use a variant specifically for inline usage.

---

### 15. `handleShorten` Creates Link Then Tags in Sequential Mutations

**Location:** `src/routes/links/index.tsx:71-91`  
**Severity:** MEDIUM

If tagging fails, the link already exists untagged with no rollback.

**Fix:** Use a single server endpoint that creates a link with tags atomically, or delete the link on tag failure.

---

### 16. `useDeleteLink` Uses Wrong Key for Invalidation

**Location:** `src/hooks/use-links.ts:285`  
**Severity:** MEDIUM

```typescript
queryClient.invalidateQueries({ queryKey: linkKeys.list() });
```

If the key factory is refactored so `list()` and `lists()` diverge, this will break. Use `linkKeys.lists()` for the explicit intent.

---

### 17. Environment Validation Throws at Module Scope

**Location:** `src/routes/__root.tsx:17-24`  
**Severity:** LOW

Module-level side effects prevent tree-shaking and cause cryptic import failures.

**Fix:** Move validation to app initialization in `main.tsx`.

---

### 18. `"use client"` Directive in Non-Next.js Project

**Location:** `src/components/sidebar/app-sidebar.tsx:1`  
**Severity:** LOW

This directive is for Next.js App Router. This is a Vite + TanStack Router project.

**Fix:** Remove the directive.

---

### 19. Accessible Name Gaps

**Locations:** Multiple icon-only buttons  
**Severity:** LOW

Remove-tag buttons on badges and some icon-only buttons lack `aria-label`. The "Show options" collapsible section lacks `aria-expanded`, `aria-controls`, and `aria-hidden`.

---

### 20. Unused Dependencies

**Location:** `package.json`  
**Severity:** LOW

`zod` (v4), `zustand` (v5), and `date-fns` (v4) are declared but not used. All state management is done via React Query + `useState`.

**Fix:** Remove unused dependencies.

---

## 🟢 Low Priority

### 21. Unimported Icons

**Locations:** `link-actions.tsx` imports `Edit2` but uses `Pencil`; `link-details-card.tsx` imports `Save` but doesn't use it

---

### 22. Quote Style Inconsistency

Some files use single quotes, others double quotes. Reflects inconsistent formatter configuration.

---

### 23. Skeleton Naming Inconsistency

`DashboardSkeleton`, `LinkPageSkeleton`, `LinksPageSkeleton`, `LinksListSkeleton` — inconsistent prefix/suffix patterns.

---

## 🟢 What's Working Well

- ✅ Modern stack (TanStack Query, TanStack Router, shadcn/ui, Recharts)
- ✅ Optimistic updates in mutations with rollback
- ✅ Loading, empty, error states on most pages
- ✅ Good component separation in `links/` directory
- ✅ Auth integration with Clerk
- ✅ Dark/light mode support
- ✅ Navigation blocking for unsaved changes
- ✅ Toast notifications via Sonner

---

## 📊 Summary

| Category | Status | Key Issues |
|----------|--------|------------|
| Architecture | ✅ Good | Component structure well-organized |
| Code Quality | ⚠️ Needs Refactoring | 583-line monolith, 4x duplicated patterns |
| Error Handling | ⚠️ Needs Attention | Empty catch blocks, no error boundary |
| Performance | ✅ Good | Minor issues with event listener churn |
| Accessibility | ⚠️ Needs Work | Missing aria-labels, no error boundary |
| Type Safety | ✅ Good | Minor API type mismatch |

---

## 🎯 Recommended Action Plan

### Immediate
1. Refactor `link-details-card.tsx` to compose section components
2. Fix empty catch blocks — either remove or add error handling
3. Replace `<a href>` with TanStack Router `<Link>`
4. Fix UTC/local timezone mismatch in expiration

### Short Term
5. Extract copy-to-clipboard into reusable hook/component
6. Extract expiration validation into shared utility
7. Add global error boundary
8. Remove dead code (`links-list-skeleton.tsx`)
9. Add `useCallback` to event handler functions

### Medium Term
10. Align API client types with server response shape
11. Fix expiration clear functionality
12. Standardize mutation error handling patterns
13. Remove unused dependencies
14. Add `aria-label` to icon-only buttons
