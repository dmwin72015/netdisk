import { describe, expect, it, vi } from 'vitest';
import { UploadRequestPool, UPLOAD_CHUNK_CONCURRENCY_PER_FILE, UPLOAD_FILE_CONCURRENCY, UPLOAD_FILE_CONCURRENCY_DEFAULT, UPLOAD_REQUEST_POOL_SIZE } from './upload-concurrency';

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
});
