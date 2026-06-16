import { api } from './client';

export type UserSettings = {
	showSystemDirs: boolean;
	uploadConcurrency: number;
	duplicateStrategy: string;
	directoryUnlockTtlHours: number;
};

export async function getUserSettings() {
	return api<UserSettings>('/api/v1/user/settings');
}

export async function saveUserSettings(settings: UserSettings) {
	return api<UserSettings>('/api/v1/user/settings', {
		method: 'PUT',
		body: JSON.stringify(settings),
	});
}
