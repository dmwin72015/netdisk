import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';
import { getClientConfig, type ClientConfig } from '$lib/api/config';

export const clientConfig = writable<ClientConfig | null>(null);
export const configError = writable<boolean>(false);

export async function fetchConfig(): Promise<void> {
	if (!browser) return;
	try {
		const cfg = await getClientConfig();
		clientConfig.set(cfg);
		configError.set(false);
	} catch {
		configError.set(true);
	}
}

export function isConfigReady(): boolean {
	return get(clientConfig) !== null;
}

function cfgVal(key: string): unknown {
	return get(clientConfig)?.configs?.[key];
}

export function getChunkSize(): number | null {
	const v = cfgVal('upload.chunkSize') as number | undefined;
	return v ?? null;
}

export function getMaxUploadSize(): number | null {
	const v = cfgVal('upload.maxUploadSize') as number | undefined;
	return v ?? null;
}

export function getAvatarMaxSize(): number | null {
	const v = cfgVal('avatar.maxSize') as number | undefined;
	return v ?? null;
}
