/* ============================================================
 * Admin Panel API Layer
 *
 * Pure async functions + TypeScript types.
 * All endpoints are rooted at /api/v1/admin/.
 * Backend response format: { data: <T>, error?: string, errCode?: number }
 * The unified `request` helper (./request) extracts `.data`
 * and handles errors / auth / timeouts.
 * ============================================================ */

import { request } from './request';

// ════════════════════════════════════════════════════════════════
//  TYPES
// ════════════════════════════════════════════════════════════════

// ─── Users ─────────────────────────────────────────────────────

export type AdminUser = {
  id: string;
  slug: string;
  username: string;
  email: string;
  role: string;
  registerMethod: string;
  status: number;
  usedBytes: number;
  baseBytes: number;
  memberBonusBytes: number;
  packBytes: number;
  totalBytes: number;
  createdAt: number;
  profile?: {
    displayName: string;
    avatarUrl: string;
    bio: string;
  };
  oauthAccounts?: {
    provider: string;
    providerAccountId: string;
    createdAt: number;
  }[];
};

export type AdminUserList = {
  items: AdminUser[];
  total: number;
  limit: number;
  offset: number;
};

export type CreateUserInput = {
  username: string;
  email: string;
  password: string;
  role?: string;
  storageBase?: number;
};

export type UserListParams = {
  limit?: number;
  offset?: number;
  search?: string;
  role?: string;
  status?: number;
};

export type UpdateRoleInput = {
  role: string;
};

export type UpdateStorageBaseInput = {
  baseBytes: number;
};

// ─── Files ─────────────────────────────────────────────────────

export type AdminFile = {
  id: string;
  userId: string;
  slug: string;
  username: string;
  fileName: string;
  isDir: boolean;
  fileSize: number;
  mimeType: string;
  fileCategory: string;
  isTrashed: boolean;
  isStarred: boolean;
  createdAt: number;
  updatedAt: number;
};

export type AdminFileList = {
  items: AdminFile[];
  total: number;
  limit: number;
  offset: number;
};

export type FileListParams = {
  limit?: number;
  offset?: number;
  userId?: string;
  search?: string;
  fileCategory?: string;
  isTrashed?: boolean;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
};

// ─── Dashboard ─────────────────────────────────────────────────

export type AdminDashboardStats = {
  totalUsers: number;
  totalFiles: number;
  totalStorage: number;
  storageUsed: number;
  newTodayUsers: number;
  newTodayFiles: number;
  diskTotal: number;
  diskUsed: number;
  diskFree: number;
};

// ─── Storage ───────────────────────────────────────────────────

export type CategoryStat = {
  category: string;
  totalSize: number;
  fileCount: number;
};

// ─── System Config ─────────────────────────────────────────────

export type SystemConfigItem = {
  key: string;
  value: string;
  description?: string;
};

export type UpdateSystemConfigInput = Record<string, string>;

// ─── Activity Logs ─────────────────────────────────────────────

export type AdminActivityLog = {
  id: number;
  userId: number;
  username: string;
  action: string;
  actionLabel: string;
  resourceType: string;
  resourceName: string;
  ip: string;
  ipRegion: string;
  userAgent: string;
  os: string;
  browser: string;
  extra: unknown;
  createdAt: string; // ISO string like "2024-01-01T00:00:00Z"
};

export type AdminActivityLogList = {
  items: AdminActivityLog[];
  total: number;
  limit: number;
  offset: number;
};

export type ActivityLogParams = {
  limit?: number;
  offset?: number;
  user_id?: number;
  action?: string;
  ip?: string;
  created_from?: string;
  created_to?: string;
  lang?: string;
};

export type AdminActionLabel = {
  action: string;
  label: string;
};

// ─── Cleanup ───────────────────────────────────────────────────

export type CleanupQueryPhysicalFile = {
  id: number;
  fileHash: string;
  fileSize: number;
  storagePath: string;
  mimeType: string;
  fileExists: boolean;
};

export type CleanupQueryUserFile = {
  id: number;
  slug: string;
  userId: number;
  username: string;
  fileName: string;
  fileSize: number;
  createdAt: number;
};

export type CleanupQueryResult = {
  physicalFile: CleanupQueryPhysicalFile | null;
  userFiles: CleanupQueryUserFile[];
  totalUploads: number;
  uniqueUsers: number;
};

export type CleanupQueryInput = {
  slug?: string;
  hash?: string;
};

export type DeleteActionResult = {
  deleted: boolean;
  message: string;
};

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Dashboard
// ════════════════════════════════════════════════════════════════

export async function fetchDashboardStats(): Promise<AdminDashboardStats> {
  return request.get<AdminDashboardStats>('/api/v1/admin/dashboard/stats');
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Users
// ════════════════════════════════════════════════════════════════

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

export async function deleteUser(id: string): Promise<void> {
  await request.delete<void>(`/api/v1/admin/users/${id}`);
}

export async function searchUsers(query: string): Promise<{ id: string; username: string }[]> {
  return request.get<{ id: string; username: string }[]>('/api/v1/admin/users/search', {
    params: { q: query },
  });
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Files
// ════════════════════════════════════════════════════════════════

export async function fetchFiles(params?: FileListParams): Promise<AdminFileList> {
  return request.get<AdminFileList>('/api/v1/admin/files', { params });
}

export async function deleteFile(id: string): Promise<void> {
  await request.delete<void>(`/api/v1/admin/files/${id}`);
}

export async function restoreFile(id: string): Promise<void> {
  await request.patch<void>(`/api/v1/admin/files/${id}/restore`);
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Storage
// ════════════════════════════════════════════════════════════════

export async function fetchStorageStats(): Promise<CategoryStat[]> {
  return request.get<CategoryStat[]>('/api/v1/admin/storage/stats');
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – System Config
// ════════════════════════════════════════════════════════════════

export async function fetchSystemConfig(): Promise<SystemConfigItem[]> {
  return request.get<SystemConfigItem[]>('/api/v1/admin/system/config');
}

export async function updateSystemConfig(updates: UpdateSystemConfigInput): Promise<SystemConfigItem[]> {
  return request.put<SystemConfigItem[]>('/api/v1/admin/system/config', updates);
}

export async function resetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
  return request.post<SystemConfigItem[]>('/api/v1/admin/system/config/reset', key ? { key } : {});
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Activity Logs
// ════════════════════════════════════════════════════════════════

export async function fetchActivityLogs(params?: ActivityLogParams): Promise<AdminActivityLogList> {
  return request.get<AdminActivityLogList>('/api/v1/admin/activity-logs', { params });
}

export async function fetchActivityLogActions(lang?: string): Promise<AdminActionLabel[]> {
  return request.get<AdminActionLabel[]>('/api/v1/admin/activity-logs/actions', {
    params: { lang },
  });
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Cleanup
// ════════════════════════════════════════════════════════════════

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
