// Global test setup: provide localStorage for module-level code
const store: Record<string, string> = {};

const lsStub = {
	getItem: (k: string) => store[k] ?? null,
	setItem: (k: string, v: string) => { store[k] = v; },
	removeItem: (k: string) => { delete store[k]; },
	clear: () => { for (const k in store) delete store[k]; },
};

(globalThis as Record<string, unknown>).localStorage = lsStub;

// Export clear helper for tests
export function clearLocalStorage() {
	lsStub.clear();
}
