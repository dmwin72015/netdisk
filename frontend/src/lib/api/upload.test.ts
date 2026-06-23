import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import {
	preCheck,
	requestChallenge,
	verify,
	initUpload,
	updateHash,
	uploadChunk,
	completeUpload,
	getUploadStatus,
	checkFileDedup,
	urlUpload,
} from './upload';

// ── helpers ────────────────────────────────────────────────────────

let store: Record<string, string>;

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

beforeEach(() => {
	store = {};
	vi.stubGlobal('localStorage', {
		getItem: (k: string) => store[k] ?? null,
		setItem: (k: string, v: string) => { store[k] = v; },
		removeItem: (k: string) => { delete store[k]; },
		clear: () => { store = {}; },
	});
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── preCheck ──────────────────────────────────────────────────────

describe('preCheck', () => {
	it('posts preHash and fileSize', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ status: 'NOT_FOUND' }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await preCheck('hash123', 1024);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/pre-check');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ preHash: 'hash123', fileSize: 1024 });
		expect(result).toEqual({ status: 'NOT_FOUND' });
	});

	it('returns SUSPECT_HIT status', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ status: 'SUSPECT_HIT' }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await preCheck('abc', 500);
		expect(result.status).toBe('SUSPECT_HIT');
	});
});

// ── requestChallenge ──────────────────────────────────────────────

describe('requestChallenge', () => {
	it('posts fileHash', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			status: 'CHALLENGE',
			challengeOffset: 100,
			challengeToken: 'tok',
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await requestChallenge('sha256hash');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/request-challenge');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ fileHash: 'sha256hash' });
		expect(result.status).toBe('CHALLENGE');
		expect(result.challengeOffset).toBe(100);
		expect(result.challengeToken).toBe('tok');
	});
});

// ── verify ────────────────────────────────────────────────────────

describe('verify', () => {
	it('posts fileHash and proofCode', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			status: 'HIT',
			physicalFileSlug: 'existing-file',
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await verify('sha256hash', 'proof123');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/verify');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ fileHash: 'sha256hash', proofCode: 'proof123' });
		expect(result.status).toBe('HIT');
		expect(result.physicalFileSlug).toBe('existing-file');
	});

	it('returns MISS status', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ status: 'MISS' }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await verify('hash', 'proof');
		expect(result.status).toBe('MISS');
	});
});

// ── initUpload ────────────────────────────────────────────────────

describe('initUpload', () => {
	it('posts all required params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			uploadSlug: 'upload-1',
			totalChunks: 5,
			chunkSize: 1024,
			completedChunks: [],
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await initUpload('fileHash', 'preHash', 5000, 'video/mp4', 'test.mp4');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/init');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({
			fileHash: 'fileHash',
			preHash: 'preHash',
			fileSize: 5000,
			mimeType: 'video/mp4',
			fileName: 'test.mp4',
			parentSlug: '',
		});
		expect(result.uploadSlug).toBe('upload-1');
		expect(result.totalChunks).toBe(5);
	});

	it('defaults fileName to empty string when omitted', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			uploadSlug: 'upload-2',
			totalChunks: 1,
			chunkSize: 1024,
			completedChunks: [],
		}));
		vi.stubGlobal('fetch', fetchSpy);

		await initUpload('hash', 'pre', 100, 'text/plain');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.fileName).toBe('');
		expect(body.parentSlug).toBe('');
	});
});

// ── updateHash ────────────────────────────────────────────────────

describe('updateHash', () => {
	it('posts uploadSlug and fileHash', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await updateHash('upload-1', 'newHash');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/update-hash');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({
			uploadSlug: 'upload-1',
			fileHash: 'newHash',
			preHash: '',
		});
	});

	it('includes preHash when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await updateHash('upload-1', 'newHash', 'preHash123');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.preHash).toBe('preHash123');
	});
});

// ── uploadChunk ───────────────────────────────────────────────────

describe('uploadChunk', () => {
	it('sends FormData with chunk data', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		const data = new ArrayBuffer(100);
		await uploadChunk('upload-1', 0, data);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/chunk');
		expect(init.method).toBe('POST');
		expect(init.body).toBeInstanceOf(FormData);
		// Verify headers override (empty Content-Type to let browser set boundary)
		expect(init.headers.get('Content-Type')).toBeNull();
	});
});

// ── completeUpload ────────────────────────────────────────────────

describe('completeUpload', () => {
	it('posts uploadSlug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await completeUpload('upload-1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/complete');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ uploadSlug: 'upload-1' });
	});
});

// ── getUploadStatus ───────────────────────────────────────────────

describe('getUploadStatus', () => {
	it('fetches status by upload slug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			status: 'completed',
			physicalFileSlug: 'file-1',
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getUploadStatus('upload-1');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/upload/upload-1/status');
		expect(result.status).toBe('completed');
		expect(result.physicalFileSlug).toBe('file-1');
	});

	it('returns error status', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			status: 'failed',
			error: 'hash mismatch',
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getUploadStatus('upload-2');
		expect(result.status).toBe('failed');
		expect(result.error).toBe('hash mismatch');
	});
});

// ── checkFileDedup ───────────────────────────────────────────────────

describe('checkFileDedup', () => {
	it('posts fileHash to dedup endpoint', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ exists: true, physicalFileSlug: 'phys-1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await checkFileDedup('abc123');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/dedup-by-hash');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ fileHash: 'abc123' });
		expect(result.exists).toBe(true);
		expect(result.physicalFileSlug).toBe('phys-1');
	});

	it('returns exists=false when no match', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ exists: false }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await checkFileDedup('newhash');
		expect(result.exists).toBe(false);
	});
});

// ── urlUpload ───────────────────────────────────────────────────────

describe('urlUpload', () => {
	it('posts url and optional fileName/parentSlug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ taskSlug: 'task-1', status: 'pending' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await urlUpload('https://example.com/video.mp4', 'video.mp4', 'dir-1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/from-url');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ url: 'https://example.com/video.mp4', fileName: 'video.mp4', parentSlug: 'dir-1' });
		expect(result.taskSlug).toBe('task-1');
	});

	it('omits fileName and parentSlug when not provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ taskSlug: 'task-2', status: 'pending' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await urlUpload('https://example.com/file.zip');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.url).toBe('https://example.com/file.zip');
		expect(body.fileName).toBe('');
		expect(body.parentSlug).toBe('');
	});
});
