import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { getProfile, updateProfile, uploadAvatar, getStorageBreakdown } from './profile';

// ── helpers ────────────────────────────────────────────────────────

let store: Record<string, string>;

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

beforeEach(() => {
	store = {};
	vi.stubGlobal('localStorage', {
		getItem: (k: string) => store[k] ?? null,
		setItem: (k: string, v: string) => { store[k] = v; },
		removeItem: (k: string) => { delete store[k]; },
		clear: () => { store = {}; },
	});
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── getProfile ────────────────────────────────────────────────────

describe('getProfile', () => {
	it('fetches /api/v1/user/me', async () => {
		const profile = {
			slug: 'alice',
			username: 'alice',
			email: 'a@b.com',
			profile: { displayName: 'Alice', avatarUrl: '', bio: '' },
			storage: { storageUsed: 100, storageQuota: 1024 },
			level: { levelCode: 'free', levelName: 'Free', expiresAt: null },
			createdAt: '2025-01-01T00:00:00Z',
		};
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(profile));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getProfile();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/user/me');
		expect(result).toEqual(profile);
	});
});

// ── updateProfile ─────────────────────────────────────────────────

describe('updateProfile', () => {
	it('patches profile data', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ message: 'ok' }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await updateProfile({ displayName: 'New Name', bio: 'Hello' });

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/user/profile');
		expect(init.method).toBe('PATCH');
		expect(JSON.parse(init.body)).toEqual({ displayName: 'New Name', bio: 'Hello' });
		expect(result).toEqual({ message: 'ok' });
	});

	it('sends partial updates', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ message: 'ok' }));
		vi.stubGlobal('fetch', fetchSpy);

		await updateProfile({ bio: 'Just bio' });

		const body = JSON.parse(fetchSpy.mock.calls[0][1].body);
		expect(body).toEqual({ bio: 'Just bio' });
	});
});

// ── uploadAvatar ──────────────────────────────────────────────────

describe('uploadAvatar', () => {
	it('sends FormData with file and returns avatar URL', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ avatar_url: '/avatars/new.jpg' }));
		vi.stubGlobal('fetch', fetchSpy);

		const file = new File(['avatar data'], 'avatar.jpg', { type: 'image/jpeg' });
		const result = await uploadAvatar(file);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/user/me/avatar');
		expect(init.method).toBe('POST');
		expect(init.body).toBeInstanceOf(FormData);
		expect(result).toBe('/avatars/new.jpg');
	});
});

// ── getStorageBreakdown ───────────────────────────────────────────

describe('getStorageBreakdown', () => {
	it('fetches and returns categories array', async () => {
		const categories = [
			{ category: 'image', bytes: 5000, count: 10 },
			{ category: 'video', bytes: 50000, count: 3 },
		];
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ categories }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await getStorageBreakdown();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/user/storage-breakdown');
		expect(result).toEqual(categories);
		expect(result).toHaveLength(2);
	});
});
