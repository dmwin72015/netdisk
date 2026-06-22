import { describe, expect, it } from 'vitest';

import { extractClipboardFiles, filterPasteFiles, isEditablePasteTarget } from './paste-upload';
import { extractClipboardText, validateTextSize } from './paste-text-upload';

function makeFile(name: string, type = 'text/plain') {
	return new File(['content'], name, { type });
}

function clipboardWithFiles(files: File[]) {
	return {
		files: {
			length: files.length,
			item: (index: number) => files[index] ?? null,
			[Symbol.iterator]: function* () {
				yield* files;
			},
		},
		items: [],
	} as unknown as DataTransfer;
}

describe('extractClipboardFiles', () => {
	it('extracts files from clipboardData.files', () => {
		const pasted = [makeFile('a.png', 'image/png'), makeFile('b.pdf', 'application/pdf')];

		expect(extractClipboardFiles(clipboardWithFiles(pasted))).toEqual(pasted);
	});

	it('extracts file items when clipboardData.files is empty', () => {
		const pasted = makeFile('photo.jpg', 'image/jpeg');
		const clipboard = {
			files: { length: 0, item: () => null, [Symbol.iterator]: function* () {} },
			items: [
				{ kind: 'file', getAsFile: () => pasted },
				{ kind: 'string', getAsFile: () => null },
			],
		} as unknown as DataTransfer;

		expect(extractClipboardFiles(clipboard)).toEqual([pasted]);
	});

	it('returns an empty array for missing clipboard data', () => {
		expect(extractClipboardFiles(null)).toEqual([]);
		expect(extractClipboardFiles(undefined)).toEqual([]);
	});
});

describe('filterPasteFiles', () => {
	it('splits accepted and rejected files', () => {
		const mp4 = makeFile('movie.mp4', 'video/mp4');
		const png = makeFile('image.png', 'image/png');

		expect(filterPasteFiles([mp4, png], (candidate) => candidate.type.startsWith('video/'))).toEqual({
			accepted: [mp4],
			rejected: [png],
		});
	});

	it('accepts every file when no predicate is provided', () => {
		const files = [makeFile('a.txt'), makeFile('b.txt')];

		expect(filterPasteFiles(files)).toEqual({ accepted: files, rejected: [] });
	});
});

describe('isEditablePasteTarget', () => {
	function target(tagName: string, isContentEditable = false) {
		return { tagName, isContentEditable } as unknown as EventTarget;
	}

	it('detects text inputs, textareas, selects, and contenteditable elements', () => {
		expect(isEditablePasteTarget(target('input'))).toBe(true);
		expect(isEditablePasteTarget(target('textarea'))).toBe(true);
		expect(isEditablePasteTarget(target('select'))).toBe(true);
		expect(isEditablePasteTarget(target('div', true))).toBe(true);
	});

	it('returns false for non-editable elements and null targets', () => {
		expect(isEditablePasteTarget(target('div'))).toBe(false);
		expect(isEditablePasteTarget(null)).toBe(false);
	});
});

describe('PasteUploadProvider - text handling', () => {
	it('should prioritize files over text when both present', () => {
		const file = new File([''], 'test.txt');
		const clipboard = {
			files: [file],
			getData: () => 'text content',
		} as unknown as DataTransfer;

		const result = extractClipboardFiles(clipboard);
		expect(result).toHaveLength(1);
	});

	it('should validate text size before showing dialog', () => {
		const text = 'a'.repeat(2 * 1024 * 1024 + 1);
		const result = validateTextSize(text);
		expect(result.valid).toBe(false);
		expect(result.error).toContain('2MB');
	});
});
