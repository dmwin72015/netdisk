import { api } from './client';

export type PreCheckResponse = {
	status: 'SUSPECT_HIT' | 'NOT_FOUND';
};

export type ChallengeResponse = {
	status: 'NOT_FOUND' | 'CHALLENGE';
	challengeOffset: number;
	challengeToken: string;
};

export type ExistingFileRef = {
	fileName: string;
	path: string;
};

export type VerifyResponse = {
	status: 'HIT' | 'MISS';
	physicalFileSlug?: string;
	existingFiles?: ExistingFileRef[];
};

export type InitResponse = {
	uploadSlug: string;
	totalChunks: number;
	chunkSize: number;
	completedChunks: number[];
};

export type StatusResponse = {
	status: string;
	physicalFileSlug?: string;
	error?: string;
};

export async function preCheck(preHash: string, fileSize: number) {
	return api<PreCheckResponse>('/api/v1/upload/pre-check', {
		method: 'POST',
		body: JSON.stringify({ preHash, fileSize })
	});
}

export async function requestChallenge(fileHash: string) {
	return api<ChallengeResponse>('/api/v1/upload/request-challenge', {
		method: 'POST',
		body: JSON.stringify({ fileHash })
	});
}

export async function verify(fileHash: string, proofCode: string) {
	return api<VerifyResponse>('/api/v1/upload/verify', {
		method: 'POST',
		body: JSON.stringify({ fileHash, proofCode })
	});
}

export async function initUpload(fileHash: string, preHash: string, fileSize: number, mimeType: string, fileName?: string) {
	return api<InitResponse>('/api/v1/upload/init', {
		method: 'POST',
		body: JSON.stringify({ fileHash, preHash, fileSize, mimeType, fileName: fileName ?? '' })
	});
}

export async function updateHash(uploadSlug: string, fileHash: string, preHash?: string) {
	return api('/api/v1/upload/update-hash', {
		method: 'POST',
		body: JSON.stringify({ uploadSlug, fileHash, preHash: preHash ?? '' })
	});
}

export async function uploadChunk(uploadSlug: string, chunkIndex: number, data: ArrayBuffer) {
	const form = new FormData();
	form.append('uploadSlug', uploadSlug);
	form.append('chunkIndex', String(chunkIndex));
	form.append('chunkData', new Blob([data]));
	return api('/api/v1/upload/chunk', {
		method: 'POST',
		body: form,
		headers: {} // Let browser set Content-Type with boundary
	});
}

export async function completeUpload(uploadSlug: string) {
	return api('/api/v1/upload/complete', {
		method: 'POST',
		body: JSON.stringify({ uploadSlug })
	});
}

export async function getUploadStatus(uploadSlug: string) {
	return api<StatusResponse>(`/api/v1/upload/${uploadSlug}/status`);
}
