import { api } from './client';

export type OAuthAccountInfo = {
	provider: string;
	providerAccountId: string;
	oauthEmail?: string;
	createdAt: string;
};

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
	oauthAccounts: OAuthAccountInfo[];
	createdAt: string;
};

export async function getProfile(): Promise<ProfileData> {
	return api<ProfileData>('/api/v1/user/me');
}

export async function updateProfile(data: { displayName?: string; bio?: string; avatarUrl?: string }): Promise<{ message: string }> {
	return api<{ message: string }>('/api/v1/user/profile', {
		method: 'PATCH',
		body: JSON.stringify(data),
	});
}

export async function uploadAvatar(file: File): Promise<string> {
	const form = new FormData();
	form.append('file', file);
	const res = await api<{ avatar_url: string }>('/api/v1/user/me/avatar', {
		method: 'POST',
		body: form,
	});
	return res.avatar_url;
}

export type CategoryStat = {
	category: string;
	bytes: number;
	count: number;
};

export async function unlinkOAuth(provider: string): Promise<void> {
	await api<{ message: string }>(`/api/v1/user/oauth/${provider}`, {
		method: 'DELETE',
	});
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
	await api('/api/v1/user/me/password', {
		method: 'POST',
		body: JSON.stringify({ oldPassword, newPassword }),
	});
}

export async function getStorageBreakdown(): Promise<CategoryStat[]> {
	const data = await api<{ categories: CategoryStat[] }>('/api/v1/user/storage-breakdown');
	return data.categories;
}
