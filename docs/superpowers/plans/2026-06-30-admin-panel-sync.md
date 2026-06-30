# Admin Panel Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Synchronize all 8 admin features from SvelteKit frontend to React + Ant Design admin-panel, using TanStack Query for data fetching.

**Architecture:** Two-layer API architecture: pure fetch functions in `api/admin.ts` + TanStack Query hooks in `api/admin.hooks.ts`. Pages consume hooks only, never raw fetch. Query invalidation ensures cross-page cache consistency.

**Tech Stack:** React 19, Ant Design 6, TanStack Query (React Query) 5+, TypeScript 6, react-router-dom 7

## Global Constraints

- All API functions go in `api/admin.ts`; all hooks go in `api/admin.hooks.ts`
- Pages must NOT import from `admin.ts` directly — only from `admin.hooks.ts`
- Use Ant Design components for all UI (Table, Form, Modal, Select, etc.)
- All dates stored as epoch seconds → convert with `new Date(epoch * 1000).toLocaleString()`
- All file sizes stored as bytes → convert with `formatBytes()`
- Forms must validate before submit; show `message.error()` on API errors
- Commit after each task completes (`git add` + `git commit`)

---

### Task 1: Install TanStack Query + Set Up Infrastructure

**Files:**
- Modify: `admin-panel/package.json`
- Modify: `admin-panel/src/main.tsx`
- Create: `admin-panel/src/api/admin.ts` (rewrite)
- Create: `admin-panel/src/api/admin.hooks.ts`

**Interfaces:**
- Consumes: existing react-router-dom v7 setup
- Produces: `QueryClientProvider` wrapper, all API functions and hooks

- [ ] **Step 1: Install TanStack Query**

```bash
cd admin-panel
pnpm add @tanstack/react-query
```

- [ ] **Step 2: Set up QueryClientProvider in main.tsx**

```typescript
// admin-panel/src/main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

// <QueryClientProvider client={queryClient}> wraps <App />
```

- [ ] **Step 3: Write pure API functions and types in api/admin.ts**

All admin API endpoints as pure async functions with full TypeScript types matching the backend response shapes. See the design doc for the complete list of types (`AdminUser`, `AdminFile`, `AdminDashboardStats`, `SystemConfigItem`, `AdminActivityLog`, `CleanupQueryResult`, etc.) and functions (`fetchDashboardStats`, `fetchUsers`, `fetchFiles`, `fetchStorageStats`, `fetchSystemConfig`, `fetchActivityLogs`, `queryCleanup`, etc.).

- [ ] **Step 4: Write TanStack Query hooks in api/admin.hooks.ts**

One hook per API function using `useQuery` for reads, `useMutation` for writes. Query keys follow `['admin', resource, params]` pattern. Mutations invalidate the corresponding query key on success.

- [ ] **Step 5: Verify build**

```bash
cd admin-panel
npx tsc --noEmit
```

- [ ] **Step 6: Commit**

```bash
git add admin-panel/
git commit -m "feat(admin): install TanStack Query, rewrite API layer with hooks"
```

---

### Task 2: Rewrite Existing Pages to Use Hooks + Fix Bugs

**Files:**
- Modify: `admin-panel/src/pages/Dashboard.tsx`
- Modify: `admin-panel/src/pages/Users.tsx`
- Modify: `admin-panel/src/pages/UserDetail.tsx`
- Modify: `admin-panel/src/pages/Files.tsx`
- Modify: `admin-panel/src/pages/Storage.tsx`
- Modify: `admin-panel/src/pages/Settings.tsx`

**Interfaces:**
- Consumes: all hooks from Task 1
- Produces: bug-free, hook-driven versions of all 6 existing pages

**Bugs to fix:**
- Users page: username link navigates to `/users/${id}` instead of `/admin/users/${id}`
- Users/Files pages: search Enter handler may use stale `page` state (React batching)
- Users page: storage base edit modal exists but has no trigger button in the UI
- Storage page: contains duplicate settings form that also exists in Settings.tsx
- Settings page: uses non-existent `/admin/settings` API instead of `/admin/system/config`
- Dashboard: silent error handling (no user-facing error message on API failure)

- [ ] **Step 1: Rewrite Dashboard.tsx**

Use `useDashboardStats()`. Handle loading (`<Spin>`), error (`<Result status="error">`), and empty states. 6 stat cards + storage progress bar + disk usage bar with color thresholds.

- [ ] **Step 2: Rewrite Users.tsx**

Use `useUsers()`, `useCreateUser()`, `useUpdateUserRole()`, `useUpdateStorageBase()`, `useDeleteUser()`. Paginated table with search, role filter, sort. Create/Edit/Delete dialogs. Fix nav link to `/admin/users/${id}`. React Query handles re-fetch on param change — no stale page state.

- [ ] **Step 3: Rewrite UserDetail.tsx**

Use `useUser(id)`. Basic Info, Storage, Profile, OAuth cards. Back link to `/admin/users`.

- [ ] **Step 4: Rewrite Files.tsx**

Use `useFiles()`, `useDeleteFile()`, `useRestoreFile()`. Search, category filter, trashed filter, sort. Per-file Restore + Permanent Delete buttons. Owner links to user detail.

- [ ] **Step 5: Rewrite Storage.tsx**

Use `useStorageStats()`. Remove the settings form. Only show category breakdown with progress bars and trash stats.

- [ ] **Step 6: Rewrite Settings.tsx**

Use `useSystemConfig()`, `useUpdateSystemConfig()`, `useResetSystemConfig()`. Table of config items with typed edit dialogs (bytes/number/string/bool) and reset buttons.

- [ ] **Step 7: Verify build**

```bash
cd admin-panel && npx tsc --noEmit
```

- [ ] **Step 8: Commit**

```bash
git commit -m "feat(admin): rewrite all existing pages with TanStack Query hooks, fix bugs"
```

---

### Task 3: Create Activity Logs Page (New)

**Files:**
- Create: `admin-panel/src/pages/ActivityLogs.tsx`
- Modify: `admin-panel/src/App.tsx`
- Modify: `admin-panel/src/components/AdminLayout.tsx`

**Interfaces:**
- Consumes: `useActivityLogs()`, `useActivityLogActions()`
- Produces: fully functional activity logs viewer

- [ ] **Step 1: Create ActivityLogs.tsx**

Paginated table: ID, User (link to detail), Action, Resource, IP, OS, Browser, Time, Detail button. Filters: user ID, action type dropdown (from server), IP, date range. Detail Modal with full log entry.

- [ ] **Step 2: Add route in App.tsx**

```typescript
<Route path="/admin/logs" element={<ActivityLogs />} />
```

- [ ] **Step 3: Add nav item in AdminLayout.tsx**

`{ key: '/admin/logs', icon: <ScrollTextOutlined />, label: 'Activity Logs' }`

- [ ] **Step 4: Verify build & commit**

---

### Task 4: Create File Cleanup Page (New)

**Files:**
- Create: `admin-panel/src/pages/Cleanup.tsx`
- Modify: `admin-panel/src/App.tsx`

**Interfaces:**
- Consumes: `useCleanupQuery()`, `useDeleteUserFile()`, `useDeletePhysicalFile()`
- Produces: orphan detection and cleanup tool

- [ ] **Step 1: Create Cleanup.tsx**

Two query modes (slug/hash tabs). Search history in localStorage. Summary cards. Physical File card with existence check. User Files table with delete buttons. Delete All confirmation.

- [ ] **Step 2: Add route in App.tsx**

```typescript
<Route path="/admin/cleanup" element={<Cleanup />} />
```

- [ ] **Step 3: Add nav item in AdminLayout.tsx**

`{ key: '/admin/cleanup', icon: <ToolOutlined />, label: 'Cleanup' }`

- [ ] **Step 4: Verify build & commit**

---

### Task 5: Final Integration & Build Verification

**Files:**
- All admin-panel source files

- [ ] **Step 1: Full TypeScript check**

```bash
cd admin-panel && npx tsc --noEmit
```

- [ ] **Step 2: Full Vite build**

```bash
cd admin-panel && npx vite build
```

- [ ] **Step 3: Verify all routes work by checking App.tsx**

Routes should include: `/admin` (Dashboard), `/admin/users` (Users), `/admin/users/:id` (UserDetail), `/admin/files` (Files), `/admin/storage` (Storage), `/admin/settings` (Settings), `/admin/logs` (ActivityLogs), `/admin/cleanup` (Cleanup).

- [ ] **Step 4: Final commit**

```bash
git add admin-panel/ && git commit -m "chore(admin): final integration and build verification"
```