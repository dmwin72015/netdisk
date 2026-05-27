import { api } from './client';

export type MediaItem = {
	media_slug: string;
	file_name: string;
	status: 'pending' | 'processing' | 'done' | 'failed';
	progress: number;
	duration_sec: number | null;
	error_msg: string | null;
	created_at: string;
};

export async function addToLibrary(fileSlug: string) {
	return api<{ media_slug: string }>('/api/v1/media/items', {
		method: 'POST',
		body: JSON.stringify({ file_slug: fileSlug })
	});
}

export async function listMedia(page = 1, pageSize = 50) {
	return api<{ items: MediaItem[]; total: number }>(`/api/v1/media/items?page=${page}&page_size=${pageSize}`);
}

export async function getMediaItem(mediaSlug: string) {
	return api<MediaItem>(`/api/v1/media/items/${mediaSlug}`);
}

export async function removeFromLibrary(mediaSlug: string) {
	return api(`/api/v1/media/items/${mediaSlug}`, { method: 'DELETE' });
}

export function getHLSUrl(mediaSlug: string, fileName: string) {
	return `/api/v1/media/hls/${mediaSlug}/${fileName}`;
}
