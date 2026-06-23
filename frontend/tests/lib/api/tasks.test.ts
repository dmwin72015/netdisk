import { describe, it, expect, vi } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { listTasks, getTask, deleteTask, computeFileSHA256, uploadFile } from '$lib/api/tasks';
import { computeSHA256 } from '$lib/upload-hash';

// ── listTasks ──────────────────────────────────────────────────────

describe('listTasks', () => {
	it('uses default limit and offset', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { items: [], total: 0 } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await listTasks();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=20');
		expect(url).toContain('offset=0');
	});

	it('passes custom limit and offset', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { items: [], total: 0 } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await listTasks(5, 10);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=5');
		expect(url).toContain('offset=10');
	});
});

// ── getTask ────────────────────────────────────────────────────────

describe('getTask', () => {
	it('fetches task by id', async () => {
		const task = { id: 't1', original_name: 'test.mp4', status: 'completed', progress: 100 };
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: task }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getTask('t1');
		expect(result).toEqual(task);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/tasks/t1');
	});
});

// ── deleteTask ─────────────────────────────────────────────────────

describe('deleteTask', () => {
	it('sends DELETE request', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(null, { status: 204 })
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		await deleteTask('t1');

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.method).toBe('DELETE');
	});
});

// ── computeFileSHA256 ──────────────────────────────────────────────

describe('computeFileSHA256', () => {
	it('produces same result as computeSHA256 for small file', async () => {
		const data = 'hello world';
		const file = new File([new TextEncoder().encode(data)], 'test.txt');
		const expected = await computeSHA256(file);
		const actual = await computeFileSHA256(file);
		expect(actual).toBe(expected);
	});

	it('handles file larger than one chunk (8 MiB)', async () => {
		// Create a 10 MiB file
		const buf = new Uint8Array(10 * 1024 * 1024);
		for (let i = 0; i < buf.length; i++) buf[i] = i % 256;
		const file = new File([buf], 'big.bin');

		const hash = await computeFileSHA256(file);
		expect(hash).toMatch(/^[a-f0-9]{64}$/);

		// Verify it matches computeSHA256 (which reads whole file at once)
		const expected = await computeSHA256(file);
		expect(hash).toBe(expected);
	});

	it('returns valid hex string for empty file', async () => {
		const file = new File([], 'empty.bin');
		const hash = await computeFileSHA256(file);
		expect(hash).toMatch(/^[a-f0-9]{64}$/);
	});
});

// ── uploadFile ─────────────────────────────────────────────────────

describe('uploadFile', () => {
	it('includes X-File-SHA256 header when sha256 is provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { taskId: 'task-1' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const file = new File(['content'], 'test.mp4', { type: 'video/mp4' });
		const result = await uploadFile(file, 'abc123sha256');

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('X-File-SHA256')).toBe('abc123sha256');
		expect(result.taskId).toBe('task-1');
	});

	it('omits X-File-SHA256 header when sha256 is not provided', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { taskId: 'task-2' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const file = new File(['content'], 'test.mp4', { type: 'video/mp4' });
		await uploadFile(file);

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('X-File-SHA256')).toBeNull();
	});

	it('sends FormData body with file appended', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { taskId: 'task-3' } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const file = new File(['content'], 'test.mp4', { type: 'video/mp4' });
		await uploadFile(file);

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.body).toBeInstanceOf(FormData);
		// FormData should contain the file
		const fd = init.body as FormData;
		expect(fd.get('file')).toBe(file);
	});
});
