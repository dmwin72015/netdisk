import { describe, expect, it } from 'vitest';
import { getUploadChunkCount, shouldUseSmallFileFastPath } from '$lib/upload-strategy';

describe('upload strategy', () => {
	it('uses fast path for files that fit in one chunk', () => {
		expect(shouldUseSmallFileFastPath(1024, 4096)).toBe(true);
		expect(shouldUseSmallFileFastPath(4096, 4096)).toBe(true);
	});

	it('keeps normal upload flow for multi-chunk files', () => {
		expect(shouldUseSmallFileFastPath(4097, 4096)).toBe(false);
		expect(getUploadChunkCount(4097, 4096)).toBe(2);
	});

	it('does not fast-path invalid sizes', () => {
		expect(shouldUseSmallFileFastPath(0, 4096)).toBe(false);
		expect(shouldUseSmallFileFastPath(1024, 0)).toBe(false);
	});
});
