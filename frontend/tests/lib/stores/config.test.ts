import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));
vi.mock('$lib/paraglide/messages', () => ({}));

// Mock the API module so fetchConfig's getClientConfig() returns our data
vi.mock('$lib/api/config', () => ({
	getClientConfig: vi.fn(),
}));

import { clientConfig, configError, fetchConfig, isConfigReady, getChunkSize, getMaxUploadSize, getAvatarMaxSize } from '$lib/stores/config';
import { get } from 'svelte/store';
import { getClientConfig } from '$lib/api/config';

const mockedGetClientConfig = getClientConfig as ReturnType<typeof vi.fn>;

beforeEach(() => {
	mockedGetClientConfig.mockClear();
});

describe('fetchConfig', () => {
	it('sets clientConfig and clears configError on success', async () => {
		mockedGetClientConfig.mockResolvedValue({ device: 'web', configs: { 'upload.chunkSize': 4194304 } });

		await fetchConfig();

		expect(get(clientConfig)).toEqual({ device: 'web', configs: { 'upload.chunkSize': 4194304 } });
		expect(get(configError)).toBe(false);
	});

	it('sets configError on failure', async () => {
		clientConfig.set(null);
		configError.set(false);
		mockedGetClientConfig.mockRejectedValue(new Error('network'));

		await fetchConfig();

		expect(get(clientConfig)).toBeNull();
		expect(get(configError)).toBe(true);
	});

	it('sets configError on non-ok response (api throws)', async () => {
		clientConfig.set(null);
		configError.set(false);
		mockedGetClientConfig.mockRejectedValue(new Error('HTTP 500'));

		await fetchConfig();

		expect(get(configError)).toBe(true);
		expect(get(clientConfig)).toBeNull();
	});
});

describe('isConfigReady', () => {
	it('returns false when config is null', () => {
		clientConfig.set(null);
		expect(isConfigReady()).toBe(false);
	});

	it('returns true when config is loaded', () => {
		clientConfig.set({ device: 'web', configs: {} });
		expect(isConfigReady()).toBe(true);
	});
});

describe('getChunkSize', () => {
	it('returns value from config', () => {
		clientConfig.set({ device: 'web', configs: { 'upload.chunkSize': 2097152 } });
		expect(getChunkSize()).toBe(2097152);
	});

	it('returns null when key is missing', () => {
		clientConfig.set({ device: 'web', configs: {} });
		expect(getChunkSize()).toBeNull();
	});

	it('returns null when config is null', () => {
		clientConfig.set(null);
		expect(getChunkSize()).toBeNull();
	});
});

describe('getMaxUploadSize', () => {
	it('returns value from config', () => {
		clientConfig.set({ device: 'web', configs: { 'upload.maxUploadSize': 1073741824 } });
		expect(getMaxUploadSize()).toBe(1073741824);
	});

	it('returns null when key is missing', () => {
		clientConfig.set({ device: 'web', configs: {} });
		expect(getMaxUploadSize()).toBeNull();
	});
});

describe('getAvatarMaxSize', () => {
	it('returns value from config', () => {
		clientConfig.set({ device: 'web', configs: { 'avatar.maxSize': 5242880 } });
		expect(getAvatarMaxSize()).toBe(5242880);
	});

	it('returns null when key is missing', () => {
		clientConfig.set({ device: 'web', configs: {} });
		expect(getAvatarMaxSize()).toBeNull();
	});
});
