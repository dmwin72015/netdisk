import { describe, it, expect } from 'vitest';
import { cn } from '$lib/utils/cn';

describe('cn', () => {
	it('joins truthy class names', () => {
		expect(cn('a', 'b', 'c')).toBe('a b c');
	});

	it('filters out falsy values', () => {
		expect(cn('a', false, null, undefined, 'b')).toBe('a b');
		expect(cn(false, null, undefined)).toBe('');
	});

	it('resolves tailwind conflicts via twMerge', () => {
		// twMerge keeps the later conflicting class
		expect(cn('px-2', 'px-4')).toBe('px-4');
		expect(cn('text-red', 'text-blue')).toBe('text-blue');
	});

	it('preserves non-conflicting classes', () => {
		expect(cn('px-2', 'py-4')).toBe('px-2 py-4');
	});
});
