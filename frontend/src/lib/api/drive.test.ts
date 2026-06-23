import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({
	parse_failed: () => 'parse failed',
	upload_failed_status: ({ status }: { status: number }) => `upload failed: ${status}`,
	network_error: () => 'network error',
}));
vi.mock('$lib/upload-hash', () => ({
	computeSHA256: vi.fn(),
}));

import { computeSHA256 } from '$lib/upload-hash';
import { getDownloadUrl, getPreviewUrl, listDrive, driveCheckUpload, initDriveUpload, uploadDriveFileWithProgress } from './drive';
import { clientConfig } from '$lib/stores/config';
import { fmtSize } from '$lib/utils/format';
import { ApiError } from './client';

const mockedComputeSHA256 = computeSHA256 as ReturnType<typeof vi.fn>;

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

// ── Additional API function tests ────────────────────────────────────

import {
	driveStats, createDriveDir, getDriveFile, getDriveAncestors,
	driveCheckHash, driveClaimHash, listDriveSessions, getDriveSession,
	uploadDriveChunk, completeDriveUpload, cancelDriveUpload,
	uploadDriveFile, driveChunkedUpload, resumeDriveUpload, renameDriveFile, deleteDriveFile,
} from './drive';

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

describe('driveStats', () => {
	it('fetches drive stats', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({
			usedBytes: 500, baseBytes: 500, memberBonusBytes: 0, packBytes: 0, totalBytes: 1024,
		}));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await driveStats();
		expect(result).toEqual({ usedBytes: 500, baseBytes: 500, memberBonusBytes: 0, packBytes: 0, totalBytes: 1024 });
		expect(fetchSpy.mock.calls[0][0]).toBe('/api/v1/drive/stats');
	});
});

describe('createDriveDir', () => {
	it('posts name and parentId', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 'dir-1', name: 'New', isDir: true }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await createDriveDir('New', 'parent-1');
		expect(result.id).toBe('dir-1');

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ name: 'New', parentId: 'parent-1' });
	});

	it('omits parentId when not provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 'dir-2', name: 'Root', isDir: true }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await createDriveDir('Root');
		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.parentId).toBeUndefined();
	});
});

describe('getDriveFile', () => {
	it('fetches file by id', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 'f1', name: 'a.txt', size: 100, isDir: false }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getDriveFile('f1');
		expect(result.name).toBe('a.txt');
		expect(fetchSpy.mock.calls[0][0]).toBe('/api/v1/drive/f1');
	});
});

describe('getDriveAncestors', () => {
	it('fetches ancestor chain', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse([
			{ id: 'root', name: 'Root', isDir: true },
			{ id: 'sub', name: 'Sub', isDir: true },
		]));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getDriveAncestors('sub');
		expect(result).toHaveLength(2);
		expect(fetchSpy.mock.calls[0][0]).toBe('/api/v1/drive/sub/ancestors');
	});
});

describe('driveCheckHash', () => {
	it('posts sha256 for dedup check', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ exists: true, fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await driveCheckHash('abc123');
		expect(result.exists).toBe(true);
		expect(result.fileId).toBe('f1');
		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ sha256: 'abc123' });
	});
});

describe('driveClaimHash', () => {
	it('claims an existing file by hash', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await driveClaimHash('hash1', 'doc.pdf', 'application/pdf', 1024, 'dir-1');
		expect(result.fileId).toBe('f1');
		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ sha256: 'hash1', originalName: 'doc.pdf', mimeType: 'application/pdf', fileSize: 1024, parentId: 'dir-1' });
	});

	it('omits parentId when null', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await driveClaimHash('hash1', 'doc.pdf', 'application/pdf', 1024, null);
		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body.parentId).toBeUndefined();
	});
});

describe('listDriveSessions', () => {
	it('returns list of upload sessions', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [{ id: 's1', name: 'f.mp4' }], total: 1 }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await listDriveSessions();
		expect(result).toHaveLength(1);
		expect(result[0].id).toBe('s1');
	});
});

describe('getDriveSession', () => {
	it('fetches session by id', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 's1', name: 'f.mp4', totalSize: 5000, receivedBytes: 1000 }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getDriveSession('s1');
		expect(result.receivedBytes).toBe(1000);
		expect(fetchSpy.mock.calls[0][0]).toBe('/api/v1/drive/uploads/s1');
	});
});

describe('uploadDriveChunk', () => {
	it('PATCHes chunk to session', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 's1', receivedBytes: 4096 }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const blob = new Blob(['data']);
		const result = await uploadDriveChunk('s1', 0, blob);
		expect(result.receivedBytes).toBe(4096);
		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/drive/uploads/s1?offset=0');
		expect(fetchSpy.mock.calls[0][1].method).toBe('PATCH');
	});
});

describe('completeDriveUpload', () => {
	it('POSTs to complete endpoint', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await completeDriveUpload('s1');
		expect(result.fileId).toBe('f1');
		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/drive/uploads/s1/complete');
		expect(fetchSpy.mock.calls[0][1].method).toBe('POST');
	});
});

describe('cancelDriveUpload', () => {
	it('DELETEs session', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await cancelDriveUpload('s1');
		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/drive/uploads/s1');
		expect(fetchSpy.mock.calls[0][1].method).toBe('DELETE');
	});
});

describe('uploadDriveFile', () => {
	it('uploads via FormData', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 'f1', name: 'a.txt' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const file = new File(['content'], 'a.txt');
		const result = await uploadDriveFile(file);
		expect(result.id).toBe('f1');
		const body = fetchSpy.mock.calls[0][1].body;
		expect(body).toBeInstanceOf(FormData);
	});
});

describe('uploadDriveFileWithProgress', () => {
	it('resolves on success', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: { id: 'f1', name: 'a.txt' } }), { status: 200 }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => 'tok' });

		const netInstance = {
			upload: { onprogress: null },
			onload: null as (() => void) | null,
			onerror: null,
			onabort: null,
			status: 200,
			responseText: JSON.stringify({ data: { id: 'f1', name: 'a.txt' } }),
			setRequestHeader() {},
			open() {},
			send() { netInstance.onload?.(); },
		};
		vi.stubGlobal('XMLHttpRequest', function SuccessXHR() { return netInstance; });

		const file = new File(['x'.repeat(100)], 'a.txt');
		const result = await uploadDriveFileWithProgress(file, () => {});
		expect(result.id).toBe('f1');
	});

	it('rejects on HTTP error with server message', async () => {
		vi.stubGlobal('fetch', undefined);
		vi.stubGlobal('localStorage', { getItem: () => 'tok' });
		const instance = {
			upload: { onprogress: null },
			onload: null,
			onerror: null,
			onabort: null,
			status: 400,
			responseText: JSON.stringify({ error: 'bad request' }),
			setRequestHeader() {},
			open() {},
			send() { instance.onload?.(); },
		};
		vi.stubGlobal('XMLHttpRequest', vi.fn().mockImplementation(function XHRCtor2() { return instance; }));

		const file = new File(['x'], 'a.txt');
		await expect(uploadDriveFileWithProgress(file, () => {})).rejects.toThrow('bad request');
	});

	it('rejects with network error', async () => {
		vi.stubGlobal('fetch', undefined);
		vi.stubGlobal('localStorage', { getItem: () => 'tok' });
		const netInstance = {
			upload: { onprogress: null },
			onload: null,
			onerror: null as (() => void) | null,
			onabort: null,
			setRequestHeader() {},
			open() {},
			send() { netInstance.onerror?.(); },
		};
		vi.stubGlobal('XMLHttpRequest', function NetXHR() { return netInstance; });

		const file = new File(['x'], 'a.txt');
		await expect(uploadDriveFileWithProgress(file, () => {})).rejects.toThrow('network error');
	});
});

describe('driveChunkedUpload / resumeDriveUpload', () => {
	beforeEach(() => {
		mockedComputeSHA256.mockClear();
	});

	it('small file dedup: computes hash and claims', async () => {
		mockedComputeSHA256.mockResolvedValue('hash1');
		const fetchSpy = vi.fn()
			.mockResolvedValueOnce(jsonResponse({ exists: true, fileId: 'f1' }))
			.mockResolvedValueOnce(jsonResponse({ fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });
		// uploadChunks reads chunk size from config store
		clientConfig.set({ device: 'web', configs: { 'upload.chunkSize': 4194304 } });

		const driveModule = await import('./drive');
		const file = new File(['x'], 'small.txt');
		const result = await driveModule.driveChunkedUpload(file, 'text/plain', {}, null);
		expect(result.fileId).toBe('f1');
		expect(mockedComputeSHA256).toHaveBeenCalled();
		expect(fetchSpy).toHaveBeenCalledTimes(2);
	});

	it('small file miss: falls through to chunked upload', async () => {
		mockedComputeSHA256.mockResolvedValue('hash1');
		const fetchSpy = vi.fn()
			.mockResolvedValueOnce(jsonResponse({ exists: false }))
			.mockResolvedValueOnce(jsonResponse({ id: 's1', name: 'f.txt', totalSize: 100, receivedBytes: 0 }))
			.mockResolvedValueOnce(jsonResponse({ id: 's1', receivedBytes: 100 }))
			.mockResolvedValueOnce(jsonResponse({ fileId: 'f2' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const driveModule = await import('./drive');
		const file = new File(['x'.repeat(100)], 'f.txt');
		const result = await driveModule.driveChunkedUpload(file, 'text/plain', {}, null);
		expect(result.fileId).toBe('f2');
	});

	it('resume uses receivedBytes from session', async () => {
		const fetchSpy = vi.fn()
			.mockResolvedValueOnce(jsonResponse({ id: 's1', name: 'f.txt', totalSize: 1000, receivedBytes: 500 }))
			.mockResolvedValueOnce(jsonResponse({ id: 's1', receivedBytes: 1000 }))
			.mockResolvedValueOnce(jsonResponse({ fileId: 'f1' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const driveModule = await import('./drive');
		const file = new File(['x'.repeat(1000)], 'f.txt');
		const result = await driveModule.resumeDriveUpload(file, 's1', {});
		expect(result.fileId).toBe('f1');
	});
});

describe('renameDriveFile', () => {
	it('PATCHes file name', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: 'f1', name: 'new.txt' }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await renameDriveFile('f1', 'new.txt');
		expect(result.name).toBe('new.txt');
		expect(fetchSpy.mock.calls[0][1].method).toBe('PATCH');
		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ name: 'new.txt' });
	});
});

describe('deleteDriveFile', () => {
	it('DELETEs file', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await deleteDriveFile('f1');
		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/drive/f1');
		expect(fetchSpy.mock.calls[0][1].method).toBe('DELETE');
	});
});
