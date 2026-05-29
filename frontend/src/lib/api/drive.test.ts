import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { getDownloadUrl, getPreviewUrl, listDrive, driveCheckUpload, initDriveUpload } from './drive';
import { fmtSize } from '$lib/utils/format';
import { ApiError } from './client';

// ── fmtSize ────────────────────────────────────────────────────────

describe('fmtSize', () => {
	it('formats bytes', () => {
		expect(fmtSize(0)).toBe('0 B');
		expect(fmtSize(512)).toBe('512 B');
		expect(fmtSize(1023)).toBe('1023 B');
	});

	it('formats kilobytes', () => {
		expect(fmtSize(1024)).toBe('1.0 KB');
		expect(fmtSize(1536)).toBe('1.5 KB');
		expect(fmtSize(1024 * 1024 - 1)).toBe('1024.0 KB');
	});

	it('formats megabytes', () => {
		expect(fmtSize(1024 * 1024)).toBe('1.0 MB');
		expect(fmtSize(5.5 * 1024 * 1024)).toBe('5.5 MB');
	});

	it('formats gigabytes', () => {
		expect(fmtSize(1024 * 1024 * 1024)).toBe('1.00 GB');
		expect(fmtSize(2.5 * 1024 * 1024 * 1024)).toBe('2.50 GB');
	});
});

// ── getDownloadUrl / getPreviewUrl ──────────────────────────────────

describe('url helpers', () => {
	beforeEach(() => {
		vi.stubGlobal('localStorage', {
			getItem: (k: string) => (k === 'nd.access' ? 'tok123' : null),
		});
	});

	it('getDownloadUrl builds correct URL', () => {
		const url = getDownloadUrl('abc');
		expect(url).toBe('/api/v1/drive/abc/download?access_token=tok123');
	});

	it('getPreviewUrl builds correct URL', () => {
		const url = getPreviewUrl('xyz');
		expect(url).toBe('/api/v1/drive/xyz/preview?access_token=tok123');
	});

	it('encodes special characters in token', () => {
		vi.stubGlobal('localStorage', {
			getItem: () => 'a&b=c',
		});
		const url = getDownloadUrl('id');
		expect(url).toContain(encodeURIComponent('a&b=c'));
	});
});

// ── listDrive ──────────────────────────────────────────────────────

describe('listDrive', () => {
	it('builds query with default params', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { items: [], total: 0 } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await listDrive();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=50');
		expect(url).toContain('offset=0');
		expect(url).not.toContain('q=');
	});

	it('includes q and parentId when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { items: [], total: 0 } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await listDrive(10, 5, 'hello world', 'dir-123');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=10');
		expect(url).toContain('offset=5');
		expect(url).toContain('q=hello%20world');
		expect(url).toContain('parentId=dir-123');
	});
});

// ── driveCheckUpload ───────────────────────────────────────────────

describe('driveCheckUpload', () => {
	it('posts sha256 and metadata', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { status: 'none' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await driveCheckUpload('abc123', 1024, 'test.mp4', 'video/mp4', 'dir-1');

		expect(result).toEqual({ status: 'none' });

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({
			sha256: 'abc123',
			fileSize: 1024,
			fileName: 'test.mp4',
			mimeType: 'video/mp4',
			parentId: 'dir-1',
		});
	});

	it('returns full status with file_id', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { status: 'full', fileId: 'file-1' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await driveCheckUpload('abc', 100, 'f', 'video/mp4');
		expect(result.status).toBe('full');
		expect(result.fileId).toBe('file-1');
	});
});

// ── initDriveUpload ────────────────────────────────────────────────

describe('initDriveUpload', () => {
	it('includes sha256 in request body', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { id: 'sess-1', name: 'test.mp4' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await initDriveUpload('test.mp4', 'video/mp4', 1024, null, 'hash123');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.sha256).toBe('hash123');
		expect(body.filename).toBe('test.mp4');
		expect(body.mimeType).toBe('video/mp4');
		expect(body.totalSize).toBe(1024);
	});

	it('omits sha256 when not provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { id: 'sess-2' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await initDriveUpload('test.mp4', 'video/mp4', 1024);

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.sha256).toBeUndefined();
	});
});
