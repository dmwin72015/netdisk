import { api } from './client';

export type UploadTaskItem = {
	slug: string;
	fileName: string;
	fileSize: number;
	mimeType: string;
	status: string;
	errorMsg: string;
	totalChunks: number;
	receivedBytes: number;
	parentSlug?: string;
	parentName?: string;
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

export async function listUploadTasks(
	limit = 20,
	offset = 0,
	startDate?: string,
	endDate?: string,
	status?: string
): Promise<UploadTaskListResponse> {
	const params = new URLSearchParams({ limit: String(limit), offset: String(offset) });
	if (startDate) params.set('start_date', startDate);
	if (endDate) params.set('end_date', endDate);
	if (status) params.set('status', status);
	const data = await api<UploadTaskListResponse>(`/api/v1/upload/tasks?${params}`);
	return data;
}

export async function retryUploadTask(slug: string): Promise<RetryResponse> {
	const data = await api<RetryResponse>(`/api/v1/upload/tasks/${slug}/retry`, {
		method: 'POST',
	});
	return data;
}

export async function deleteUploadTask(slug: string): Promise<void> {
	await api<void>(`/api/v1/upload/tasks/${slug}`, { method: 'DELETE' });
}

export async function deleteUploadTasks(slugs: string[]): Promise<void> {
	await api<void>('/api/v1/upload/tasks', {
		method: 'DELETE',
		body: JSON.stringify({ slugs }),
	});
}
