import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { listVideos, getVideo, uploadThumbnail, captureFrameThumbnail } from './videos';

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

// ── listVideos ─────────────────────────────────────────────────────

describe('listVideos', () => {
	it('fetches with default pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listVideos();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/videos?');
		expect(url).toContain('limit=20');
		expect(url).toContain('offset=0');
	});

	it('uses custom pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listVideos(50, 10);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=50');
		expect(url).toContain('offset=10');
	});

	it('returns parsed data', async () => {
		const data = { items: [{ id: 'v1', originalName: 'test.mp4' }], total: 1, limit: 20, offset: 0 };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(data));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await listVideos();
		expect(result).toEqual(data);
	});
});

// ── getVideo ───────────────────────────────────────────────────────

describe('getVideo', () => {
	it('fetches video by id', async () => {
		const video = { id: 'v1', originalName: 'test.mp4', status: 'completed' };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(video));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getVideo('v1');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/videos/v1');
		expect(result).toEqual(video);
	});
});

// ── uploadThumbnail ────────────────────────────────────────────────

describe('uploadThumbnail', () => {
	it('posts FormData with file', async () => {
		const video = { id: 'v1', thumbnailUrl: '/thumb.jpg' };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(video));
		vi.stubGlobal('fetch', fetchSpy);

		const file = new File([new ArrayBuffer(100)], 'thumb.jpg', { type: 'image/jpeg' });
		const result = await uploadThumbnail('v1', file);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/videos/v1/thumbnail');
		expect(init.method).toBe('POST');
		expect(init.body).toBeInstanceOf(FormData);
		expect(result).toEqual(video);
	});
});

// ── captureFrameThumbnail ──────────────────────────────────────────

describe('captureFrameThumbnail', () => {
	it('posts to frame endpoint with encoded time', async () => {
		const video = { id: 'v1', thumbnailUrl: '/frame.jpg' };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(video));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await captureFrameThumbnail('v1', 12.5);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/videos/v1/thumbnail/frame?at=12.50');
		expect(init.method).toBe('POST');
		expect(result).toEqual(video);
	});

	it('encodes fractional seconds correctly', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({}));
		vi.stubGlobal('fetch', fetchSpy);

		await captureFrameThumbnail('v1', 0.1);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('at=0.10');
	});
});
