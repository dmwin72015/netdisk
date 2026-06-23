import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { computeSHA256, computeSHA256Chunked } from './upload-hash';

// ── Worker mock ──────────────────────────────────────────────────────

let mockWorker: {
	onmessage: ((e: { data: Record<string, unknown> }) => void) | null;
	onerror: ((e: { error: Error }) => void) | null;
	postMessage(msg: unknown, transfer?: Transferable[]): void;
	terminate(): void;
};
const sentMessages: unknown[] = [];

beforeEach(() => {
	sentMessages.length = 0;
	mockWorker = {
		onmessage: null,
		onerror: null,
		postMessage(msg: unknown) { sentMessages.push(msg); },
		terminate() {},
	};
	// Use a function constructor so `new Worker(...)` works
	vi.stubGlobal('Worker', vi.fn().mockImplementation(function MockWorkerCtor() {
		return mockWorker;
	}));
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── computeSHA256 ────────────────────────────────────────────────────

describe('computeSHA256', () => {
	it('computes SHA-256 of a file', async () => {
		const file = new File(['hello'], 'test.txt', { type: 'text/plain' });
		const hash = await computeSHA256(file);
		// SHA-256 of "hello"
		expect(hash).toBe('2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824');
	});

	it('computes SHA-256 of empty content', async () => {
		const file = new File([], 'empty.txt');
		const hash = await computeSHA256(file);
		// SHA-256 of empty string
		expect(hash).toBe('e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855');
	});
});

// ── computeSHA256Chunked ─────────────────────────────────────────────

describe('computeSHA256Chunked', () => {
	function triggerWorkerMessage(data: Record<string, unknown>) {
		if (mockWorker.onmessage) {
			mockWorker.onmessage({ data } as MessageEvent);
		}
	}

	function triggerWorkerError(error: Error) {
		if (mockWorker.onerror) {
			mockWorker.onerror({ error } as Event);
		}
	}

	it('sends file chunks and resolves with hash from worker', async () => {
		const content = 'x'.repeat(1024);
		const file = new File([content], 'test.bin', { type: 'application/octet-stream' });

		const resultPromise = computeSHA256Chunked(file, {});

		// Simulate worker processing the chunks and responding
		triggerWorkerMessage({ type: 'pre_hash', hash: 'prehash123' });
		triggerWorkerMessage({ type: 'complete', hash: 'finalhash456', totalChunks: 1 });

		const result = await resultPromise;
		expect(result.preHash).toBe('prehash123');
		expect(result.hash).toBe('finalhash456');
		expect(result.totalChunks).toBe(1);
	});

	it('sends init with skipPreHash when option is set', async () => {
		const file = new File(['x'], 'test.bin');
		const resultPromise = computeSHA256Chunked(file, {}, undefined, { skipPreHash: true });

		// Verify the 'init' message was sent
		const initMsg = sentMessages.find(m => (m as Record<string, unknown>).type === 'init');
		expect(initMsg).toBeDefined();
		expect((initMsg as Record<string, unknown>).skipPreHash).toBe(true);

		triggerWorkerMessage({ type: 'complete', hash: 'hash1', totalChunks: 1 });
		await resultPromise;
	});

	it('sends correct chunk messages with data', async () => {
		const content = 'x'.repeat(8192); // 8KB
		const file = new File([content], 'test.bin');

		const resultPromise = computeSHA256Chunked(file, {});

		// Give the async IIFE a tick to start reading the file
		await new Promise(r => setTimeout(r, 0));

		// Find chunk messages
		const chunkMessages = sentMessages.filter(m => (m as Record<string, unknown>).type === 'chunk');
		expect(chunkMessages.length).toBeGreaterThan(0);

		// Each chunk message should have index and data
		for (const msg of chunkMessages) {
			const record = msg as Record<string, unknown>;
			expect(record).toHaveProperty('index');
			expect(record).toHaveProperty('data');
			expect((record.data as ArrayBuffer).byteLength).toBeGreaterThan(0);
		}

		// Verify 'done' message
		const doneMsg = sentMessages.find(m => (m as Record<string, unknown>).type === 'done');
		expect(doneMsg).toBeDefined();

		triggerWorkerMessage({ type: 'complete', hash: 'hash1', totalChunks: chunkMessages.length });
		await resultPromise;
	});

	it('calls onPreHash callback when pre_hash is received', async () => {
		const file = new File(['x'], 'test.bin');
		const preHashes: string[] = [];
		const resultPromise = computeSHA256Chunked(file, {
			onPreHash: (hash) => preHashes.push(hash),
		});

		triggerWorkerMessage({ type: 'pre_hash', hash: 'the-pre-hash' });
		triggerWorkerMessage({ type: 'complete', hash: 'the-full-hash', totalChunks: 1 });

		await resultPromise;
		expect(preHashes).toEqual(['the-pre-hash']);
	});

	it('calls onProgress callback when progress is received', async () => {
		const file = new File(['x'], 'test.bin');
		const progresses: number[] = [];
		const resultPromise = computeSHA256Chunked(file, {
			onProgress: (pct) => progresses.push(pct),
		});

		triggerWorkerMessage({ type: 'progress', percent: 50 });
		triggerWorkerMessage({ type: 'progress', percent: 100 });
		triggerWorkerMessage({ type: 'complete', hash: 'h', totalChunks: 1 });

		await resultPromise;
		expect(progresses).toContain(50);
		expect(progresses).toContain(100);
	});

	it('rejects when worker errors', async () => {
		const file = new File(['x'], 'test.bin');
		const resultPromise = computeSHA256Chunked(file, {});

		triggerWorkerError(new Error('worker crashed'));
		await expect(resultPromise).rejects.toThrow('SHA-256 computation failed');
	});

	it('terminates worker on error', async () => {
		const terminateSpy = vi.spyOn(mockWorker, 'terminate');
		const file = new File(['x'], 'test.bin');
		const resultPromise = computeSHA256Chunked(file, {});

		triggerWorkerError(new Error('crash'));
		try { await resultPromise; } catch { /* expected */ }

		expect(terminateSpy).toHaveBeenCalledTimes(1);
	});

	it('returns after second completion message (idempotent)', async () => {
		const file = new File(['x'], 'test.bin');
		const resultPromise = computeSHA256Chunked(file, {});

		triggerWorkerMessage({ type: 'complete', hash: 'hash1', totalChunks: 1 });
		triggerWorkerMessage({ type: 'complete', hash: 'hash2', totalChunks: 1 }); // duplicate

		const result = await resultPromise;
		expect(result.hash).toBe('hash1'); // first response wins (settled flag)
	});
});
