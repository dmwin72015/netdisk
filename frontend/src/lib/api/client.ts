// Lightweight fetch wrapper that injects the bearer token, transparently
// refreshes a stale access token on 401, and normalises errors.

import { browser } from '$app/environment';

const ACCESS_KEY = 'nd.access';
const REFRESH_KEY = 'nd.refresh';
const USER_KEY = 'nd.user';

export type UserInfo = {
	slug: string;
	username: string;
	email: string;
	status: number;
	role?: string;
	profile: {
		displayName: string;
		avatarUrl: string;
		bio: string;
	};
	storage: {
		storageUsed: number;
		storageQuota: number;
	};
	level: {
		levelCode: string;
		levelName: string;
		expiresAt: string | null;
	};
	createdAt: string;
};

export type Tokens = {
	accessToken: string;
	refreshToken: string;
	expiresIn: number;
};

export class ApiError extends Error {
	status: number;
	errCode: number;
	constructor(message: string, status: number, errCode: number) {
		super(message);
		this.status = status;
		this.errCode = errCode;
	}
}

function readStorage(key: string): string | null {
	if (!browser) return null;
	return localStorage.getItem(key);
}

function writeStorage(key: string, value: string | null) {
	if (!browser) return;
	if (value === null) localStorage.removeItem(key);
	else localStorage.setItem(key, value);
}

export function getAccessToken(): string | null {
	return readStorage(ACCESS_KEY);
}

export function getRefreshToken(): string | null {
	return readStorage(REFRESH_KEY);
}

export function getStoredUser(): UserInfo | null {
	const raw = readStorage(USER_KEY);
	if (!raw) return null;
	try {
		return JSON.parse(raw) as UserInfo;
	} catch {
		return null;
	}
}

export function setSession(user: UserInfo | null, tokens: Tokens | null) {
	if (user && tokens) {
		writeStorage(USER_KEY, JSON.stringify(user));
		writeStorage(ACCESS_KEY, tokens.accessToken);
		writeStorage(REFRESH_KEY, tokens.refreshToken);
	} else {
		writeStorage(USER_KEY, null);
		writeStorage(ACCESS_KEY, null);
		writeStorage(REFRESH_KEY, null);
	}
}

export function updateTokens(tokens: Tokens) {
	writeStorage(ACCESS_KEY, tokens.accessToken);
	writeStorage(REFRESH_KEY, tokens.refreshToken);
}

async function rawRequest(
	input: RequestInfo | URL,
	init: RequestInit,
	token: string | null
): Promise<Response> {
	const headers = new Headers(init.headers ?? {});
	if (token) headers.set('Authorization', `Bearer ${token}`);
	if (init.body && !headers.has('Content-Type') && !(init.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}
	return fetch(input, { ...init, headers });
}

let refreshing: Promise<string | null> | null = null;

async function tryRefresh(): Promise<string | null> {
	if (refreshing) return refreshing;
	const refresh = getRefreshToken();
	if (!refresh) return null;
	refreshing = (async () => {
		try {
			const res = await fetch('/api/v1/auth/refresh', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refreshToken: refresh })
			});
			if (!res.ok) {
				setSession(null, null);
				return null;
			}
			const json = await res.json();
			const tokens = json.data as Tokens;
			updateTokens(tokens);
			return tokens.accessToken;
		} catch {
			setSession(null, null);
			return null;
		} finally {
			refreshing = null;
		}
	})();
	return refreshing;
}

export type ApiOptions = RequestInit & {
	/** Set false to skip the 401 → refresh → retry flow (e.g. for auth endpoints). */
	auth?: boolean;
};

export async function api<T = unknown>(path: string, options: ApiOptions = {}): Promise<T> {
	const { auth = true, ...init } = options;
	const method = init.method ?? 'GET';
	const reqId = Math.random().toString(36).slice(2, 8);

	console.debug(`[api:${reqId}] → ${method} ${path}`, {
		contentType: init.headers ? (init.headers as Record<string, string>)['Content-Type'] : undefined,
		bodySize: init.body ? (typeof init.body === 'string' ? init.body.length : init.body instanceof FormData ? 'form-data' : 'unknown') : 0,
	});

	let token = auth ? getAccessToken() : null;
	let res = await rawRequest(path, init, token);

	if (res.status === 401 && auth) {
		console.warn(`[api:${reqId}] 401 on ${method} ${path}, attempting token refresh`);
		const fresh = await tryRefresh();
		if (fresh) {
			console.info(`[api:${reqId}] token refreshed, retrying ${method} ${path}`);
			res = await rawRequest(path, init, fresh);
		} else if (browser) {
			console.warn(`[api:${reqId}] token refresh failed, redirecting to /login`);
			window.location.href = '/login';
			return new Promise<never>(() => {});
		}
	}

	const contentType = res.headers.get('content-type') ?? '';
	if (!res.ok) {
		let message = res.statusText;
		let errCode = 0;
		if (contentType.includes('application/json')) {
			try {
				const body = await res.json();
				message = body?.error ?? message;
				errCode = body?.errCode ?? 0;
			} catch {
				// ignore
			}
		}
		console.error(`[api:${reqId}] ✗ ${method} ${path} → ${res.status} "${message}" (errCode=${errCode})`);
		throw new ApiError(message || `HTTP ${res.status}`, res.status, errCode);
	}

	if (res.status === 204) {
		console.debug(`[api:${reqId}] ✓ ${method} ${path} → 204`);
		return undefined as T;
	}
	if (contentType.includes('application/json')) {
		const body = await res.json();
		// Business error: HTTP 200 with non-zero code
		if (body?.code && body.code !== 0) {
			console.error(`[api:${reqId}] ✗ ${method} ${path} → business error code=${body.code} "${body.error}"`);
			throw new ApiError(body.error || 'Business error', 200, body.code);
		}
		console.debug(`[api:${reqId}] ✓ ${method} ${path} → ${res.status}`, body);
		return (body?.data ?? body) as T;
	}
	const text = await res.text();
	console.debug(`[api:${reqId}] ✓ ${method} ${path} → ${res.status} (text, ${text.length} chars)`);
	return text as unknown as T;
}
