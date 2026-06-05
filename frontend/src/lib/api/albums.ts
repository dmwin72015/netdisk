import { api } from './client';
import type { PhotoItem } from './photos';

export type Album = {
	slug: string;
	title: string;
	description: string;
	coverFileSlug: string | null;
	coverUrl: string | null;
	itemCount: number;
	createdAt: string;
	updatedAt: string;
};

export type AlbumListResponse = {
	items: Album[];
	total: number;
	page: number;
};

export type AlbumDetail = Album & {
	photos: PhotoItem[];
};

export async function listAlbums(page = 1, pageSize = 50): Promise<AlbumListResponse> {
	return api<AlbumListResponse>(`/api/v1/albums?page=${page}&pageSize=${pageSize}`);
}

export async function getAlbum(albumSlug: string): Promise<AlbumDetail> {
	return api<AlbumDetail>(`/api/v1/albums/${albumSlug}`);
}

export async function createAlbum(data: { title: string; description?: string }): Promise<Album> {
	return api<Album>('/api/v1/albums', {
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateAlbum(albumSlug: string, data: { title?: string; description?: string }): Promise<Album> {
	return api<Album>(`/api/v1/albums/${albumSlug}`, {
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteAlbum(albumSlug: string): Promise<void> {
	await api(`/api/v1/albums/${albumSlug}`, { method: 'DELETE' });
}

export async function addPhotosToAlbum(albumSlug: string, fileSlugs: string[]): Promise<void> {
	await api(`/api/v1/albums/${albumSlug}/photos`, {
		method: 'POST',
		body: JSON.stringify({ fileSlugs })
	});
}

export async function listAlbumPhotos(albumSlug: string, page = 1, pageSize = 50): Promise<{ items: PhotoItem[]; total: number }> {
	return api<{ items: PhotoItem[]; total: number }>(`/api/v1/albums/${albumSlug}/photos?page=${page}&pageSize=${pageSize}`);
}

export async function removePhotoFromAlbum(albumSlug: string, fileSlug: string): Promise<void> {
	await api(`/api/v1/albums/${albumSlug}/photos/${fileSlug}`, { method: 'DELETE' });
}
