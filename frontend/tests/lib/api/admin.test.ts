import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import {
	adminListUsers,
	adminGetUser,
	adminUpdateRole,
	adminUpdateStorageBase,
	adminDeleteUser,
} from '$lib/api/admin';

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

// ── adminListUsers ─────────────────────────────────────────────────

describe('adminListUsers', () => {
	it('fetches with default pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await adminListUsers();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('/api/v1/admin/users?');
		expect(url).toContain('limit=20');
		expect(url).toContain('offset=0');
	});

	it('uses custom pagination', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ items: [], total: 0 }));
		vi.stubGlobal('fetch', fetchSpy);

		await adminListUsers(50, 10);

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toContain('limit=50');
		expect(url).toContain('offset=10');
	});

	it('returns parsed data', async () => {
		const data = { items: [{ id: '1', username: 'alice' }], total: 1, limit: 20, offset: 0 };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(data));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await adminListUsers();
		expect(result).toEqual(data);
	});
});

// ── adminGetUser ───────────────────────────────────────────────────

describe('adminGetUser', () => {
	it('fetches user by id', async () => {
		const user = { id: 'u1', username: 'alice', role: 'user' };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(user));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await adminGetUser('u1');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/admin/users/u1');
		expect(result).toEqual(user);
	});
});

// ── adminUpdateRole ────────────────────────────────────────────────

describe('adminUpdateRole', () => {
	it('sends PATCH with role', async () => {
		const user = { id: 'u1', username: 'alice', role: 'admin' };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(user));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await adminUpdateRole('u1', 'admin');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/admin/users/u1');
		expect(init.method).toBe('PATCH');
		expect(JSON.parse(init.body)).toEqual({ role: 'admin' });
		expect(result).toEqual(user);
	});
});

// ── adminUpdateStorageBase ─────────────────────────────────────────

describe('adminUpdateStorageBase', () => {
	it('sends PATCH with baseBytes', async () => {
		const user = { id: 'u1', baseBytes: 1073741824 };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse(user));
		vi.stubGlobal('fetch', fetchSpy);

		const result = await adminUpdateStorageBase('u1', 1073741824);

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/admin/users/u1/storage-base');
		expect(init.method).toBe('PATCH');
		expect(JSON.parse(init.body)).toEqual({ baseBytes: 1073741824 });
		expect(result).toEqual(user);
	});
});

// ── adminDeleteUser ────────────────────────────────────────────────

describe('adminDeleteUser', () => {
	it('sends DELETE to user endpoint', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
		vi.stubGlobal('fetch', fetchSpy);

		await adminDeleteUser('u1');

		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('/api/v1/admin/users/u1');
		expect(init.method).toBe('DELETE');
	});
});
