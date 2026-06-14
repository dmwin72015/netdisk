import { UPLOAD_FILE_CONCURRENCY, UPLOAD_FILE_CONCURRENCY_DEFAULT } from '$lib/upload-concurrency';

const browser = typeof window !== 'undefined';

function save(key: string, value: string) {
	if (browser) localStorage.setItem(key, value);
}

let _showSystemDirs = $state<boolean>(false);
let _uploadConcurrency = $state<number>(UPLOAD_FILE_CONCURRENCY_DEFAULT);
let _duplicateStrategy = $state<string>('prompt');

if (browser) {
	const sv = localStorage.getItem('nd.files.showSystemDirs');
	if (sv === 'false') _showSystemDirs = false;
	else _showSystemDirs = true;
	const cv = localStorage.getItem('nd.files.uploadConcurrency');
	if (cv) _uploadConcurrency = Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, parseInt(cv, 10) || UPLOAD_FILE_CONCURRENCY_DEFAULT));
	_duplicateStrategy = localStorage.getItem('nd.files.duplicateStrategy') || 'prompt';
}

export function getShowSystemDirs() { return _showSystemDirs; }
export function getUploadConcurrency() { return _uploadConcurrency; }
export function getDuplicateStrategy() { return _duplicateStrategy; }

export function setShowSystemDirs(value: boolean) {
	_showSystemDirs = value;
	save('nd.files.showSystemDirs', String(value));
}

export function setUploadConcurrency(value: number) {
	const clamped = Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, value || UPLOAD_FILE_CONCURRENCY_DEFAULT));
	_uploadConcurrency = clamped;
	save('nd.files.uploadConcurrency', String(clamped));
}

export function setDuplicateStrategy(value: string) {
	_duplicateStrategy = value;
	save('nd.files.duplicateStrategy', value);
}
