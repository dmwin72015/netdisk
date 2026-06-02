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

export async function adminListUsers(limit = 20, offset = 0): Promise<AdminUserList> {
	return api<AdminUserList>(`/api/v1/admin/users?limit=${limit}&offset=${offset}`);
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

export async function adminListFiles(limit = 20, offset = 0): Promise<AdminFileList> {
	return api<AdminFileList>(`/api/v1/admin/files?limit=${limit}&offset=${offset}`);
}
