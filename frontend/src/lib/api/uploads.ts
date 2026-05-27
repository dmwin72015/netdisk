// Chunked / resumable upload client. Server is the single source of truth —
// progress is read back via Status() / list() so the user can resume from any
// device. The client only orchestrates which chunk to PATCH next.

import { api, ApiError } from './client';
import * as m from '$lib/paraglide/messages';

export type UploadSession = {
	id: string;
	filename: string;
	total_size: number;
	received_bytes: number;
	created_at: number;
	updated_at: number;
};

export type ChunkResult = { id: string; received_bytes: number };

/** Default chunk size — large enough to amortize round-trips, small enough
 *  that one chunk failing isn't catastrophic. */
export const DEFAULT_CHUNK_BYTES = 5 * 1024 * 1024;

export async function listUploadSessions(): Promise<UploadSession[]> {
	const data = await api<{ items: UploadSession[] }>('/api/v1/uploads');
	return data.items;
}

export async function initUpload(filename: string, totalSize: number, sha256?: string): Promise<UploadSession & { task_id?: string; status?: string }> {
	const headers: Record<string, string> = {};
	if (sha256) headers['X-File-SHA256'] = sha256;
	return api<UploadSession & { task_id?: string; status?: string }>('/api/v1/uploads', {
		method: 'POST',
		body: JSON.stringify({ filename, total_size: totalSize }),
		headers
	});
}

export type CheckHashResult = {
	exists: boolean;
	deduped?: boolean;
	transcoded?: boolean;
	task_id?: string;
	status?: string;
};

export type ClaimResult = {
	task_id: string;
	status: string;
};

/**
 * Pre-upload dedup check. Before uploading any bytes, call this to see if the
 * file (identified by SHA-256) already exists in the system — either as the
 * current user's own task or as a cross-user file_store entry.
 */
export async function checkHash(sha256: string): Promise<CheckHashResult> {
	return api<CheckHashResult>('/api/v1/check-sha256', {
		method: 'POST',
		body: JSON.stringify({ sha256 }),
	});
}

/**
 * Claim an existing file_store entry by hash, creating a task for the current
 * user without uploading any data. Only works when the file has already been
 * transcoded (hls_output_dir set).
 */
export async function claimHash(
	sha256: string,
	originalName: string,
	fileSize: number
): Promise<ClaimResult> {
	return api<ClaimResult>('/api/v1/claim', {
		method: 'POST',
		body: JSON.stringify({ sha256, original_name: originalName, file_size: fileSize }),
	});
}

const CHUNK_SIZE = 8 * 1024 * 1024; // 8 MiB per read chunk

/**
 * Compute SHA-256 of a File using the Web Crypto API.
 * Uses file.arrayBuffer() directly which is more memory-efficient than
 * the manual chunking approach (single buffer vs chunk array + concat).
 * For files larger than 200 MB, callers should skip client-side hashing
 * and let the server compute it.
 */
export async function computeSHA256(file: File): Promise<string> {
	const buf = await file.arrayBuffer();
	const digest = await crypto.subtle.digest('SHA-256', buf);
	return Array.from(new Uint8Array(digest))
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
}

const FEATURE_CHUNK = 1024 * 1024; // 1 MB

/**
 * Compute a feature hash: SHA-256 of (head 1MB || middle 1MB || tail 1MB).
 * Only reads 3 MB total regardless of file size — sub-second for large files.
 * For files <= 3 MB, computes full SHA-256 (which doubles as the feature hash).
 */
export async function computeFeatureHash(file: File): Promise<{ featureHash: string; isFullHash: boolean }> {
	if (file.size <= 3 * FEATURE_CHUNK) {
		const fullHash = await computeSHA256(file);
		return { featureHash: fullHash, isFullHash: true };
	}

	const head = await file.slice(0, FEATURE_CHUNK).arrayBuffer();
	const midStart = Math.floor((file.size - FEATURE_CHUNK) / 2);
	const middle = await file.slice(midStart, midStart + FEATURE_CHUNK).arrayBuffer();
	const tail = await file.slice(file.size - FEATURE_CHUNK).arrayBuffer();

	const combined = new Uint8Array(3 * FEATURE_CHUNK);
	combined.set(new Uint8Array(head), 0);
	combined.set(new Uint8Array(middle), FEATURE_CHUNK);
	combined.set(new Uint8Array(tail), 2 * FEATURE_CHUNK);

	const digest = await crypto.subtle.digest('SHA-256', combined);
	const featureHash = Array.from(new Uint8Array(digest))
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');

	return { featureHash, isFullHash: false };
}

export type QuickCheckResult = {
	found: boolean;
	sha256?: string;
	transcoded?: boolean;
	status?: string;
};

/**
 * Stage 1 quick pre-filter: send file_size + feature_hash to the server
 * to quickly check if a matching file exists. Only returns the candidate's
 * full SHA-256 for Stage 2 verification.
 */
export async function quickCheck(fileSize: number, featureHash: string): Promise<QuickCheckResult> {
	return api<QuickCheckResult>('/api/v1/quick-check', {
		method: 'POST',
		body: JSON.stringify({ file_size: fileSize, feature_hash: featureHash }),
	});
}

export async function getUploadSession(id: string): Promise<UploadSession> {
	return api<UploadSession>(`/api/v1/uploads/${id}`);
}

export async function uploadChunk(
	id: string,
	offset: number,
	chunk: Blob
): Promise<ChunkResult> {
	// fetch loses the upload-progress stream for binary bodies, but since
	// the bytes-per-chunk are small the per-PATCH progress is meaningful by
	// itself (each PATCH ≈ one progress jump on the bar).
	return api<ChunkResult>(`/api/v1/uploads/${id}?offset=${offset}`, {
		method: 'PATCH',
		body: chunk,
		headers: { 'Content-Type': 'application/octet-stream' }
	});
}

export async function completeUpload(id: string): Promise<{ task_id: string }> {
	return api<{ task_id: string }>(`/api/v1/uploads/${id}/complete`, { method: 'POST' });
}

export async function cancelUpload(id: string): Promise<void> {
	await api<void>(`/api/v1/uploads/${id}`, { method: 'DELETE' });
}

export type ChunkProgress = {
	uploaded: number;
	total: number;
};

export type DriverOptions = {
	chunkSize?: number;
	onProgress?: (p: ChunkProgress) => void;
	signal?: AbortSignal;
};

/**
 * Stream a File into an existing session, starting from `startOffset`. Used by
 * both the "fresh upload" and the "resume" flows — they only differ in where
 * the session and starting offset come from.
 */
export async function driveUpload(
	sessionId: string,
	file: File,
	startOffset: number,
	opts: DriverOptions = {}
): Promise<{ task_id: string }> {
	const chunkSize = opts.chunkSize ?? DEFAULT_CHUNK_BYTES;
	let offset = startOffset;
	opts.onProgress?.({ uploaded: offset, total: file.size });

	while (offset < file.size) {
		if (opts.signal?.aborted) throw new DOMException('Aborted', 'AbortError');
		const end = Math.min(offset + chunkSize, file.size);
		const chunk = file.slice(offset, end);
		const result = await uploadChunk(sessionId, offset, chunk);
		offset = result.received_bytes;
		opts.onProgress?.({ uploaded: offset, total: file.size });
	}
	return completeUpload(sessionId);
}

/** Throws when the picked file doesn't match the in-progress session by name
 *  or size. The user must pick the same file they were uploading before. */
export function assertFileMatchesSession(file: File, session: UploadSession): void {
	if (file.size !== session.total_size) {
		throw new ApiError(
			m.file_size_mismatch({ fileSize: file.size, sessionSize: session.total_size }),
			400
		);
	}
	if (file.name !== session.filename) {
		// Different name with matching size is suspicious; don't silently accept it.
		throw new ApiError(
			m.file_name_mismatch({ fileName: file.name, sessionName: session.filename }),
			400
		);
	}
}
