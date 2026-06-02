import { api, getAccessToken } from './client';
import { computeSHA256 } from '$lib/upload-hash';
import * as m from '$lib/paraglide/messages';
import { getChunkSize, configError } from '$lib/stores/config';
import { get } from 'svelte/store';

export type DriveFile = {
	id: string;
	name: string;
	mimeType: string;
	fileCategory: string;
	size: number;
	createdAt: number;
	isDir: boolean;
	parentId?: string;
};

export type DriveList = {
	items: DriveFile[];
	total: number;
	limit: number;
	offset: number;
};

export type DriveSession = {
	id: string;
	name: string;
	totalSize: number;
	receivedBytes: number;
	createdAt: number;
	updatedAt: number;
};

export type DriveChunkResult = { id: string; receivedBytes: number };

export type DriveCheckHashResult = {
	exists: boolean;
	fileId?: string;
};

export type DriveClaimResult = {
	fileId: string;
};

export { computeSHA256 };

/** Files larger than this skip client-side SHA-256 to avoid OOM in the browser.
 *  The server computes the hash during Complete() for all files regardless. */
const MAX_HASH_SIZE = 200 * 1024 * 1024;

export async function driveStats(): Promise<{ usedBytes: number; baseBytes: number; memberBonusBytes: number; packBytes: number; totalBytes: number }> {
	return api<{ usedBytes: number; baseBytes: number; memberBonusBytes: number; packBytes: number; totalBytes: number }>('/api/v1/drive/stats');
}

export async function createDriveDir(name: string, parentId?: string): Promise<DriveFile> {
	return api<DriveFile>('/api/v1/drive/dir', {
		method: 'POST',
		body: JSON.stringify({ name, parentId }),
	});
}

export async function listDrive(limit = 50, offset = 0, q?: string, parentId?: string, category?: string | null): Promise<DriveList> {
	const params = `limit=${limit}&offset=${offset}${q ? `&q=${encodeURIComponent(q)}` : ''}${parentId ? `&parentId=${encodeURIComponent(parentId)}` : ''}${category ? `&fileCategory=${encodeURIComponent(category)}` : ''}`;
	return api<DriveList>(`/api/v1/drive?${params}`);
}

export async function getDriveFile(id: string): Promise<DriveFile> {
	return api<DriveFile>(`/api/v1/drive/${id}`);
}

export async function getDriveAncestors(id: string): Promise<DriveFile[]> {
	return api<DriveFile[]>(`/api/v1/drive/${id}/ancestors`);
}

// ---- Dedup ----

export async function driveCheckHash(sha256: string): Promise<DriveCheckHashResult> {
	return api<DriveCheckHashResult>('/api/v1/drive/check-sha256', {
		method: 'POST',
		body: JSON.stringify({ sha256 }),
	});
}

export async function driveClaimHash(
	sha256: string,
	originalName: string,
	mimeType: string,
	fileSize: number,
	parentId?: string | null,
): Promise<DriveClaimResult> {
	return api<DriveClaimResult>('/api/v1/drive/claim', {
		method: 'POST',
		body: JSON.stringify({ sha256, originalName, mimeType, fileSize, parentId: parentId || undefined }),
	});
}

// ---- Unified pre-upload check ----

export type DriveCheckUploadResult = {
	status: 'full' | 'partial' | 'none';
	fileId?: string;
	ownFile?: boolean;
	sessionId?: string;
	receivedBytes?: number;
};

export async function driveCheckUpload(
	sha256: string,
	fileSize: number,
	fileName: string,
	mimeType: string,
	parentId?: string | null,
): Promise<DriveCheckUploadResult> {
	return api<DriveCheckUploadResult>('/api/v1/drive/check-upload', {
		method: 'POST',
		body: JSON.stringify({ sha256, fileSize, fileName, mimeType, parentId: parentId || undefined }),
	});
}

// ---- Session management ----

export async function listDriveSessions(): Promise<DriveSession[]> {
	const data = await api<{ items: DriveSession[] }>('/api/v1/drive/uploads');
	return data.items;
}

export async function initDriveUpload(filename: string, mimeType: string, totalSize: number, parentId?: string | null, sha256?: string): Promise<DriveSession & { status?: string; fileId?: string }> {
	return api<DriveSession & { status?: string; fileId?: string }>('/api/v1/drive/uploads', {
		method: 'POST',
		body: JSON.stringify({ filename, mimeType, totalSize, parentId: parentId || undefined, sha256: sha256 || undefined }),
	});
}

export async function getDriveSession(id: string): Promise<DriveSession> {
	return api<DriveSession>(`/api/v1/drive/uploads/${id}`);
}

export async function uploadDriveChunk(
	id: string,
	offset: number,
	chunk: Blob,
): Promise<DriveChunkResult> {
	return api<DriveChunkResult>(`/api/v1/drive/uploads/${id}?offset=${offset}`, {
		method: 'PATCH',
		body: chunk,
		headers: { 'Content-Type': 'application/octet-stream' },
	});
}

export async function completeDriveUpload(id: string): Promise<{ fileId: string }> {
	return api<{ fileId: string }>(`/api/v1/drive/uploads/${id}/complete`, { method: 'POST' });
}

export async function cancelDriveUpload(id: string): Promise<void> {
	await api<void>(`/api/v1/drive/uploads/${id}`, { method: 'DELETE' });
}

// ---- Legacy multipart upload (kept for backward compatibility) ----

export async function uploadDriveFile(file: File): Promise<{ id: string; name: string }> {
	const fd = new FormData();
	fd.append('file', file);
	return api<{ id: string; name: string }>('/api/v1/drive/upload', {
		method: 'POST',
		body: fd,
	});
}

export function uploadDriveFileWithProgress(
	file: File,
	onProgress: (pct: number) => void,
	signal?: AbortSignal,
): Promise<{ id: string; name: string }> {
	return new Promise((resolve, reject) => {
		const fd = new FormData();
		fd.append('file', file);
		const xhr = new XMLHttpRequest();
		xhr.upload.onprogress = (e) => {
			if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
		};
		xhr.onload = () => {
			if (xhr.status >= 200 && xhr.status < 300) {
				try {
					const body = JSON.parse(xhr.responseText);
					resolve((body?.data ?? body) as { id: string; name: string });
				} catch {
					reject(new Error(m.parse_failed()));
				}
			} else {
				let msg = m.upload_failed_status({ status: xhr.status });
				try {
					const body = JSON.parse(xhr.responseText);
					if (body?.error) msg = body.error;
				} catch { /* ignore */ }
				reject(new Error(msg));
			}
		};
		xhr.onerror = () => reject(new Error(m.network_error()));
		xhr.onabort = () => reject(new DOMException('Aborted', 'AbortError'));
		const token = getAccessToken() ?? '';
		xhr.open('POST', '/api/v1/drive/upload');
		if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
		if (signal) signal.addEventListener('abort', () => xhr.abort());
		xhr.send(fd);
	});
}

// ---- Chunked upload driver ----

export type DriveChunkProgress = {
	uploaded: number;
	total: number;
};

export type DriveUploadOptions = {
	onProgress?: (p: DriveChunkProgress) => void;
	signal?: AbortSignal;
};

/**
 * Upload a file using the chunked upload protocol:
 * 1. If file ≤ 200 MB: compute SHA-256 client-side → check dedup
 *    - If deduped → claim existing file, return immediately
 * 2. Otherwise: skip client-side hash (server computes it)
 * 3. Init session → upload chunks → complete
 */
export async function driveChunkedUpload(
	file: File,
	mimeType: string,
	opts: DriveUploadOptions = {},
	parentId?: string | null,
): Promise<{ fileId: string }> {
	if (file.size <= MAX_HASH_SIZE) {
		const hash = await computeSHA256(file);
		const check = await driveCheckHash(hash);
		if (check.exists && check.fileId) {
			const result = await driveClaimHash(hash, file.name, mimeType, file.size, parentId);
			return result;
		}
	}

	const session = await initDriveUpload(file.name, mimeType, file.size, parentId);
	return uploadChunks(file, session.id, 0, opts);
}

/** Resume a previously paused upload from the server's last known offset. */
export async function resumeDriveUpload(
	file: File,
	sessionId: string,
	opts: DriveUploadOptions = {},
): Promise<{ fileId: string }> {
	const session = await getDriveSession(sessionId);
	return uploadChunks(file, sessionId, session.receivedBytes, opts);
}

export async function uploadChunks(
	file: File,
	sessionId: string,
	startOffset: number,
	opts: DriveUploadOptions = {},
): Promise<{ fileId: string }> {
	if (get(configError)) throw new Error(m.config_unavailable());
	const chunkSize = getChunkSize();
	if (chunkSize === null) throw new Error('config not loaded');
	let offset = startOffset;

	opts.onProgress?.({ uploaded: offset, total: file.size });

	while (offset < file.size) {
		if (opts.signal?.aborted) throw new DOMException('Aborted', 'AbortError');
		const end = Math.min(offset + chunkSize, file.size);
		const chunk = file.slice(offset, end);

		let lastErr: unknown;
		for (let attempt = 0; attempt < 3; attempt++) {
			try {
				const result = await uploadDriveChunk(sessionId, offset, chunk);
				offset = result.receivedBytes;
				lastErr = null;
				break;
			} catch (e) {
				lastErr = e;
				if (opts.signal?.aborted) throw new DOMException('Aborted', 'AbortError');
				if (attempt < 2) await new Promise((r) => setTimeout(r, 1000 * (attempt + 1)));
			}
		}
		if (lastErr) throw lastErr;

		opts.onProgress?.({ uploaded: offset, total: file.size });
	}

	return completeDriveUpload(sessionId);
}

// ---- Helper functions ----

export async function renameDriveFile(id: string, newName: string): Promise<DriveFile> {
	return api<DriveFile>(`/api/v1/drive/${id}`, {
		method: 'PATCH',
		body: JSON.stringify({ name: newName }),
	});
}

export async function deleteDriveFile(id: string): Promise<void> {
	await api<void>(`/api/v1/drive/${id}`, { method: 'DELETE' });
}

function authedUrl(path: string): string {
	const token = getAccessToken() ?? '';
	return `${path}?access_token=${encodeURIComponent(token)}`;
}

export function getDownloadUrl(id: string): string {
	return authedUrl(`/api/v1/drive/${id}/download`);
}

export function getPreviewUrl(id: string): string {
	return authedUrl(`/api/v1/drive/${id}/preview`);
}



