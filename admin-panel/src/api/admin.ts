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

export type AdminDashboard = {
  totalUsers: number;
  totalFiles: number;
  totalStorage: number;
  todayActive: number;
};

export type AdminStorageStat = {
  id: string;
  username: string;
  usedBytes: number;
  totalBytes: number;
  baseBytes?: number;
};

export type AdminSettings = {
  siteName: string;
  allowRegistration: boolean;
  defaultQuota: number;
  maxUploadSize: number;
};

function getToken(): string | null {
  try {
    return localStorage.getItem('nd.access');
  } catch {
    return null;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
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

// Users
export async function adminListUsers(
  limit = 20,
  offset = 0
): Promise<AdminUserList> {
  return request<AdminUserList>(
    `/api/v1/admin/users?limit=${limit}&offset=${offset}`
  );
}

export async function adminGetUser(id: string): Promise<AdminUser> {
  return request<AdminUser>(`/api/v1/admin/users/${id}`);
}

export async function adminUpdateRole(
  id: string,
  role: string
): Promise<AdminUser> {
  return request<AdminUser>(`/api/v1/admin/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ role }),
  });
}

export async function adminUpdateStorageBase(
  id: string,
  baseBytes: number
): Promise<AdminUser> {
  return request<AdminUser>(`/api/v1/admin/users/${id}/storage-base`, {
    method: 'PATCH',
    body: JSON.stringify({ baseBytes }),
  });
}

export async function adminDeleteUser(id: string): Promise<void> {
  await request<void>(`/api/v1/admin/users/${id}`, { method: 'DELETE' });
}

export async function adminDeleteUsers(ids: string[]): Promise<void> {
  await request<void>(`/api/v1/admin/users/batch`, {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  });
}

export async function adminSearchUsers(
  query: string
): Promise<{ id: string; username: string }[]> {
  return request<{ id: string; username: string }[]>(
    `/api/v1/admin/users/search?q=${encodeURIComponent(query)}`
  );
}

// Files
export async function adminListFiles(
  limit = 20,
  offset = 0
): Promise<AdminFileList> {
  return request<AdminFileList>(
    `/api/v1/admin/files?limit=${limit}&offset=${offset}`
  );
}

export async function adminDeleteFiles(ids: string[]): Promise<void> {
  await request<void>(`/api/v1/admin/files/batch`, {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  });
}

// Dashboard
export async function adminGetDashboard(): Promise<AdminDashboard> {
  return request<AdminDashboard>(`/api/v1/admin/dashboard`);
}

// Storage
export async function adminListStorageStats(
  limit = 20
): Promise<{ items: AdminStorageStat[]; total: number }> {
  return request<{ items: AdminStorageStat[]; total: number }>(
    `/api/v1/admin/storage?limit=${limit}`
  );
}

// Settings
export async function adminGetSettings(): Promise<AdminSettings> {
  return request<AdminSettings>(`/api/v1/admin/settings`);
}

export async function adminUpdateSettings(
  input: AdminSettings
): Promise<AdminSettings> {
  return request<AdminSettings>(`/api/v1/admin/settings`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}
