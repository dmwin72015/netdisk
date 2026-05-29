import { api } from './client';

export type AdminUser = {
	id: string;
	username: string;
	email: string;
	role: string;
	usedBytes: number;
	baseBytes: number;
	memberBonusBytes: number;
	packBytes: number;
	totalBytes: number;
	createdAt: number;
};

export type AdminUserList = {
	items: AdminUser[];
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
