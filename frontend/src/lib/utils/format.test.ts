import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({
	duration_seconds: () => '0s',
	duration_h_m: ({ h, m }: { h: number; m: number }) => `${h}h ${m}min`,
	duration_h: ({ h }: { h: number }) => `${h}h`,
	duration_m_s: ({ m, s }: { m: number; s: number }) => `${m}min ${s}s`,
	duration_m: ({ m }: { m: number }) => `${m}min`,
	duration_s: ({ s }: { s: number }) => `${s}s`,
	just_now: () => 'just now',
	minutes_ago: ({ n }: { n: number }) => `${n} minutes ago`,
	hours_ago: ({ n }: { n: number }) => `${n} hours ago`,
	days_ago: ({ n }: { n: number }) => `${n} days ago`,
	months_ago: ({ n }: { n: number }) => `${n} months ago`,
	years_ago: ({ n }: { n: number }) => `${n} years ago`,
}));

import { fmtSize, fmtTime, fmtSpeed, fmtDurationHMS, fmtDurationText, timeAgo, authedUrl } from './format';

// ── localStorage mock ─────────────────────────────────────────────

let store: Record<string, string>;

beforeEach(() => {
	store = {};
	vi.stubGlobal('localStorage', {
		getItem: (k: string) => store[k] ?? null,
		setItem: (k: string, v: string) => { store[k] = v; },
		removeItem: (k: string) => { delete store[k]; },
		clear: () => { store = {}; },
	});
	vi.stubGlobal('window', { location: { origin: 'http://localhost' } });
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── fmtSize ────────────────────────────────────────────────────────

describe('fmtSize', () => {
	it('formats bytes', () => {
		expect(fmtSize(0)).toBe('0 B');
		expect(fmtSize(512)).toBe('512 B');
		expect(fmtSize(1023)).toBe('1023 B');
	});

	it('formats kilobytes', () => {
		expect(fmtSize(1024)).toBe('1.0 KB');
		expect(fmtSize(1536)).toBe('1.5 KB');
	});

	it('formats megabytes', () => {
		expect(fmtSize(1024 * 1024)).toBe('1.0 MB');
		expect(fmtSize(5.5 * 1024 * 1024)).toBe('5.5 MB');
	});

	it('formats gigabytes', () => {
		expect(fmtSize(1024 * 1024 * 1024)).toBe('1.00 GB');
		expect(fmtSize(2.5 * 1024 * 1024 * 1024)).toBe('2.50 GB');
	});
});

// ── fmtTime ────────────────────────────────────────────────────────

describe('fmtTime', () => {
	it('formats ISO string to readable date-time', () => {
		const result = fmtTime('2025-01-15T08:30:00Z');
		expect(result).toMatch(/2025\/01\/15/);
		// The time portion depends on the test machine timezone, so just verify format
		expect(result).toMatch(/\d{2}:\d{2}$/);
	});

	it('pads single-digit months and days', () => {
		const result = fmtTime('2025-03-05T02:09:00Z');
		expect(result).toContain('03/05');
	});
});

// ── fmtSpeed ───────────────────────────────────────────────────────

describe('fmtSpeed', () => {
	it('returns empty string for zero or negative', () => {
		expect(fmtSpeed(0)).toBe('');
		expect(fmtSpeed(-100)).toBe('');
	});

	it('formats bytes per second', () => {
		expect(fmtSpeed(500)).toBe('500 B/s');
	});

	it('formats kilobytes per second', () => {
		expect(fmtSpeed(1024)).toBe('1.0 KB/s');
		expect(fmtSpeed(1536)).toBe('1.5 KB/s');
	});

	it('formats megabytes per second', () => {
		expect(fmtSpeed(1024 * 1024)).toBe('1.0 MB/s');
	});
});

// ── fmtDurationHMS ────────────────────────────────────────────────

describe('fmtDurationHMS', () => {
	it('returns null for undefined, zero, or negative', () => {
		expect(fmtDurationHMS()).toBeNull();
		expect(fmtDurationHMS(0)).toBeNull();
		expect(fmtDurationHMS(-5)).toBeNull();
	});

	it('formats seconds only (m:ss)', () => {
		expect(fmtDurationHMS(5)).toBe('0:05');
		expect(fmtDurationHMS(59)).toBe('0:59');
	});

	it('formats minutes and seconds (m:ss)', () => {
		expect(fmtDurationHMS(60)).toBe('1:00');
		expect(fmtDurationHMS(90)).toBe('1:30');
		expect(fmtDurationHMS(3599)).toBe('59:59');
	});

	it('formats hours, minutes, and seconds (h:mm:ss)', () => {
		expect(fmtDurationHMS(3600)).toBe('1:00:00');
		expect(fmtDurationHMS(3661)).toBe('1:01:01');
		expect(fmtDurationHMS(7200)).toBe('2:00:00');
		expect(fmtDurationHMS(7384)).toBe('2:03:04');
	});
});

// ── fmtDurationText ───────────────────────────────────────────────

describe('fmtDurationText', () => {
	it('returns "0s" for undefined, zero, or negative', () => {
		expect(fmtDurationText()).toBe('0s');
		expect(fmtDurationText(0)).toBe('0s');
		expect(fmtDurationText(-10)).toBe('0s');
	});

	it('formats seconds only', () => {
		expect(fmtDurationText(5)).toBe('5s');
		expect(fmtDurationText(59)).toBe('59s');
	});

	it('formats minutes only', () => {
		expect(fmtDurationText(60)).toBe('1min');
		expect(fmtDurationText(120)).toBe('2min');
	});

	it('formats minutes and seconds', () => {
		expect(fmtDurationText(90)).toBe('1min 30s');
		expect(fmtDurationText(150)).toBe('2min 30s');
	});

	it('formats hours only', () => {
		expect(fmtDurationText(3600)).toBe('1h');
		expect(fmtDurationText(7200)).toBe('2h');
	});

	it('formats hours and minutes', () => {
		expect(fmtDurationText(5400)).toBe('1h 30min');
		expect(fmtDurationText(3661)).toBe('1h 1min');
	});
});

// ── timeAgo ───────────────────────────────────────────────────────

describe('timeAgo', () => {
	it('returns "just now" for less than 60 seconds ago', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now)).toBe('just now');
		expect(timeAgo(now - 30)).toBe('just now');
	});

	it('returns minutes ago for < 1 hour', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now - 60)).toBe('1 minutes ago');
		expect(timeAgo(now - 300)).toBe('5 minutes ago');
		expect(timeAgo(now - 3599)).toBe('59 minutes ago');
	});

	it('returns hours ago for < 1 day', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now - 3600)).toBe('1 hours ago');
		expect(timeAgo(now - 7200)).toBe('2 hours ago');
	});

	it('returns days ago for < 30 days', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now - 86400)).toBe('1 days ago');
		expect(timeAgo(now - 86400 * 7)).toBe('7 days ago');
	});

	it('returns months ago for < 1 year', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now - 86400 * 30)).toBe('1 months ago');
		expect(timeAgo(now - 86400 * 90)).toBe('3 months ago');
	});

	it('returns years ago for >= 1 year', () => {
		const now = Math.floor(Date.now() / 1000);
		expect(timeAgo(now - 86400 * 365)).toBe('1 years ago');
		expect(timeAgo(now - 86400 * 365 * 2)).toBe('2 years ago');
	});
});

// ── authedUrl ─────────────────────────────────────────────────────

describe('authedUrl', () => {
	it('appends access_token when token exists', () => {
		store['nd.access'] = 'mytoken';
		const result = authedUrl('/api/v1/files/abc/download');
		expect(result).toContain('access_token=mytoken');
		expect(result).toContain('/api/v1/files/abc/download');
	});

	it('returns original URL when no token', () => {
		const result = authedUrl('/api/v1/files/abc/download');
		expect(result).toBe('/api/v1/files/abc/download');
	});

	it('encodes special characters in token', () => {
		store['nd.access'] = 'a&b=c';
		const result = authedUrl('/api/v1/test');
		expect(result).toContain(encodeURIComponent('a&b=c'));
	});
});
