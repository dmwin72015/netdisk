import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock $app/environment before importing the module under test
vi.mock('$app/environment', () => ({ browser: true }));

// Mock $lib/paraglide/messages (not used by client.ts directly, but imported transitively)
vi.mock('$lib/paraglide/messages', () => ({}));

import {
	ApiError,
	setSession,
	getStoredUser,
	getAccessToken,
	getRefreshToken,
	updateTokens,
	api,
} from '$lib/api/client';
import type { UserInfo, Tokens } from '$lib/api/client';

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
});

afterEach(() => {
	vi.restoreAllMocks();
});

// ── helpers ────────────────────────────────────────────────────────

const mockUser: UserInfo = {
	slug: 'alice',
	username: 'alice',
	email: 'a@b.com',
	status: 1,
	profile: { displayName: 'Alice', avatarUrl: '', bio: '' },
	storage: { storageUsed: 0, storageQuota: 1024 },
	level: { levelCode: 'free', levelName: 'Free', expiresAt: null },
	createdAt: '2025-01-01T00:00:00Z',
};
const mockTokens: Tokens = {
	accessToken: 'acc',
	refreshToken: 'ref',
	expiresIn: 9999,
};

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

function errorResponse(status: number, error: string, errCode?: number): Response {
	return new Response(JSON.stringify({ error, errCode: errCode ?? 0 }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

// ── ApiError ───────────────────────────────────────────────────────

describe('ApiError', () => {
	it('sets status and errCode', () => {
		const err = new ApiError('not found', 404, 1001);
		expect(err.message).toBe('not found');
		expect(err.status).toBe(404);
		expect(err.errCode).toBe(1001);
		expect(err).toBeInstanceOf(Error);
	});
});

// ── setSession / getStoredUser ─────────────────────────────────────

describe('session storage', () => {
	it('setSession stores user and tokens', () => {
		setSession(mockUser, mockTokens);
		expect(getStoredUser()).toEqual(mockUser);
		expect(getAccessToken()).toBe('acc');
		expect(getRefreshToken()).toBe('ref');
	});

	it('setSession(null, null) clears everything', () => {
		setSession(mockUser, mockTokens);
		setSession(null, null);
		expect(getStoredUser()).toBeNull();
		expect(getAccessToken()).toBeNull();
		expect(getRefreshToken()).toBeNull();
	});

	it('getStoredUser returns null for corrupt JSON', () => {
		store['nd.user'] = '{broken';
		expect(getStoredUser()).toBeNull();
	});

	it('getStoredUser returns null when empty', () => {
		expect(getStoredUser()).toBeNull();
	});

	it('updateTokens refreshes tokens without touching user', () => {
		setSession(mockUser, mockTokens);
		updateTokens({ ...mockTokens, accessToken: 'new_acc', refreshToken: 'new_ref' });
		expect(getAccessToken()).toBe('new_acc');
		expect(getRefreshToken()).toBe('new_ref');
		expect(getStoredUser()).toEqual(mockUser);
	});
});

// ── api() ──────────────────────────────────────────────────────────

describe('api', () => {
	it('returns parsed data on success', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ id: '1', name: 'test' }));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await api<{ id: string; name: string }>('/api/v1/test');
		expect(result).toEqual({ id: '1', name: 'test' });
		expect(fetchSpy).toHaveBeenCalledOnce();
	});

	it('returns body directly when no data field', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ id: '1' }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' },
			})
		);
		vi.stubGlobal('fetch', fetchSpy);

		const result = await api<{ id: string }>('/api/v1/test');
		expect(result).toEqual({ id: '1' });
	});

	it('returns undefined for 204 No Content', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })));
		const result = await api('/api/v1/test');
		expect(result).toBeUndefined();
	});

	it('throws ApiError on non-2xx response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(errorResponse(400, 'bad request', 1004)));
		await expect(api('/api/v1/test')).rejects.toMatchObject({
			message: 'bad request',
			status: 400,
			errCode: 1004,
		});
	});

	it('throws ApiError with statusText when body has no error', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('oops', { status: 500, statusText: 'Internal Server Error' })));
		await expect(api('/api/v1/test')).rejects.toMatchObject({
			status: 500,
		});
	});

	it('injects Authorization header when token exists', async () => {
		setSession(mockUser, mockTokens);
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await api('/api/v1/test');

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('Authorization')).toBe('Bearer acc');
	});

	it('skips auth when auth=false', async () => {
		setSession(mockUser, mockTokens);
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await api('/api/v1/test', { auth: false });

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('Authorization')).toBeNull();
	});

	it('refreshes on 401 and retries', async () => {
		setSession(mockUser, mockTokens);

		const refreshRes = jsonResponse({
			accessToken: 'fresh',
			refreshToken: 'fresh_ref',
			expiresIn: 99999,
		});

		const fetchSpy = vi.fn()
			.mockResolvedValueOnce(errorResponse(401, 'unauthorized'))  // first call
			.mockResolvedValueOnce(refreshRes)                          // refresh call
			.mockResolvedValueOnce(jsonResponse({ ok: true }));         // retry

		vi.stubGlobal('fetch', fetchSpy);

		const result = await api<{ ok: boolean }>('/api/v1/test');
		expect(result).toEqual({ ok: true });
		expect(fetchSpy).toHaveBeenCalledTimes(3);

		// Verify the retry used the new token
		const retryCall = fetchSpy.mock.calls[2];
		expect(retryCall[1].headers.get('Authorization')).toBe('Bearer fresh');
	});

	it('sets Content-Type: application/json for non-FormData bodies', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		await api('/api/v1/test', { method: 'POST', body: JSON.stringify({ a: 1 }) });

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('Content-Type')).toBe('application/json');
	});

	it('does not set Content-Type for FormData bodies', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(null));
		vi.stubGlobal('fetch', fetchSpy);

		const fd = new FormData();
		fd.append('file', new Blob());
		await api('/api/v1/test', { method: 'POST', body: fd });

		const [, init] = fetchSpy.mock.calls[0];
		expect(init.headers.get('Content-Type')).toBeNull();
	});
});
