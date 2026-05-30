import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { addToLibrary, listMedia, getMediaItem, removeFromLibrary, getHLSUrl } from './media';

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

// ── addToLibrary ──────────────────────────────────────────────────

describe('addToLibrary', () => {
	it('posts fileSlug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			mediaSlug: 'media-1',
			transcodeSlug: 'tc-1',
			transcodeStatus: 'pending',
			transcodeReused: false,
		}));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await addToLibrary('file-abc');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/media/items');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ fileSlug: 'file-abc' });
		expect(result.mediaSlug).toBe('media-1');
		expect(result.transcodeReused).toBe(false);
	});
});

// ── listMedia ─────────────────────────────────────────────────────

describe('listMedia', () => {
	it('uses default pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listMedia();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/media/items?');
		expect(url).toContain('page=1');
		expect(url).toContain('pageSize=50');
	});

	it('uses custom pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listMedia(3, 20);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('page=3');
		expect(url).toContain('pageSize=20');
	});
});

// ── getMediaItem ──────────────────────────────────────────────────

describe('getMediaItem', () => {
	it('fetches item by slug', async () => {
		const item = {
			mediaSlug: 'media-1',
			fileName: 'video.mp4',
			status: 'done',
			progress: 100,
			durationSec: 120,
			errorMsg: null,
			posterUrl: '/poster.jpg',
			playUrl: '/play.m3u8',
			createdAt: '2025-01-01T00:00:00Z',
		};
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(item));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getMediaItem('media-1');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/media/items/media-1');
		expect(result).toEqual(item);
	});
});

// ── removeFromLibrary ─────────────────────────────────────────────

describe('removeFromLibrary', () => {
	it('sends DELETE to item slug', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await removeFromLibrary('media-1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/media/items/media-1');
		expect(init.method).toBe('DELETE');
	});
});

// ── getHLSUrl ─────────────────────────────────────────────────────

describe('getHLSUrl', () => {
	it('builds correct HLS URL', () => {
		expect(getHLSUrl('media-1', 'video.m3u8')).toBe('/api/v1/media/hls/media-1/video.m3u8');
	});

	it('handles different slugs and filenames', () => {
		expect(getHLSUrl('abc-123', 'master.m3u8')).toBe('/api/v1/media/hls/abc-123/master.m3u8');
	});
});
