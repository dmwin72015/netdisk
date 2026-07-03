import { request } from '@/utils/request';

// Re-export all types
export type {
  AdminUser,
  AdminUserList,
  CreateUserInput,
  UserListParams,
  UpdateRoleInput,
  UpdateStorageBaseInput,
  AdminFile,
  AdminFileList,
  FileListParams,
  AdminPhysicalFile,
  AdminPhysicalFileList,
  PhysicalFileListParams,
  PhysicalFileDetailResult,
  AdminDashboardStats,
  CategoryStat,
  SystemConfigItem,
  UpdateSystemConfigInput,
  AdminActivityLog,
  AdminActivityLogList,
  ActivityLogParams,
  AdminActionLabel,
  CleanupQueryPhysicalFile,
  CleanupQueryUserFile,
  CleanupQueryResult,
  CleanupQueryInput,
  DeleteActionResult,
} from './types';

import type {
  AdminUser,
  AdminUserList,
  CreateUserInput,
  UserListParams,
  UpdateRoleInput,
  UpdateStorageBaseInput,
  AdminFileList,
  FileListParams,
  AdminPhysicalFileList,
  PhysicalFileListParams,
  PhysicalFileDetailResult,
  AdminDashboardStats,
  CategoryStat,
  SystemConfigItem,
  UpdateSystemConfigInput,
  AdminActivityLogList,
  ActivityLogParams,
  AdminActionLabel,
  CleanupQueryInput,
  CleanupQueryResult,
  DeleteActionResult,
} from './types';

// ---- Auth ----

export interface LoginResponse {
  user: {
    role: string;
    [key: string]: unknown;
  };
  tokens: {
    accessToken: string;
  };
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  return request.post<LoginResponse>('/api/v1/auth/login', { email, password });
}

// ---- Dashboard ----

export async function fetchDashboardStats(): Promise<AdminDashboardStats> {
  return request.get<AdminDashboardStats>('/api/v1/admin/dashboard/stats');
}

// ---- Users ----

export async function fetchUsers(params?: UserListParams): Promise<AdminUserList> {
  return request.get<AdminUserList>('/api/v1/admin/users', { params });
}

export async function fetchUser(id: string): Promise<AdminUser> {
  return request.get<AdminUser>(`/api/v1/admin/users/${id}`);
}

export async function createUser(data: CreateUserInput): Promise<AdminUser> {
  return request.post<AdminUser>('/api/v1/admin/users', data);
}

export async function updateUserRole(id: string, role: string): Promise<AdminUser> {
  return request.patch<AdminUser>(`/api/v1/admin/users/${id}`, { role } satisfies UpdateRoleInput);
}

export async function updateStorageBase(id: string, baseBytes: number): Promise<AdminUser> {
  return request.patch<AdminUser>(
    `/api/v1/admin/users/${id}/storage-base`,
    { baseBytes } satisfies UpdateStorageBaseInput,
  );
}

export async function deleteUser(id: string): Promise<DeleteActionResult> {
  return request.delete<DeleteActionResult>(`/api/v1/admin/users/${id}`);
}

export async function searchUsers(query: string): Promise<AdminUser[]> {
  return request.get<AdminUser[]>('/api/v1/admin/users/search', {
    params: { q: query },
  });
}

// ---- Files ----

export async function fetchFiles(params?: FileListParams): Promise<AdminFileList> {
  return request.get<AdminFileList>('/api/v1/admin/files', { params });
}

export async function deleteFile(id: string): Promise<DeleteActionResult> {
  return request.delete<DeleteActionResult>(`/api/v1/admin/files/${id}`);
}

export async function restoreFile(id: string): Promise<DeleteActionResult> {
  return request.patch<DeleteActionResult>(`/api/v1/admin/files/${id}/restore`);
}

// ---- Physical Files ----

export async function fetchPhysicalFiles(params?: PhysicalFileListParams): Promise<AdminPhysicalFileList> {
  return request.get<AdminPhysicalFileList>('/api/v1/admin/files/physical', { params });
}

export async function fetchPhysicalFileDetail(id: string): Promise<PhysicalFileDetailResult> {
  return request.get<PhysicalFileDetailResult>(`/api/v1/admin/files/physical/${id}`);
}

// ---- Storage ----

export async function fetchStorageStats(): Promise<CategoryStat[]> {
  return request.get<CategoryStat[]>('/api/v1/admin/storage/stats');
}

// ---- System Config ----

export async function fetchSystemConfig(): Promise<SystemConfigItem[]> {
  return request.get<SystemConfigItem[]>('/api/v1/admin/system/config');
}

export async function updateSystemConfig(updates: UpdateSystemConfigInput): Promise<SystemConfigItem[]> {
  return request.put<SystemConfigItem[]>('/api/v1/admin/system/config', updates);
}

export async function resetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
  return request.post<SystemConfigItem[]>('/api/v1/admin/system/config/reset', key ? { key } : {});
}

// ---- Activity Logs ----

export async function fetchActivityLogs(params?: ActivityLogParams): Promise<AdminActivityLogList> {
  return request.get<AdminActivityLogList>('/api/v1/admin/activity-logs', { params });
}

export async function fetchActivityLogActions(lang?: string): Promise<AdminActionLabel[]> {
  return request.get<AdminActionLabel[]>('/api/v1/admin/activity-logs/actions', {
    params: { lang },
  });
}

// ---- Cleanup ----

export async function queryCleanup(data: CleanupQueryInput): Promise<CleanupQueryResult> {
  return request.post<CleanupQueryResult>('/api/v1/admin/cleanup/query', data);
}

export async function deleteUserFile(userFileId: number): Promise<DeleteActionResult> {
  return request.post<DeleteActionResult>('/api/v1/admin/cleanup/delete-user-file', { userFileId });
}

export async function deletePhysicalFile(physicalFileId: number): Promise<DeleteActionResult> {
  return request.post<DeleteActionResult>('/api/v1/admin/cleanup/delete-physical-file', {
    physicalFileId,
  });
}
