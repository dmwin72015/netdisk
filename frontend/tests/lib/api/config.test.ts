import { describe, it, expect, vi } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

import { getClientConfig, DeviceType } from '$lib/api/config';

// ── helpers ────────────────────────────────────────────────────────

function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'Content-Type': 'application/json' },
	});
}

// ── getClientConfig ────────────────────────────────────────────────

describe('getClientConfig', () => {
	it('uses default device "web" when no argument', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			jsonResponse({ device: 'web', configs: {} })
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getClientConfig();

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/config?device=web');
		expect(result.device).toBe('web');
	});

	it('passes explicit device "pc"', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			jsonResponse({ device: 'pc', configs: { 'upload.chunkSize': 4194304 } })
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getClientConfig('pc');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/config?device=pc');
		expect(result.device).toBe('pc');
	});

	it('passes explicit device "mobile"', async () => {
		const fetchSpy = vi.fn().mockResolvedValue(
			jsonResponse({ device: 'mobile', configs: {} })
		);
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getClientConfig('mobile');

		const url = fetchSpy.mock.calls[0][0] as string;
		expect(url).toBe('/api/v1/config?device=mobile');
		expect(result.device).toBe('mobile');
	});

	it('returns the configs object from the response', async () => {
		const configs = { 'upload.chunkSize': 2097152, 'upload.maxFileSize': 1073741824 };
		const fetchSpy = vi.fn().mockResolvedValue(jsonResponse({ device: 'web', configs }));
		vi.stubGlobal('fetch', fetchSpy);
		vi.stubGlobal('localStorage', { getItem: () => null });

		const result = await getClientConfig();
		expect(result.configs).toEqual(configs);
	});
});
