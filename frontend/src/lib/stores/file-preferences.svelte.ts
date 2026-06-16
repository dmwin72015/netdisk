import { browser } from '$app/environment';
import { getUserSettings, saveUserSettings, type UserSettings } from '$lib/api/settings';
import { UPLOAD_FILE_CONCURRENCY, UPLOAD_FILE_CONCURRENCY_DEFAULT } from '$lib/upload-concurrency';

export const DIRECTORY_UNLOCK_TTL_OPTIONS = [1, 2, 6, 24, -1] as const;
export type DirectoryUnlockTtlHours = (typeof DIRECTORY_UNLOCK_TTL_OPTIONS)[number];

const LS_SHOW_SYSTEM_DIRS = 'nd.files.showSystemDirs';
const LS_UPLOAD_CONCURRENCY = 'nd.files.uploadConcurrency';
const LS_DUPLICATE_STRATEGY = 'nd.files.duplicateStrategy';
const LS_DIRECTORY_UNLOCK_TTL = 'nd.files.directoryUnlockTtlHours';

const DEFAULT_SETTINGS: UserSettings = {
	showSystemDirs: true,
	uploadConcurrency: UPLOAD_FILE_CONCURRENCY_DEFAULT,
	duplicateStrategy: 'prompt',
	directoryUnlockTtlHours: 2,
};

function save(key: string, value: string) {
	if (browser) localStorage.setItem(key, value);
}

function clampUploadConcurrency(value: number) {
	return Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, value || UPLOAD_FILE_CONCURRENCY_DEFAULT));
}

function normalizeDuplicateStrategy(value: string) {
	return ['prompt', 'overwrite', 'keep_both', 'skip'].includes(value) ? value : DEFAULT_SETTINGS.duplicateStrategy;
}

function normalizeDirectoryUnlockTtlHours(value: number) {
	return DIRECTORY_UNLOCK_TTL_OPTIONS.includes(value as DirectoryUnlockTtlHours)
		? value
		: DEFAULT_SETTINGS.directoryUnlockTtlHours;
}

function normalizeSettings(settings: Partial<UserSettings>): UserSettings {
	return {
		showSystemDirs: typeof settings.showSystemDirs === 'boolean' ? settings.showSystemDirs : DEFAULT_SETTINGS.showSystemDirs,
		uploadConcurrency: clampUploadConcurrency(Number(settings.uploadConcurrency ?? DEFAULT_SETTINGS.uploadConcurrency)),
		duplicateStrategy: normalizeDuplicateStrategy(String(settings.duplicateStrategy ?? DEFAULT_SETTINGS.duplicateStrategy)),
		directoryUnlockTtlHours: normalizeDirectoryUnlockTtlHours(Number(settings.directoryUnlockTtlHours ?? DEFAULT_SETTINGS.directoryUnlockTtlHours)),
	};
}

let _showSystemDirs = $state<boolean>(DEFAULT_SETTINGS.showSystemDirs);
let _uploadConcurrency = $state<number>(DEFAULT_SETTINGS.uploadConcurrency);
let _duplicateStrategy = $state<string>(DEFAULT_SETTINGS.duplicateStrategy);
let _directoryUnlockTtlHours = $state<number>(DEFAULT_SETTINGS.directoryUnlockTtlHours);
let serverLoaded = false;
let persistTimer: ReturnType<typeof setTimeout> | null = null;

if (browser) {
	const sv = localStorage.getItem(LS_SHOW_SYSTEM_DIRS);
	if (sv === 'false') _showSystemDirs = false;
	else _showSystemDirs = true;
	const cv = localStorage.getItem(LS_UPLOAD_CONCURRENCY);
	if (cv) _uploadConcurrency = clampUploadConcurrency(parseInt(cv, 10));
	_duplicateStrategy = normalizeDuplicateStrategy(localStorage.getItem(LS_DUPLICATE_STRATEGY) || 'prompt');
	const tv = localStorage.getItem(LS_DIRECTORY_UNLOCK_TTL);
	if (tv) _directoryUnlockTtlHours = normalizeDirectoryUnlockTtlHours(parseInt(tv, 10));
}

function applySettings(settings: UserSettings, persistLocal = true) {
	_showSystemDirs = settings.showSystemDirs;
	_uploadConcurrency = settings.uploadConcurrency;
	_duplicateStrategy = settings.duplicateStrategy;
	_directoryUnlockTtlHours = settings.directoryUnlockTtlHours;
	if (persistLocal) {
		save(LS_SHOW_SYSTEM_DIRS, String(settings.showSystemDirs));
		save(LS_UPLOAD_CONCURRENCY, String(settings.uploadConcurrency));
		save(LS_DUPLICATE_STRATEGY, settings.duplicateStrategy);
		save(LS_DIRECTORY_UNLOCK_TTL, String(settings.directoryUnlockTtlHours));
	}
}

function currentSettings(): UserSettings {
	return normalizeSettings({
		showSystemDirs: _showSystemDirs,
		uploadConcurrency: _uploadConcurrency,
		duplicateStrategy: _duplicateStrategy,
		directoryUnlockTtlHours: _directoryUnlockTtlHours,
	});
}

function schedulePersist() {
	if (!browser || !serverLoaded) return;
	if (persistTimer) clearTimeout(persistTimer);
	persistTimer = setTimeout(() => {
		persistTimer = null;
		void saveUserSettings(currentSettings())
			.then((settings) => applySettings(settings))
			.catch(() => undefined)
	}, 300);
}

export function getShowSystemDirs() { return _showSystemDirs; }
export function getUploadConcurrency() { return _uploadConcurrency; }
export function getDuplicateStrategy() { return _duplicateStrategy; }
export function getDirectoryUnlockTtlHours() { return _directoryUnlockTtlHours; }

export function setShowSystemDirs(value: boolean) {
	_showSystemDirs = value;
	save(LS_SHOW_SYSTEM_DIRS, String(value));
	schedulePersist();
}

export function setUploadConcurrency(value: number) {
	const clamped = clampUploadConcurrency(value);
	_uploadConcurrency = clamped;
	save(LS_UPLOAD_CONCURRENCY, String(clamped));
	schedulePersist();
}

export function setDuplicateStrategy(value: string) {
	_duplicateStrategy = normalizeDuplicateStrategy(value);
	save(LS_DUPLICATE_STRATEGY, _duplicateStrategy);
	schedulePersist();
}

export function setDirectoryUnlockTtlHours(value: number) {
	_directoryUnlockTtlHours = normalizeDirectoryUnlockTtlHours(value);
	save(LS_DIRECTORY_UNLOCK_TTL, String(_directoryUnlockTtlHours));
	schedulePersist();
}

export async function loadPreferencesFromServer() {
	if (!browser) return currentSettings();
	try {
		const settings = await getUserSettings();
		applySettings(normalizeSettings(settings));
		serverLoaded = true;
	} catch {
		serverLoaded = true;
	}
	return currentSettings();
}

export function exportPreferences(): UserSettings {
	return currentSettings();
}

export async function importPreferences(settings: Partial<UserSettings>) {
	const normalized = normalizeSettings(settings);
	serverLoaded = true;
	if (persistTimer) {
		clearTimeout(persistTimer);
		persistTimer = null;
	}
	const saved = await saveUserSettings(normalized);
	applySettings(saved);
	return currentSettings();
}
