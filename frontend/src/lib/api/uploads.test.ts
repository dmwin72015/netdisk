import { describe, it, expect, vi } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({
	file_size_mismatch: ({ fileSize, sessionSize }: { fileSize: number; sessionSize: number }) =>
		`size mismatch: ${fileSize} vs ${sessionSize}`,
	file_name_mismatch: ({ fileName, sessionName }: { fileName: string; sessionName: string }) =>
		`name mismatch: ${fileName} vs ${sessionName}`,
}));

import { computeSHA256, assertFileMatchesSession, DEFAULT_CHUNK_BYTES } from './uploads';
import type { UploadSession } from './uploads';
import { ApiError } from './client';

// ── computeSHA256 ──────────────────────────────────────────────────

describe('computeSHA256', () => {
	it('returns hex digest of file contents', async () => {
		// "hello" → SHA-256 known value
		const file = new File([new TextEncoder().encode('hello')], 'hello.txt', { type: 'text/plain' });
		const hash = await computeSHA256(file);
		expect(hash).toBe('2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824');
	});

	it('returns same hash for same content', async () => {
		const a = new File(['test data'], 'a.bin');
		const b = new File(['test data'], 'b.bin');
		expect(await computeSHA256(a)).toBe(await computeSHA256(b));
	});

	it('returns different hash for different content', async () => {
		const a = new File(['content A'], 'a.bin');
		const b = new File(['content B'], 'b.bin');
		expect(await computeSHA256(a)).not.toBe(await computeSHA256(b));
	});

	it('handles empty file', async () => {
		const file = new File([], 'empty.bin');
		const hash = await computeSHA256(file);
		expect(hash).toMatch(/^[a-f0-9]{64}$/);
	});
});

// ── assertFileMatchesSession ───────────────────────────────────────

describe('assertFileMatchesSession', () => {
	const session: UploadSession = {
		id: 's1',
		filename: 'video.mp4',
		totalSize: 1024,
		receivedBytes: 0,
		createdAt: 0,
		updatedAt: 0,
	};

	it('passes when file matches session', () => {
		const file = new File([new ArrayBuffer(1024)], 'video.mp4');
		expect(() => assertFileMatchesSession(file, session)).not.toThrow();
	});

	it('throws ApiError(400) on size mismatch', () => {
		const file = new File([new ArrayBuffer(512)], 'video.mp4');
		expect(() => assertFileMatchesSession(file, session)).toThrow(ApiError);
		try {
			assertFileMatchesSession(file, session);
		} catch (e) {
			expect((e as ApiError).status).toBe(400);
			expect((e as ApiError).message).toContain('512');
		}
	});

	it('throws ApiError(400) on name mismatch', () => {
		const file = new File([new ArrayBuffer(1024)], 'other.mp4');
		expect(() => assertFileMatchesSession(file, session)).toThrow(ApiError);
		try {
			assertFileMatchesSession(file, session);
		} catch (e) {
			expect((e as ApiError).status).toBe(400);
			expect((e as ApiError).message).toContain('other.mp4');
		}
	});
});

// ── DEFAULT_CHUNK_BYTES ────────────────────────────────────────────

describe('DEFAULT_CHUNK_BYTES', () => {
	it('is 5 MiB', () => {
		expect(DEFAULT_CHUNK_BYTES).toBe(5 * 1024 * 1024);
	});
});
