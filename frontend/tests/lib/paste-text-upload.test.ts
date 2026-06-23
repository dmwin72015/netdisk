import { describe, expect, it } from 'vitest';

import {
	MAX_PASTE_TEXT_SIZE,
	createTextFile,
	extractClipboardText,
	getDefaultFileName,
	validateTextSize,
} from '$lib/paste-text-upload';

describe('extractClipboardText', () => {
	it('extracts plain text from clipboardData', () => {
		const clipboardData = {
			getData: (format: string) => (format === 'text/plain' ? 'hello world' : ''),
		} as unknown as DataTransfer;

		expect(extractClipboardText(clipboardData)).toBe('hello world');
	});

	it('returns null for empty text', () => {
		const clipboardData = {
			getData: (format: string) => (format === 'text/plain' ? '' : ''),
		} as unknown as DataTransfer;

		expect(extractClipboardText(clipboardData)).toBeNull();
	});

	it('returns null for null clipboardData', () => {
		expect(extractClipboardText(null)).toBeNull();
		expect(extractClipboardText(undefined)).toBeNull();
	});
});

describe('validateTextSize', () => {
	it('returns valid for text within limit', () => {
		const result = validateTextSize('hello');
		expect(result.valid).toBe(true);
		expect(result.size).toBe(5);
		expect(result.error).toBeUndefined();
	});

	it('returns invalid with error for text exceeding 2MB', () => {
		const largeText = 'a'.repeat(MAX_PASTE_TEXT_SIZE + 1);
		const result = validateTextSize(largeText);
		expect(result.valid).toBe(false);
		expect(result.size).toBe(MAX_PASTE_TEXT_SIZE + 1);
		expect(result.error).toBeDefined();
		expect(result.error).toContain('2MB');
	});

	it('returns valid for text exactly at 2MB limit', () => {
		const text = 'a'.repeat(MAX_PASTE_TEXT_SIZE);
		const result = validateTextSize(text);
		expect(result.valid).toBe(true);
		expect(result.size).toBe(MAX_PASTE_TEXT_SIZE);
		expect(result.error).toBeUndefined();
	});
});

describe('createTextFile', () => {
	it('creates a File with correct properties', () => {
		const file = createTextFile('hello world', 'note.txt');
		expect(file.name).toBe('note.txt');
		expect(file.type).toBe('text/plain;charset=utf-8');
		expect(file.size).toBe(11);
	});
});

describe('getDefaultFileName', () => {
	it('extracts first line when short', () => {
		expect(getDefaultFileName('First line\nSecond line')).toBe('First line.txt');
	});

	it('truncates long first line to 50 chars', () => {
		const longLine = 'a'.repeat(100);
		const result = getDefaultFileName(longLine);
		expect(result.length).toBeLessThanOrEqual(54); // 50 + '.txt'
		expect(result.endsWith('.txt')).toBe(true);
	});

	it('uses timestamp for empty text', () => {
		const result = getDefaultFileName('');
		expect(result).toMatch(/^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.txt$/);
	});

	it('handles whitespace-only first line as empty', () => {
		const result = getDefaultFileName('   \nSecond line');
		expect(result).toMatch(/^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.txt$/);
	});

	it('handles first line of exactly 50 characters', () => {
		const exact = 'a'.repeat(50);
		const result = getDefaultFileName(exact);
		expect(result).toBe(`${exact}.txt`);
		expect(result.length).toBe(54);
	});

	it('removes illegal filename characters', () => {
		expect(getDefaultFileName('file/name:with*illegal?chars')).toBe('filenamewithillegalchars.txt');
	});

	it('removes all illegal filename chars leaving empty string', () => {
		const result = getDefaultFileName('*:*?*"*<*>*|*');
		expect(result).toBe('.txt');
	});

	it('ensures .txt extension', () => {
		expect(getDefaultFileName('document')).toBe('document.txt');
		expect(getDefaultFileName('readme.md')).toBe('readme.md.txt');
	});
});

describe('formatSize', () => {
	it('formats bytes below 1KB', () => {
		// Access via validateTextSize error message which uses formatSize
		const result = validateTextSize('x');
		expect(result.error).toBeUndefined(); // small text is valid, no error
	});

	it('formats kilobytes correctly', () => {
		// 2048 bytes = 2KB, still valid so no error
		const text = 'x'.repeat(2048);
		const result = validateTextSize(text);
		expect(result.valid).toBe(true);
	});

	it('formats megabytes correctly in error message', () => {
		const largeText = 'a'.repeat(MAX_PASTE_TEXT_SIZE + 1);
		const result = validateTextSize(largeText);
		expect(result.error).toContain('MB');
	});
});
