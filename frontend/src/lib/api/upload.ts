import { api } from './client';
import { uploadRequestPool } from '$lib/upload-concurrency';

const TIMEOUTS = {
	preCheck: 30_000,
	challenge: 30_000,
	verify: 30_000,
	init: 30_000,
	chunk: 300_000,
	updateHash: 30_000,
	complete: 300_000,
	status: 30_000,
} as const;

function withTimeout(timeoutMs: number, signal?: AbortSignal) {
	const timeoutSignal = AbortSignal.timeout(timeoutMs);
	if (!signal) return timeoutSignal;

	const controller = new AbortController();
	const abortFrom = (source: AbortSignal) => {
		if (!controller.signal.aborted) controller.abort(source.reason);
	};

	if (signal.aborted) {
		abortFrom(signal);
		return controller.signal;
	}
	if (timeoutSignal.aborted) {
		abortFrom(timeoutSignal);
		return controller.signal;
	}

	signal.addEventListener('abort', () => abortFrom(signal), { once: true });
	timeoutSignal.addEventListener('abort', () => abortFrom(timeoutSignal), { once: true });
	return controller.signal;
}

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

export type CompleteResponse = {
	status: string;
	physicalFileSlug?: string;
};

export type StatusResponse = {
	status: string;
	physicalFileSlug?: string;
	error?: string;

	taskType?: string;
	sourceUrl?: string;
	fileName?: string;
	fileSize?: number;
	receivedBytes?: number;
};

export type FileDedupResponse = {
	exists: boolean;
	physicalFileSlug?: string;
};

export async function checkFileDedup(fileHash: string) {
	return api<FileDedupResponse>('/api/v1/upload/dedup-by-hash', {
		method: 'POST',
		body: JSON.stringify({ fileHash }),
		signal: AbortSignal.timeout(30_000),
	});
}

export async function preCheck(preHash: string, fileSize: number) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.preCheck);
		console.debug(`[upload-api] preCheck: size=${fileSize}`);
		return api<PreCheckResponse>('/api/v1/upload/pre-check', {
		method: 'POST',
		body: JSON.stringify({ preHash, fileSize }),
		signal,
		});
	});
}

export async function requestChallenge(fileHash: string) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.challenge);
		console.debug(`[upload-api] requestChallenge: hash=${fileHash.slice(0, 8)}...`);
		return api<ChallengeResponse>('/api/v1/upload/request-challenge', {
		method: 'POST',
		body: JSON.stringify({ fileHash }),
		signal,
		});
	});
}

export async function verify(fileHash: string, proofCode: string) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.verify);
		console.debug(`[upload-api] verify: hash=${fileHash.slice(0, 8)}..., proofLength=${proofCode.length}`);
		return api<VerifyResponse>('/api/v1/upload/verify', {
		method: 'POST',
		body: JSON.stringify({ fileHash, proofCode }),
		signal,
		});
	});
}

export async function initUpload(fileHash: string, preHash: string, fileSize: number, mimeType: string, fileName?: string, parentSlug?: string) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.init);
		console.debug(`[upload-api] initUpload: name=${fileName} size=${fileSize} hash=${fileHash.slice(0, 8)}... parent=${parentSlug ?? ''}`);
		return api<InitResponse>('/api/v1/upload/init', {
		method: 'POST',
		body: JSON.stringify({ fileHash, preHash, fileSize, mimeType, fileName: fileName ?? '', parentSlug: parentSlug ?? '' }),
		signal,
		});
	});
}

export async function updateHash(uploadSlug: string, fileHash: string, preHash?: string) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.updateHash);
		console.debug(`[upload-api] updateHash: slug=${uploadSlug} hash=${fileHash.slice(0, 8)}...`);
		return api('/api/v1/upload/update-hash', {
		method: 'POST',
		body: JSON.stringify({ uploadSlug, fileHash, preHash: preHash ?? '' }),
		signal,
		});
	});
}

export async function uploadChunk(uploadSlug: string, chunkIndex: number, data: ArrayBuffer, signal?: AbortSignal) {
	return uploadRequestPool.schedule(() => {
		const requestSignal = withTimeout(TIMEOUTS.chunk, signal);
		const form = new FormData();
		form.append('uploadSlug', uploadSlug);
		form.append('chunkIndex', String(chunkIndex));
		form.append('chunkData', new Blob([data]));
		console.debug(`[upload-api] uploadChunk: slug=${uploadSlug} chunk=${chunkIndex} size=${data.byteLength}`);
		return api('/api/v1/upload/chunk', {
		method: 'POST',
		body: form,
		headers: {},
		signal: requestSignal,
		});
	}, signal);
}

export async function completeUpload(uploadSlug: string): Promise<CompleteResponse> {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.complete);
		console.debug(`[upload-api] completeUpload: slug=${uploadSlug}`);
		return api<CompleteResponse>('/api/v1/upload/complete', {
		method: 'POST',
		body: JSON.stringify({ uploadSlug }),
		signal,
		});
	});
}

export type URLUploadResponse = {
  taskSlug: string;
  status: string;
};

export async function urlUpload(url: string, fileName?: string, parentSlug?: string) {
  return api<URLUploadResponse>('/api/v1/upload/from-url', {
    method: 'POST',
    body: JSON.stringify({ url, fileName: fileName ?? '', parentSlug: parentSlug ?? '' }),
    signal: AbortSignal.timeout(30_000),
  });
}

export async function getUploadStatus(uploadSlug: string) {
	return uploadRequestPool.schedule(() => {
		const signal = AbortSignal.timeout(TIMEOUTS.status);
		return api<StatusResponse>(`/api/v1/upload/${uploadSlug}/status`, {
		signal,
		});
	});
}
