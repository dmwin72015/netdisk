import { api } from './client';

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
	slug: string;
	userId: string;
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

export async function adminDashboardStats(): Promise<AdminDashboardStats> {
	return api<AdminDashboardStats>('/api/v1/admin/dashboard/stats');
}

export async function adminListUsers(
	limit = 20,
	offset = 0,
	search?: string,
	role?: string,
	sort?: string
): Promise<AdminUserList> {
	const params = new URLSearchParams();
	params.set('limit', String(limit));
	params.set('offset', String(offset));
	if (search) params.set('search', search);
	if (role) params.set('role', role);
	if (sort) params.set('sort', sort);
	return api<AdminUserList>(`/api/v1/admin/users?${params}`);
}

export async function adminCreateUser(username: string, email: string, password: string, role = 'user'): Promise<AdminUser> {
	return api<AdminUser>('/api/v1/admin/users', {
		method: 'POST',
		body: JSON.stringify({ username, email, password, role }),
	});
}

export async function adminGetUser(id: string): Promise<AdminUser> {
	return api<AdminUser>(`/api/v1/admin/users/${id}`);
}

export async function adminUpdateRole(id: string, role: string): Promise<AdminUser> {
	return api<AdminUser>(`/api/v1/admin/users/${id}`, {
		method: 'PATCH',
		body: JSON.stringify({ role }),
	});
}

export async function adminUpdateStorageBase(id: string, baseBytes: number): Promise<AdminUser> {
	return api<AdminUser>(`/api/v1/admin/users/${id}/storage-base`, {
		method: 'PATCH',
		body: JSON.stringify({ baseBytes }),
	});
}

export async function adminDeleteUser(id: string): Promise<void> {
	await api<void>(`/api/v1/admin/users/${id}`, { method: 'DELETE' });
}

export async function adminListFiles(
	limit = 20,
	offset = 0,
	search?: string,
	category?: string,
	trashed?: string,
	sort?: string
): Promise<AdminFileList> {
	const params = new URLSearchParams();
	params.set('limit', String(limit));
	params.set('offset', String(offset));
	if (search) params.set('search', search);
	if (category) params.set('category', category);
	if (trashed) params.set('trashed', trashed);
	if (sort) params.set('sort', sort);
	return api<AdminFileList>(`/api/v1/admin/files?${params}`);
}

export async function adminDeleteFile(id: string): Promise<void> {
	await api<void>(`/api/v1/admin/files/${id}`, { method: 'DELETE' });
}

export async function adminRestoreFile(id: string): Promise<void> {
	await api<void>(`/api/v1/admin/files/${id}/restore`, { method: 'PATCH' });
}

export type CategoryStat = {
	category: string;
	bytes: number;
	count: number;
};

export async function adminStorageStats(): Promise<CategoryStat[]> {
	return api<CategoryStat[]>('/api/v1/admin/storage/stats');
}

export type AdminSystemInfo = {
	upload: { chunkSize: number; maxUploadSize: number };
	limits: { defaultStorageQuota: number; avatarMaxSize: number };
	trash: { retentionDays: number };
	jwt: { accessTTLMin: number; refreshTTLHour: number };
	server: { port: number };
};

export async function adminSystemInfo(): Promise<AdminSystemInfo> {
	return api<AdminSystemInfo>('/api/v1/admin/system/info');
}

export type SystemConfigItem = {
	key: string;
	value: number | string | boolean;
	defaultValue: number | string | boolean;
	description: string;
	type: 'bytes' | 'number' | 'string' | 'bool';
};

export async function adminListSystemConfig(): Promise<SystemConfigItem[]> {
	return api<SystemConfigItem[]>('/api/v1/admin/system/config');
}

export async function adminUpdateSystemConfig(updates: Record<string, unknown>): Promise<SystemConfigItem[]> {
	return api<SystemConfigItem[]>('/api/v1/admin/system/config', {
		method: 'PUT',
		body: JSON.stringify(updates),
	});
}

export async function adminResetSystemConfig(key?: string): Promise<SystemConfigItem[]> {
	return api<SystemConfigItem[]>('/api/v1/admin/system/config/reset', {
		method: 'POST',
		body: JSON.stringify(key ? { key } : {}),
	});
}
