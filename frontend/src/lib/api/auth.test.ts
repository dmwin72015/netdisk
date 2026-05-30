import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { login, register, logout } from './auth';
import { getAccessToken, getRefreshToken, getStoredUser } from './client';

// ── helpers ────────────────────────────────────────────────────────

let store: Record<string, string>;

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

function errorResponse(status: number, error: string): Response {
	return new Response(JSON.stringify({ error }), {
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

// ── login ─────────────────────────────────────────────────────────

describe('login', () => {
	const loginData = {
		user: {
			slug: 'alice',
			username: 'alice',
			email: 'a@b.com',
			status: 1,
			profile: { displayName: 'Alice', avatarUrl: '', bio: '' },
			storage: { storageUsed: 0, storageQuota: 1024 },
			level: { levelCode: 'free', levelName: 'Free', expiresAt: null },
			createdAt: '2025-01-01T00:00:00Z',
		},
		tokens: {
			accessToken: 'acc-token',
			refreshToken: 'ref-token',
			expiresIn: 3600,
		},
	};

	it('posts email and password with auth=false', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(loginData));
		vi.stubGlobal('fetch', fetchSpy);

		await login('a@b.com', 'secret');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/auth/login');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ email: 'a@b.com', password: 'secret' });
		// auth=false means no Authorization header
		expect(init.headers.get('Authorization')).toBeNull();
	});

	it('stores session after successful login', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(loginData));
		vi.stubGlobal('fetch', fetchSpy);

		await login('a@b.com', 'secret');

		expect(getStoredUser()).toEqual(loginData.user);
		expect(getAccessToken()).toBe('acc-token');
		expect(getRefreshToken()).toBe('ref-token');
	});

	it('returns login data', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(loginData));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await login('a@b.com', 'secret');
		expect(result.user.username).toBe('alice');
		expect(result.tokens.accessToken).toBe('acc-token');
	});
});

// ── register ──────────────────────────────────────────────────────

describe('register', () => {
	const userInfo = {
		slug: 'newuser',
		username: 'newuser',
		email: 'new@b.com',
		status: 1,
		profile: { displayName: 'New User', avatarUrl: '', bio: '' },
		storage: { storageUsed: 0, storageQuota: 1024 },
		level: { levelCode: 'free', levelName: 'Free', expiresAt: null },
		createdAt: '2025-01-01T00:00:00Z',
	};

	it('posts username, email, and password with auth=false', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(userInfo));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await register('newuser', 'new@b.com', 'password123');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/auth/register');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({
			username: 'newuser',
			email: 'new@b.com',
			password: 'password123',
		});
		expect(init.headers.get('Authorization')).toBeNull();
		expect(result).toEqual(userInfo);
	});
});

// ── logout ────────────────────────────────────────────────────────

describe('logout', () => {
	it('posts refreshToken and clears session', async () => {
		// Pre-populate session
		store['nd.refresh'] = 'ref-token';
		store['nd.access'] = 'acc-token';

		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await logout();

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/auth/logout');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ refreshToken: 'ref-token' });

		// Session should be cleared
		expect(getAccessToken()).toBeNull();
		expect(getRefreshToken()).toBeNull();
		expect(getStoredUser()).toBeNull();
	});

	it('clears session even if no refresh token', async () => {
		// No refresh token in store
		vi.stubGlobal('fetch', vi.fn());

		await logout();

		expect(getAccessToken()).toBeNull();
		expect(getRefreshToken()).toBeNull();
	});

	it('clears session even if API call fails', async () => {
		store['nd.refresh'] = 'ref-token';
		store['nd.access'] = 'acc-token';

		vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network error')));

		await logout();

		expect(getAccessToken()).toBeNull();
		expect(getRefreshToken()).toBeNull();
	});
});
