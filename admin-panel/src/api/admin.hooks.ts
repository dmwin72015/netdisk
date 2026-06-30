/* ============================================================
 * Admin Panel TanStack Query Hooks
 *
 * One useQuery hook per read function, one useMutation hook
 * per write function. Mutations invalidate their related query
 * keys on success.
 * ============================================================ */

import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';

import {
  createUser,
  deleteFile,
  deletePhysicalFile,
  deleteUser,
  deleteUserFile,
  fetchActivityLogActions,
  fetchActivityLogs,
  fetchDashboardStats,
  fetchFiles,
  fetchStorageStats,
  fetchSystemConfig,
  fetchUser,
  fetchUsers,
  queryCleanup,
  resetSystemConfig,
  restoreFile,
  searchUsers,
  updateStorageBase,
  updateSystemConfig,
  updateUserRole,
  type ActivityLogParams,
  type CleanupQueryInput,
  type CreateUserInput,
  type FileListParams,
  type UpdateSystemConfigInput,
  type UserListParams,
} from './admin';

// ─── Dashboard ─────────────────────────────────────────────────

export function useDashboardStats() {
  return useQuery({
    queryKey: ['admin', 'dashboard', 'stats'],
    queryFn: fetchDashboardStats,
  });
}

// ─── Users ─────────────────────────────────────────────────────

export function useUsers(params: UserListParams) {
  return useQuery({
    queryKey: ['admin', 'users', params],
    queryFn: () => fetchUsers(params),
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: ['admin', 'users', id],
    queryFn: () => fetchUser(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUserInput) => createUser(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateUserRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) => updateUserRole(id, role),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateStorageBase() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, baseBytes }: { id: string; baseBytes: number }) =>
      updateStorageBase(id, baseBytes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useSearchUsers(query: string) {
  return useQuery({
    queryKey: ['admin', 'users', 'search', query],
    queryFn: () => searchUsers(query),
    enabled: query.length > 0,
  });
}

// ─── Files ─────────────────────────────────────────────────────

export function useFiles(params: FileListParams) {
  return useQuery({
    queryKey: ['admin', 'files', params],
    queryFn: () => fetchFiles(params),
  });
}

export function useDeleteFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFile(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}

export function useRestoreFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => restoreFile(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}

// ─── Storage ───────────────────────────────────────────────────

export function useStorageStats() {
  return useQuery({
    queryKey: ['admin', 'storage', 'stats'],
    queryFn: fetchStorageStats,
  });
}

// ─── System Config ─────────────────────────────────────────────

export function useSystemConfig() {
  return useQuery({
    queryKey: ['admin', 'system', 'config'],
    queryFn: fetchSystemConfig,
  });
}

export function useUpdateSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (updates: UpdateSystemConfigInput) => updateSystemConfig(updates),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}

export function useResetSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key?: string) => resetSystemConfig(key),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}

// ─── Activity Logs ─────────────────────────────────────────────

export function useActivityLogs(params: ActivityLogParams) {
  return useQuery({
    queryKey: ['admin', 'activity-logs', params],
    queryFn: () => fetchActivityLogs(params),
  });
}

export function useActivityLogActions(lang?: string) {
  return useQuery({
    queryKey: ['admin', 'activity-logs', 'actions', lang],
    queryFn: () => fetchActivityLogActions(lang),
  });
}

// ─── Cleanup ───────────────────────────────────────────────────

export function useCleanupQuery() {
  return useMutation({
    mutationFn: (data: CleanupQueryInput) => queryCleanup(data),
  });
}

export function useDeleteUserFile() {
  return useMutation({
    mutationFn: (userFileId: string) => deleteUserFile(userFileId),
  });
}

export function useDeletePhysicalFile() {
  return useMutation({
    mutationFn: (physicalFileId: string) => deletePhysicalFile(physicalFileId),
  });
}