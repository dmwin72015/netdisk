import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

// ── localStorage mock ──────────────────────────────────────────────

let store: Record<string, string>;

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

// ── setUser ────────────────────────────────────────────────────────

describe('setUser', () => {
	it('stores user in localStorage', async () => {
		const { setUser } = await import('$lib/stores/auth');
		const userInfo = {
			slug: 'alice',
			username: 'alice',
			email: 'a@b.com',
			status: 1,
			profile: { displayName: 'Alice', avatarUrl: '', bio: '' },
			storage: { storageUsed: 0, storageQuota: 1024 },
			level: { levelCode: 'free', levelName: 'Free', expiresAt: null },
			createdAt: '2025-01-01T00:00:00Z',
		};

		setUser(userInfo);

		expect(store['nd.user']).toBe(JSON.stringify(userInfo));
	});

	it('clears user from localStorage when null', async () => {
		const { setUser } = await import('$lib/stores/auth');

		setUser(null);

		expect(store['nd.user']).toBe('null');
	});

	it('updates the store value', async () => {
		const { setUser, user } = await import('$lib/stores/auth');
		const userInfo = {
			slug: 'bob',
			username: 'bob',
			email: 'b@b.com',
			status: 1,
			profile: { displayName: 'Bob', avatarUrl: '', bio: '' },
			storage: { storageUsed: 0, storageQuota: 2048 },
			level: { levelCode: 'pro', levelName: 'Pro', expiresAt: null },
			createdAt: '2025-06-01T00:00:00Z',
		};

		let value: unknown;
		const unsub = user.subscribe((v) => { value = v; });

		setUser(userInfo);
		expect(value).toEqual(userInfo);
		unsub();
	});

	it('sets authReady to true', async () => {
		const { setUser, authReady } = await import('$lib/stores/auth');

		let ready = false;
		const unsub = authReady.subscribe((v) => { ready = v; });

		setUser(null);
		expect(ready).toBe(true);
		unsub();
	});
});
