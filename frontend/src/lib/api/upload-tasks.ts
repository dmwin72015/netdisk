import { api } from './client';

export type UploadTaskItem = {
	slug: string;
	fileName: string;
	fileSize: number;
	mimeType: string;
	status: string;
	errorMsg: string;
	totalChunks: number;
	createdAt: string;
	updatedAt: string;
};

export type UploadTaskListResponse = {
	items: UploadTaskItem[];
	total: number;
	limit: number;
	offset: number;
};

export type RetryResponse = {
	uploadSlug: string;
	totalChunks: number;
	chunkSize: number;
	completedChunks: number[];
};

export async function listUploadTasks(limit = 20, offset = 0): Promise<UploadTaskListResponse> {
	const data = await api<UploadTaskListResponse>(`/api/v1/upload/tasks?limit=${limit}&offset=${offset}`);
	return data;
}

export async function retryUploadTask(slug: string): Promise<RetryResponse> {
	const data = await api<RetryResponse>(`/api/v1/upload/tasks/${slug}/retry`, {
		method: 'POST',
	});
	return data;
}
