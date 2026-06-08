import { api } from './client';
import type { FileItem } from './files';

export type MediaItem = {
	mediaSlug: string;
	fileName: string;
	status: 'pending' | 'processing' | 'done' | 'failed';
	progress: number;
	durationSec: number | null;
	errorMsg: string | null;
	posterUrl: string | null;
	playUrl: string | null;
	createdAt: string;
};

export type AddToLibraryResponse = {
	mediaSlug: string;
	transcodeSlug: string;
	transcodeStatus: string;
	transcodeReused: boolean;
	alreadyInLibrary: boolean;
};

export async function addToLibrary(fileSlug: string) {
	return api<AddToLibraryResponse>('/api/v1/media/items', {
		method: 'POST',
		body: JSON.stringify({ fileSlug })
	});
}

export async function readdExistingUploadToLibrary(physicalFileSlug: string, fileName: string) {
	return api<AddToLibraryResponse>('/api/v1/media/items/readd-existing', {
		method: 'POST',
		body: JSON.stringify({ physicalFileSlug, fileName })
	});
}

export async function listMedia(page = 1, pageSize = 50) {
	return api<{ items: MediaItem[]; total: number }>(`/api/v1/media/items?page=${page}&pageSize=${pageSize}`);
}

export async function getMediaItem(mediaSlug: string) {
	return api<MediaItem>(`/api/v1/media/items/${mediaSlug}`);
}

export async function removeFromLibrary(mediaSlug: string) {
	return api(`/api/v1/media/items/${mediaSlug}`, { method: 'DELETE' });
}

export async function batchRemoveFromLibrary(mediaSlugs: string[]) {
	return api('/api/v1/media/items/batch-delete', {
		method: 'POST',
		body: JSON.stringify({ mediaSlugs })
	});
}

export async function renameMediaItem(mediaSlug: string, newName: string) {
	return api<MediaItem>(`/api/v1/media/items/${mediaSlug}/rename`, {
		method: 'POST',
		body: JSON.stringify({ newName })
	});
}

export async function ensureMediaUploadDir() {
	return api<FileItem>('/api/v1/media/upload-dir');
}

export function getHLSUrl(mediaSlug: string, fileName: string) {
	return `/api/v1/media/hls/${mediaSlug}/${fileName}`;
}
