import { api } from './client';

export type Task = {
	id: string;
	originalName: string;
	fileSize: number;
	status: 'pending' | 'processing' | 'completed' | 'failed';
	progress: number;
	errorMessage?: string;
	m3u8Url?: string;
	thumbnailUrl?: string;
	durationSec?: number;
	createdAt: number;
	startedAt?: number;
	completedAt?: number;
};

export type TaskList = {
	items: Task[];
	total: number;
	limit: number;
	offset: number;
};

export async function listTasks(limit = 20, offset = 0): Promise<TaskList> {
	return api<TaskList>(`/api/v1/tasks?limit=${limit}&offset=${offset}`);
}

export async function getTask(id: string): Promise<Task> {
	return api<Task>(`/api/v1/tasks/${id}`);
}

export async function deleteTask(id: string): Promise<void> {
	await api<void>(`/api/v1/tasks/${id}`, { method: 'DELETE' });
}

/**
 * Compute SHA-256 digest of a File. Reads in 8 MiB chunks to avoid loading
 * very large files entirely into memory.
 */
export async function computeFileSHA256(file: File): Promise<string> {
	const CHUNK = 8 * 1024 * 1024;
	const chunks: ArrayBuffer[] = [];
	let offset = 0;
	while (offset < file.size) {
		const end = Math.min(offset + CHUNK, file.size);
		chunks.push(await file.slice(offset, end).arrayBuffer());
		offset = end;
	}
	const totalLen = chunks.reduce((a, b) => a + b.byteLength, 0);
	const full = new Uint8Array(totalLen);
	let pos = 0;
	for (const buf of chunks) {
		full.set(new Uint8Array(buf), pos);
		pos += buf.byteLength;
	}
	const digest = await crypto.subtle.digest('SHA-256', full);
	return Array.from(new Uint8Array(digest))
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
}

export async function uploadFile(file: File, sha256?: string): Promise<{ taskId: string }> {
	const headers: Record<string, string> = {};
	if (sha256) headers['X-File-SHA256'] = sha256;
	const fd = new FormData();
	fd.append('file', file);
	return api<{ taskId: string }>('/api/v1/upload', {
		method: 'POST',
		body: fd,
		headers
	});
}
