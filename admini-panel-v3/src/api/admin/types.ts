// --- Users ---
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

export type UpdateRoleInput = { role: string; };
export type UpdateStorageBaseInput = { baseBytes: number; };

// --- Files ---
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

// --- Physical Files ---
export type AdminPhysicalFile = {
  id: number;
  slug: string;
  hashAlgo: string;
  fileHash: string;
  preHash: string;
  fileSize: number;
  mimeType: string;
  storagePath: string;
  status: string;
  createdAt: number;
  userFileCount: number;
  mediaItemCount: number;
};

export type AdminPhysicalFileList = {
  items: AdminPhysicalFile[];
  total: number;
  limit: number;
  offset: number;
};

export type PhysicalFileListParams = {
  limit?: number;
  offset?: number;
  search?: string;
  status?: string;
  hash_filter?: string;
  mime_filter?: string;
  min_size?: number;
  max_size?: number;
  created_from?: string;
  created_to?: string;
};

export type PhysicalFileDetailResult = {
  physicalFile: AdminPhysicalFile;
  userFiles: CleanupQueryUserFile[];
  totalUploads: number;
  uniqueUsers: number;
  fullPath: string;
};

// --- Dashboard ---
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

// --- Storage ---
export type CategoryStat = {
  category: string;
  bytes: number;
  count: number;
};

// --- System Config ---
export type SystemConfigItem = {
  key: string;
  value: any;
  defaultValue?: any;
  description?: string;
};

export type UpdateSystemConfigInput = Record<string, string>;

// --- Activity Logs ---
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
  createdAt: string;
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
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
};

export type AdminActionLabel = {
  action: string;
  label: string;
};

// --- Cleanup ---
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
