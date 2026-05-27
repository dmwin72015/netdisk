import { api } from './client';

export type PreCheckResponse = {
	status: 'SUSPECT_HIT' | 'NOT_FOUND';
};

export type ChallengeResponse = {
	status: 'NOT_FOUND' | 'CHALLENGE';
	challenge_offset: number;
	challenge_token: string;
};

export type VerifyResponse = {
	status: 'HIT' | 'MISS';
	physical_file_slug?: string;
};

export type InitResponse = {
	upload_slug: string;
	total_chunks: number;
	chunk_size: number;
	completed_chunks: number[];
};

export type StatusResponse = {
	status: string;
	physical_file_slug?: string;
	error?: string;
};

export async function preCheck(preHash: string, fileSize: number) {
	return api<PreCheckResponse>('/api/v1/upload/pre-check', {
		method: 'POST',
		body: JSON.stringify({ pre_hash: preHash, file_size: fileSize })
	});
}

export async function requestChallenge(fileHash: string) {
	return api<ChallengeResponse>('/api/v1/upload/request-challenge', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash })
	});
}

export async function verify(fileHash: string, proofCode: string) {
	return api<VerifyResponse>('/api/v1/upload/verify', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, proof_code: proofCode })
	});
}

export async function initUpload(fileHash: string, preHash: string, fileSize: number, mimeType: string) {
	return api<InitResponse>('/api/v1/upload/init', {
		method: 'POST',
		body: JSON.stringify({ file_hash: fileHash, pre_hash: preHash, file_size: fileSize, mime_type: mimeType })
	});
}

export async function updateHash(uploadSlug: string, fileHash: string, preHash?: string) {
	return api('/api/v1/upload/update-hash', {
		method: 'POST',
		body: JSON.stringify({ upload_slug: uploadSlug, file_hash: fileHash, pre_hash: preHash ?? '' })
	});
}

export async function uploadChunk(uploadSlug: string, chunkIndex: number, data: ArrayBuffer) {
	const form = new FormData();
	form.append('upload_slug', uploadSlug);
	form.append('chunk_index', String(chunkIndex));
	form.append('chunk_data', new Blob([data]));
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
	return api<StatusResponse>(`/api/v1/upload/${uploadSlug}/status`);
}
