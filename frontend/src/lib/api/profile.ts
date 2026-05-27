import { api } from './client';

export type ProfileData = {
	user_id: number;
	nickname: string;
	avatar_url: string;
	bio: string;
};

export async function getProfile(): Promise<ProfileData> {
	return api<ProfileData>('/api/v1/profile');
}

export async function updateProfile(data: { nickname?: string; bio?: string }): Promise<ProfileData> {
	return api<ProfileData>('/api/v1/profile', {
		method: 'PUT',
		body: JSON.stringify(data),
	});
}

export async function uploadAvatar(file: File): Promise<{ avatar_url: string }> {
	const form = new FormData();
	form.append('file', file);
	return api<{ avatar_url: string }>('/api/v1/profile/avatar', {
		method: 'POST',
		body: form,
	});
}
