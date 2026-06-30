/* ============================================================
 * Admin Panel API Layer
 *
 * Pure async fetch functions + TypeScript types.
 * All endpoints are rooted at /api/v1/admin/.
 * Backend response format: { data: <T>, error?: string, errCode?: number }
 * The request<T>() helper extracts `.data` and handles errors.
 * ============================================================ */

// ─── Token ─────────────────────────────────────────────────────

function getToken(): string | null {
  try {
    return localStorage.getItem('nd.access');
  } catch {
    return null;
  }
}

// ─── Generic request helper ────────────────────────────────────

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> | undefined),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(path, {
    ...options,
    headers,
  });

  if (!res.ok) {
    let errorMsg = `HTTP ${res.status}`;
    try {
      const err = await res.json();
      errorMsg = err.error || errorMsg;
    } catch {
      // ignore
    }
    throw new Error(errorMsg);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const json = await res.json();
  return json.data as T;
}

const BASE = '/api/v1/admin';

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
  const parts: string[] = [];
  for (const [key, val] of Object.entries(params)) {
    if (val !== undefined && val !== '') {
      parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(val))}`);
    }
  }
  return parts.length ? `?${parts.join('&')}` : '';
}

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
  bytes: number;
  count: number;
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
  id: string;
  fileId: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
  uploadAt: number;
  refCount: number;
};

export type CleanupQueryUserFile = {
  id: string;
  userId: string;
  username: string;
  fileName: string;
  fileSize: number;
  isTrashed: boolean;
  physicalFileId: string;
};

export type CleanupQueryResult = {
  userFiles: CleanupQueryUserFile[];
  physicalFiles: CleanupQueryPhysicalFile[];
  totalUserFiles: number;
  totalPhysicalFiles: number;
};

export type CleanupQueryInput = {
  orphanDays?: number;
  trashedDays?: number;
  minFileSize?: number;
  userId?: string;
};

export type DeleteActionResult = {
  success: boolean;
  message?: string;
};

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Dashboard
// ════════════════════════════════════════════════════════════════

export async function fetchDashboardStats(): Promise<AdminDashboardStats> {
  return request<AdminDashboardStats>(`${BASE}/dashboard/stats`);
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Users
// ════════════════════════════════════════════════════════════════

export async function fetchUsers(params?: UserListParams): Promise<AdminUserList> {
  const qs = params ? buildQuery(params as Record<string, string | number | boolean | undefined>) : '';
  return request<AdminUserList>(`${BASE}/users${qs}`);
}

export async function fetchUser(id: string): Promise<AdminUser> {
  return request<AdminUser>(`${BASE}/users/${id}`);
}

export async function createUser(data: CreateUserInput): Promise<AdminUser> {
  return request<AdminUser>(`${BASE}/users`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function updateUserRole(id: string, role: string): Promise<AdminUser> {
  return request<AdminUser>(`${BASE}/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ role } satisfies UpdateRoleInput),
  });
}

export async function updateStorageBase(id: string, baseBytes: number): Promise<AdminUser> {
  return request<AdminUser>(`${BASE}/users/${id}/storage-base`, {
    method: 'PATCH',
    body: JSON.stringify({ baseBytes } satisfies UpdateStorageBaseInput),
  });
}

export async function deleteUser(id: string): Promise<void> {
  await request<void>(`${BASE}/users/${id}`, { method: 'DELETE' });
}

export async function searchUsers(query: string): Promise<{ id: string; username: string }[]> {
  return request<{ id: string; username: string }[]>(
    `${BASE}/users/search?q=${encodeURIComponent(query)}`,
  );
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Files
// ════════════════════════════════════════════════════════════════

export async function fetchFiles(params?: FileListParams): Promise<AdminFileList> {
  const qs = params ? buildQuery(params as Record<string, string | number | boolean | undefined>) : '';
  return request<AdminFileList>(`${BASE}/files${qs}`);
}

export async function deleteFile(id: string): Promise<void> {
  await request<void>(`${BASE}/files/${id}`, { method: 'DELETE' });
}

export async function restoreFile(id: string): Promise<void> {
  await request<void>(`${BASE}/files/${id}/restore`, { method: 'PATCH' });
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Storage
// ════════════════════════════════════════════════════════════════

export async function fetchStorageStats(): Promise<CategoryStat[]> {
  return request<CategoryStat[]>(`${BASE}/storage/stats`);
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – System Config
// ════════════════════════════════════════════════════════════════

export async function fetchSystemConfig(): Promise<SystemConfigItem[]> {
  return request<SystemConfigItem[]>(`${BASE}/system/config`);
}

export async function updateSystemConfig(updates: UpdateSystemConfigInput): Promise<SystemConfigItem[]> {
  return request<SystemConfigItem[]>(`${BASE}/system/config`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  });
}

export async function resetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
  return request<SystemConfigItem[]>(`${BASE}/system/config/reset`, {
    method: 'POST',
    body: key ? JSON.stringify({ key }) : JSON.stringify({}),
  });
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Activity Logs
// ════════════════════════════════════════════════════════════════

export async function fetchActivityLogs(params?: ActivityLogParams): Promise<AdminActivityLogList> {
  const qs = params ? buildQuery(params as Record<string, string | number | boolean | undefined>) : '';
  return request<AdminActivityLogList>(`${BASE}/activity-logs${qs}`);
}

export async function fetchActivityLogActions(lang?: string): Promise<AdminActionLabel[]> {
  const qs = lang ? `?lang=${encodeURIComponent(lang)}` : '';
  return request<AdminActionLabel[]>(`${BASE}/activity-logs/actions${qs}`);
}

// ════════════════════════════════════════════════════════════════
//  API FUNCTIONS – Cleanup
// ════════════════════════════════════════════════════════════════

export async function queryCleanup(data: CleanupQueryInput): Promise<CleanupQueryResult> {
  return request<CleanupQueryResult>(`${BASE}/cleanup/query`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function deleteUserFile(userFileId: string): Promise<DeleteActionResult> {
  return request<DeleteActionResult>(`${BASE}/cleanup/delete-user-file`, {
    method: 'POST',
    body: JSON.stringify({ userFileId }),
  });
}

export async function deletePhysicalFile(physicalFileId: string): Promise<DeleteActionResult> {
  return request<DeleteActionResult>(`${BASE}/cleanup/delete-physical-file`, {
    method: 'POST',
    body: JSON.stringify({ physicalFileId }),
  });
}

// ════════════════════════════════════════════════════════════════
//  OLD API – backward-compatible aliases so existing pages work
//  until they are migrated to the new naming convention.
// ════════════════════════════════════════════════════════════════

// Types
export type AdminDashboard = AdminDashboardStats;

// Users
export const adminListUsers = (limit = 20, offset = 0): Promise<AdminUserList> =>
  fetchUsers({ limit, offset });
export const adminGetUser = fetchUser;
export const adminUpdateRole = updateUserRole;
export const adminUpdateStorageBase = updateStorageBase;
export const adminDeleteUser = deleteUser;
export async function adminDeleteUsers(ids: string[]): Promise<void> {
  await request<void>(`${BASE}/users/batch`, {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  });
}
export const adminSearchUsers = searchUsers;

// Files
export const adminListFiles = (limit = 20, offset = 0): Promise<AdminFileList> =>
  fetchFiles({ limit, offset });
export async function adminDeleteFiles(ids: string[]): Promise<void> {
  await request<void>(`${BASE}/files/batch`, {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  });
}

// Dashboard
export const adminGetDashboard = fetchDashboardStats;

// Storage
export type AdminStorageStat = {
  id: string;
  username: string;
  usedBytes: number;
  totalBytes: number;
  baseBytes?: number;
};
export const adminListStorageStats = (limit = 20): Promise<{ items: AdminStorageStat[]; total: number }> =>
  request<{ items: AdminStorageStat[]; total: number }>(`${BASE}/storage?limit=${limit}`);

// Settings
export type AdminSettings = {
  siteName: string;
  allowRegistration: boolean;
  defaultQuota: number;
  maxUploadSize: number;
};
export async function adminGetSettings(): Promise<AdminSettings> {
  return request<AdminSettings>(`${BASE}/settings`);
}
export async function adminUpdateSettings(input: AdminSettings): Promise<AdminSettings> {
  return request<AdminSettings>(`${BASE}/settings`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}