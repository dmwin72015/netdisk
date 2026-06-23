import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { listUploadTasks, retryUploadTask, deleteUploadTask, deleteUploadTasks } from '$lib/api/upload-tasks';

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

// ── listUploadTasks ───────────────────────────────────────────────

describe('listUploadTasks', () => {
	it('uses default limit and offset', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, limit: 20, offset: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listUploadTasks();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/upload/tasks?');
		expect(url).toContain('limit=20');
		expect(url).toContain('offset=0');
	});

	it('uses custom limit and offset', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, limit: 5, offset: 10 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listUploadTasks(5, 10);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=5');
		expect(url).toContain('offset=10');
	});

	it('includes date filters when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, limit: 20, offset: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listUploadTasks(20, 0, '2025-01-01', '2025-12-31');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('start_date=2025-01-01');
		expect(url).toContain('end_date=2025-12-31');
	});

	it('omits date filters when not provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, limit: 20, offset: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listUploadTasks();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).not.toContain('start_date');
		expect(url).not.toContain('end_date');
	});

	it('includes status filter when provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0, limit: 20, offset: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await listUploadTasks(20, 0, undefined, undefined, 'failed');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('status=failed');
	});

	it('returns the response data', async () => {
		const responseData = {
			items: [
				{
					slug: 'task-1',
					fileName: 'video.mp4',
					fileSize: 5000,
					mimeType: 'video/mp4',
					status: 'completed',
					errorMsg: '',
					totalChunks: 5,
					receivedBytes: 5000,
					parentSlug: 'dir-1',
					parentName: '工作文档',
					createdAt: '2025-01-01T00:00:00Z',
					updatedAt: '2025-01-01T00:01:00Z',
				},
			],
			total: 1,
			limit: 20,
			offset: 0,
		};
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(responseData));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await listUploadTasks();
		expect(result.items).toHaveLength(1);
		expect(result.items[0].slug).toBe('task-1');
		expect(result.items[0].parentName).toBe('工作文档');
		expect(result.total).toBe(1);
	});
});

// ── retryUploadTask ───────────────────────────────────────────────

describe('retryUploadTask', () => {
	it('sends POST to retry endpoint', async () => {
		const responseData = {
			uploadSlug: 'task-1',
			totalChunks: 5,
			chunkSize: 1024,
			completedChunks: [0, 1],
		};
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(responseData));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await retryUploadTask('task-1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/tasks/task-1/retry');
		expect(init.method).toBe('POST');
		expect(result).toEqual(responseData);
	});

	it('returns completed chunks for resume', async () => {
		const responseData = {
			uploadSlug: 'task-2',
			totalChunks: 10,
			chunkSize: 2048,
			completedChunks: [0, 1, 2, 3],
		};
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(responseData));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await retryUploadTask('task-2');
		expect(result.completedChunks).toEqual([0, 1, 2, 3]);
		expect(result.totalChunks).toBe(10);
	});
});

// ── deleteUploadTask ───────────────────────────────────────────────

describe('deleteUploadTask', () => {
	it('sends DELETE request for a single task', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await deleteUploadTask('task-1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/tasks/task-1');
		expect(init.method).toBe('DELETE');
	});
});

// ── deleteUploadTasks ──────────────────────────────────────────────

describe('deleteUploadTasks', () => {
	it('sends DELETE with slugs array to batch delete', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await deleteUploadTasks(['task-1', 'task-2']);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/upload/tasks');
		expect(init.method).toBe('DELETE');
		expect(JSON.parse(init.body)).toEqual({ slugs: ['task-1', 'task-2'] });
	});
});
