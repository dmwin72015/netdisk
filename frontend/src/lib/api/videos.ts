import { api } from './client';
import type { Task } from './tasks';

export type VideoList = {
	items: Task[];
	limit: number;
	offset: number;
};

export async function listVideos(limit = 20, offset = 0): Promise<VideoList> {
	return api<VideoList>(`/api/v1/videos?limit=${limit}&offset=${offset}`);
}

export async function getVideo(id: string): Promise<Task> {
	return api<Task>(`/api/v1/videos/${id}`);
}

export async function uploadThumbnail(id: string, file: File): Promise<Task> {
	const fd = new FormData();
	fd.append('file', file);
	return api<Task>(`/api/v1/videos/${id}/thumbnail`, { method: 'POST', body: fd });
}

export async function captureFrameThumbnail(id: string, atSec: number): Promise<Task> {
	const at = encodeURIComponent(atSec.toFixed(2));
	return api<Task>(`/api/v1/videos/${id}/thumbnail/frame?at=${at}`, { method: 'POST' });
}
