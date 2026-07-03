import { message } from 'antd';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListQueryResult } from '@/types/query';
import {
  createUser,
  deleteFile,
  deletePhysicalFile,
  deleteUser,
  deleteUserFile,
  fetchActivityLogActions,
  fetchDashboardStats,
  fetchFiles,
  fetchPhysicalFileDetail,
  fetchPhysicalFiles,
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
} from './admin';
import type {
  AdminFile,
  AdminPhysicalFile,
  AdminUser,
  CleanupQueryInput,
  CreateUserInput,
  FileListParams,
  PhysicalFileListParams,
  UpdateSystemConfigInput,
  UserListParams,
} from './admin';

// --- Dashboard ---
export function useDashboardStats() {
  return useQuery({
    queryKey: ['admin', 'dashboard', 'stats'],
    queryFn: fetchDashboardStats,
  });
}

// --- Users ---
export function useUsers(params: UserListParams): ListQueryResult<AdminUser> {
  const query = useQuery({
    queryKey: ['admin', 'users', params],
    queryFn: () => fetchUsers(params),
  });
  return {
    data: query.data?.items ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error as Error | null,
    refetch: () => query.refetch(),
  };
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
      message.success('创建成功');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateUserRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) => updateUserRole(id, role),
    onSuccess: () => {
      message.success('角色已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateStorageBase() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, baseBytes }: { id: string; baseBytes: number }) => updateStorageBase(id, baseBytes),
    onSuccess: () => {
      message.success('存储配额已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      message.success('用户已删除');
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

// --- Files ---
export function useFiles(params: FileListParams): ListQueryResult<AdminFile> {
  const query = useQuery({
    queryKey: ['admin', 'files', params],
    queryFn: () => fetchFiles(params),
  });
  return {
    data: query.data?.items ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error as Error | null,
    refetch: () => query.refetch(),
  };
}

export function useDeleteFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteFile(id),
    onSuccess: () => {
      message.success('文件已删除');
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}

export function useRestoreFile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => restoreFile(id),
    onSuccess: () => {
      message.success('文件已恢复');
      qc.invalidateQueries({ queryKey: ['admin', 'files'] });
    },
  });
}

// --- Physical Files ---
export function usePhysicalFiles(params: PhysicalFileListParams): ListQueryResult<AdminPhysicalFile> {
  const query = useQuery({
    queryKey: ['admin', 'physical-files', params],
    queryFn: () => fetchPhysicalFiles(params),
  });
  return {
    data: query.data?.items ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    isError: query.isError,
    error: query.error as Error | null,
    refetch: () => query.refetch(),
  };
}

export function usePhysicalFileDetail(id: string) {
  return useQuery({
    queryKey: ['admin', 'physical-file-detail', id],
    queryFn: () => fetchPhysicalFileDetail(id),
    enabled: !!id,
  });
}

// --- Storage ---
export function useStorageStats() {
  return useQuery({
    queryKey: ['admin', 'storage', 'stats'],
    queryFn: fetchStorageStats,
  });
}

// --- System Config ---
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
      message.success('配置已更新');
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}

export function useResetSystemConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key?: string) => resetSystemConfig(key),
    onSuccess: () => {
      message.success('配置已重置');
      qc.invalidateQueries({ queryKey: ['admin', 'system', 'config'] });
    },
  });
}

// --- Activity Logs ---
export function useActivityLogActions(lang?: string) {
  return useQuery({
    queryKey: ['admin', 'activity-logs', 'actions', lang],
    queryFn: () => fetchActivityLogActions(lang),
  });
}

// --- Cleanup ---
export function useCleanupQuery() {
  return useMutation({
    mutationFn: (data: CleanupQueryInput) => queryCleanup(data),
  });
}

export function useDeleteUserFile() {
  return useMutation({
    mutationFn: (userFileId: number) => deleteUserFile(userFileId),
  });
}

export function useDeletePhysicalFile() {
  return useMutation({
    mutationFn: (physicalFileId: number) => deletePhysicalFile(physicalFileId),
  });
}
