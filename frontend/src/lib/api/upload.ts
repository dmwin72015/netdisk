import { api } from './client';

export type PreCheckResponse = {
	status: 'HIT' | 'MISS';
	physical_file_slug?: string;
};

export type ChallengeResponse = {
	offset: number;
	challenge_token: string;
};

export type VerifyResponse = {
	status: 'OK' | 'MISMATCH';
	physical_file_slug?: string;
};

export type UploadTask = {
	upload_slug: string;
	file_hash: string;
	total_chunks: number;
	chunks_uploaded: number;
	status: string;
	physical_file_slug: string | null;
	created_at: string;
	expires_at: string;
};

export async function preCheck(fileHash: string, fileSize: number) {
	return api<PreCheckResponse>('/api/v1/upload/pre-check', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, file_size: fileSize })
	});
}

export async function requestChallenge(fileHash: string, fileSize: number, totalChunks: number) {
	return api<ChallengeResponse>('/api/v1/upload/request-challenge', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, file_size: fileSize, total_chunks: totalChunks })
	});
}

export async function verify(fileHash: string, fileSize: number, challengeToken: string, offset: number, sample: string) {
	return api<VerifyResponse>('/api/v1/upload/verify', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, file_size: fileSize, challenge_token: challengeToken, offset, sample })
	});
}

export async function initUpload(fileHash: string, fileSize: number, totalChunks: number, fileName: string) {
	return api<UploadTask>('/api/v1/upload/init', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, file_size: fileSize, total_chunks: totalChunks, file_name: fileName })
	});
}

export async function uploadChunk(uploadSlug: string, chunkIndex: number, data: ArrayBuffer) {
	const form = new FormData();
	form.append('upload_slug', uploadSlug);
	form.append('chunk_index', String(chunkIndex));
	form.append('chunk', new Blob([data]));
	return api('/api/v1/upload/chunk', {
		method: 'POST',
		body: form,
		headers: {} // Let browser set Content-Type with boundary
	});
}

export async function completeUpload(uploadSlug: string) {
	return api('/api/v1/upload/complete', {
		method: 'POST',
		body: JSON.stringify({ upload_slug: uploadSlug })
	});
}

export async function getUploadStatus(uploadSlug: string) {
	return api<UploadTask>(`/api/v1/upload/${uploadSlug}/status`);
}
