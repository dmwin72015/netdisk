import { describe, expect, it } from 'vitest';
import { getUploadedBytesFromCompletedChunks, getUploadProgress, normalizeCompletedChunks } from '$lib/upload-resume';

describe('normalizeCompletedChunks', () => {
	it('keeps valid unique chunk indexes only', () => {
		const chunks = normalizeCompletedChunks([0, 2, 2, -1, 5, 1.5], 4);
		expect([...chunks].sort((a, b) => a - b)).toEqual([0, 2]);
	});
});

describe('getUploadedBytesFromCompletedChunks', () => {
	it('counts non-contiguous completed chunks by exact size', () => {
		const completed = new Set([0, 2]);
		expect(getUploadedBytesFromCompletedChunks(completed, 3, 4, 10)).toBe(6);
	});
});

describe('getUploadProgress', () => {
	it('rounds progress and caps at 100', () => {
		expect(getUploadProgress(5, 10)).toBe(50);
		expect(getUploadProgress(12, 10)).toBe(100);
		expect(getUploadProgress(5, 0)).toBe(0);
	});
});
