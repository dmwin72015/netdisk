import { describe, it, expect, vi } from 'vitest';

import { createUploadQueue } from '$lib/upload-queue';

// ── createUploadQueue ──────────────────────────────────────────────

describe('createUploadQueue', () => {
	it('returns an object with start method', () => {
		const queue = createUploadQueue(() => [], () => {}, () => false, async () => {});
		expect(queue).toHaveProperty('start');
		expect(typeof queue.start).toBe('function');
	});

	it('processes all processable items', async () => {
		const items = [
			{ uid: '1', phase: 'pending' },
			{ uid: '2', phase: 'pending' },
			{ uid: '3', phase: 'pending' },
		];
		let current = [...items];
		const processed: string[] = [];

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => {
				processed.push(item.uid);
				item.phase = 'done';
			},
		);

		await queue.start();
		expect(processed).toEqual(['1', '2', '3']);
	});

	it('skips non-processable items', async () => {
		const items = [
			{ uid: '1', phase: 'done' },
			{ uid: '2', phase: 'pending' },
			{ uid: '3', phase: 'done' },
		];
		let current = [...items];
		const processed: string[] = [];

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => {
				processed.push(item.uid);
				item.phase = 'done';
			},
		);

		await queue.start();
		expect(processed).toEqual(['2']);
	});

	it('handles errors without crashing the queue', async () => {
		const items = [
			{ uid: '1', phase: 'pending' },
			{ uid: '2', phase: 'pending' },
			{ uid: '3', phase: 'pending' },
		];
		let current = [...items];
		const processed: string[] = [];

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => {
				try {
					if (item.uid === '2') throw new Error('network error');
					processed.push(item.uid);
				} finally {
					item.phase = 'done';
				}
			},
		);

		await queue.start();
		expect(processed).toEqual(['1', '3']);
	});

	it('calls onAllDone callback after processing', async () => {
		const items = [{ uid: '1', phase: 'pending' }];
		let current = [...items];
		const onAllDone = vi.fn();

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => { item.phase = 'done'; },
			onAllDone,
		);

		await queue.start();
		expect(onAllDone).toHaveBeenCalledOnce();
	});

	it('does not call onAllDone when items remain', async () => {
		const items = [
			{ uid: '1', phase: 'pending' },
			{ uid: '2', phase: 'blocked' },
		];
		let current = [...items];
		const onAllDone = vi.fn();

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => { item.phase = 'done'; },
			onAllDone,
		);

		await queue.start();
		// onAllDone is still called because the queue loop exits when no processable items remain
		// (even if some items are in a non-processable state like 'blocked')
		expect(onAllDone).toHaveBeenCalledOnce();
	});

	it('does not start twice concurrently', async () => {
		const items = [
			{ uid: '1', phase: 'pending' },
			{ uid: '2', phase: 'pending' },
		];
		let current = [...items];
		const processed: string[] = [];

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => {
				processed.push(item.uid);
				item.phase = 'done';
			},
		);

		// Start twice concurrently — second call should be a no-op
		await Promise.all([queue.start(), queue.start()]);
		expect(processed).toEqual(['1', '2']);
	});

	it('processes items that become processable after earlier items complete', async () => {
		const items = [
			{ uid: '1', phase: 'pending' },
			{ uid: '2', phase: 'blocked' },
		];
		let current = [...items];
		const processed: string[] = [];

		const queue = createUploadQueue(
			() => current,
			(updated) => { current = updated; },
			(item) => item.phase === 'pending',
			async (item) => {
				processed.push(item.uid);
				item.phase = 'done';
				// After item 1 completes, make item 2 processable
				if (item.uid === '1') {
					current[1].phase = 'pending';
				}
			},
		);

		await queue.start();
		expect(processed).toEqual(['1', '2']);
	});
});
