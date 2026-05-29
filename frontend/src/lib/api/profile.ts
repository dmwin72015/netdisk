import { api } from './client';

export type ProfileData = {
	slug: string;
	username: string;
	email: string;
	profile: {
		displayName: string;
		avatarUrl: string;
		bio: string;
	};
	storage: {
		storageUsed: number;
		storageQuota: number;
	};
	level: {
		levelCode: string;
		levelName: string;
		expiresAt: string | null;
	};
	createdAt: string;
};

export async function getProfile(): Promise<ProfileData> {
	return api<ProfileData>('/api/v1/user/me');
}

export async function updateProfile(data: { displayName?: string; bio?: string }): Promise<{ message: string }> {
	return api<{ message: string }>('/api/v1/user/profile', {
		method: 'PATCH',
		body: JSON.stringify(data),
	});
}

export async function uploadAvatar(file: File): Promise<{ avatarUrl: string }> {
	const form = new FormData();
	form.append('file', file);
	return api<{ avatarUrl: string }>('/api/v1/user/me/avatar', {
		method: 'POST',
		body: form,
	});
}

export type CategoryStat = {
	category: string;
	bytes: number;
	count: number;
};

export async function getStorageBreakdown(): Promise<CategoryStat[]> {
	const data = await api<{ categories: CategoryStat[] }>('/api/v1/user/storage-breakdown');
	return data.categories;
}
