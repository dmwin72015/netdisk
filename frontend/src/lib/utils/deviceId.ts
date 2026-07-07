import FingerprintJS from '@fingerprintjs/fingerprintjs';

const STORAGE_KEY = 'nd.device_id';
let cache: string | null = null;

function fallbackId(): string {
	let v = localStorage.getItem(STORAGE_KEY);
	if (!v) {
		v = crypto.randomUUID();
		localStorage.setItem(STORAGE_KEY, v);
	}
	return v;
}

/**
 * Returns a stable per-device identifier for the "login devices" feature.
 *
 * Prefers the FingerprintJS visitor id. If the library is unavailable or fails
 * (network/load error), falls back to a localStorage-persisted UUID so the id
 * stays stable across reloads. The fallback id is also persisted on success so
 * a transient FingerprintJS failure does not rotate the id.
 */
export async function getDeviceId(): Promise<string> {
	if (cache) return cache;
	try {
		const fp = await FingerprintJS.load();
		const result = await fp.get();
		cache = result.visitorId;
		localStorage.setItem(STORAGE_KEY, cache);
		return cache;
	} catch {
		cache = fallbackId();
		return cache;
	}
}
