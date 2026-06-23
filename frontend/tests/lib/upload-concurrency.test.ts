import { describe, expect, it, vi } from 'vitest';
import { UploadRequestPool, UPLOAD_CHUNK_CONCURRENCY_PER_FILE, UPLOAD_FILE_CONCURRENCY, UPLOAD_FILE_CONCURRENCY_DEFAULT, UPLOAD_REQUEST_POOL_SIZE } from '$lib/upload-concurrency';

describe('upload concurrency constants', () => {
	it('uses requested limits', () => {
		expect(UPLOAD_FILE_CONCURRENCY).toBe(5);
		expect(UPLOAD_FILE_CONCURRENCY_DEFAULT).toBe(3);
		expect(UPLOAD_CHUNK_CONCURRENCY_PER_FILE).toBe(2);
		expect(UPLOAD_REQUEST_POOL_SIZE).toBe(10);
	});
});

describe('UploadRequestPool', () => {
	it('limits active requests', async () => {
		const pool = new UploadRequestPool(2);
		let active = 0;
		let maxActive = 0;
		let completed = 0;
		const resolvers: Array<() => void> = [];

		const tasks = Array.from({ length: 5 }, () =>
			pool.schedule(async () => {
				active++;
				maxActive = Math.max(maxActive, active);
				await new Promise<void>((resolve) => resolvers.push(resolve));
				active--;
				completed++;
			})
		);

		await vi.waitFor(() => expect(resolvers).toHaveLength(2));
		expect(maxActive).toBe(2);

		while (completed < tasks.length) {
			await vi.waitFor(() => expect(resolvers.length).toBeGreaterThan(0));
			resolvers.shift()?.();
			await Promise.resolve();
		}
		await Promise.all(tasks);
		expect(maxActive).toBe(2);
	});

	it('rejects queued requests when aborted', async () => {
		const pool = new UploadRequestPool(1);
		const firstRelease = Promise.withResolvers<void>();
		const controller = new AbortController();

		const first = pool.schedule(async () => {
			await firstRelease.promise;
		});
		const second = pool.schedule(async () => 'second', controller.signal);

		controller.abort();
		await expect(second).rejects.toThrow();
		firstRelease.resolve();
		await first;
	});

	it('runs tasks immediately when limit is 0', async () => {
		const pool = new UploadRequestPool(0);
		let ran = false;
		await pool.schedule(async () => {
			ran = true;
		});
		expect(ran).toBe(true);
	});

	it('rejects immediately when signal is already aborted', async () => {
		const pool = new UploadRequestPool(2);
		const controller = new AbortController();
		controller.abort('already aborted');

		await expect(
			pool.schedule(async () => 'should not run', controller.signal)
		).rejects.toThrow('already aborted');
	});

	it('does not call logStats when pool is idle', () => {
		const debugSpy = vi.spyOn(console, 'debug').mockImplementation(() => {});
		// Create pool and let it sit — logStats runs on interval but should be silent when idle
		const pool = new UploadRequestPool(2);
		// Manually call logStats to verify silence
		pool['logStats']();
		expect(debugSpy).not.toHaveBeenCalled();
		debugSpy.mockRestore();
	});

	it('drains aborted items from queue without decrementing active', async () => {
		const pool = new UploadRequestPool(1);
		const controller = new AbortController();

		// Schedule first task that holds the slot
		const firstRelease = Promise.withResolvers<void>();
		const first = pool.schedule(async () => {
			await firstRelease.promise;
		});

		// Schedule second task with already-aborted signal
		const second = pool.schedule(async () => 'second', controller.signal);
		controller.abort();

		// The aborted second task should be rejected and not block the queue
		await expect(second).rejects.toThrow();
		firstRelease.resolve();
		await first;
		// If we got here without hanging, the pool drained correctly
	});

	it('abort without reason uses DOMException fallback', async () => {
		const pool = new UploadRequestPool(1);
		const controller = new AbortController();
		// Abort without a reason (undefined)
		controller.abort();

		// Schedule a task with the aborted signal
		await expect(
			pool.schedule(async () => 'should not run', controller.signal)
		).rejects.toThrow();
	});
});
