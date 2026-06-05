import { api } from './client';

export type PhotoItem = {
	slug: string;
	fileName: string;
	mimeType: string;
	fileSize: number;
	width: number;
	height: number;
	fileHash: string;
	isStarred: boolean;
	createdAt: string;
	albums: { slug: string; title: string }[];
};

export type PhotoListResponse = {
	items: PhotoItem[];
	total: number;
	page: number;
};

export async function listPhotos(page = 1, pageSize = 50): Promise<PhotoListResponse> {
	return api<PhotoListResponse>(`/api/v1/photos?page=${page}&pageSize=${pageSize}`);
}

export async function getPhotoDetail(fileSlug: string): Promise<PhotoItem> {
	return api<PhotoItem>(`/api/v1/photos/${fileSlug}`);
}

export function thumbnailUrl(fileSlug: string): string {
	return `/api/v1/photos/${fileSlug}/thumbnail`;
}

export function photoDetailUrl(fileSlug: string): string {
	return `/api/v1/photos/${fileSlug}`;
}

export async function listPhotoAlbums(fileSlug: string): Promise<{ items: { slug: string; title: string }[] }> {
	return api<{ items: { slug: string; title: string }[] }>(`/api/v1/photos/${fileSlug}/albums`);
}
