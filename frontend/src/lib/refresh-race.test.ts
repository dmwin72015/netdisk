/**
 * Tests for the refresh stale-guard race condition fix.
 *
 * Background: when a large file is uploading, the user may trigger a
 * `refresh(true)` (e.g. by clicking sort).  Later, when the upload
 * completes, the upload manager calls `refresh(false, true)` so the
 * file list is re-listed with the newly imported file.
 *
 * Without `force=true`, if the user-triggered request's response
 * arrives *after* the upload-completion response, the stale guard
 * (`id !== refreshId`) discards the upload-completion result and the
 * new file never appears.
 *
 * These tests verify the stale-guard logic with the `force` bypass.
 */

import { describe, it, expect } from 'vitest';

// ---------------------------------------------------------------------------
// We test the core guard pattern that lives inside refresh():
//
//   if (id === refreshId || force) { apply result }
// ---------------------------------------------------------------------------

describe('refresh stale guard', () => {
	it('applies result when request is not stale (no force needed)', async () => {
		let files: string[] = [];
		let refreshId = 0;

		const applyIfCurrent = (id: number, data: string[], force = false) => {
			if (id === refreshId || force) {
				files = data;
			}
		};

		const id1 = ++refreshId; // id === refreshId → applies
		applyIfCurrent(id1, ['a.txt']);

		expect(files).toEqual(['a.txt']);
	});

	it('discards stale result when force is false', async () => {
		let files: string[] = [];
		let refreshId = 0;

		const applyIfCurrent = (id: number, data: string[], force = false) => {
			if (id === refreshId || force) {
				files = data;
			}
		};

		const id1 = ++refreshId;
		refreshId++; // simulate a newer refresh being started

		applyIfCurrent(id1, ['stale.txt']);
		expect(files).toEqual([]); // stale → discarded
	});

	it('applies stale result when force is true', async () => {
		let files: string[] = [];
		let refreshId = 0;

		const applyIfCurrent = (id: number, data: string[], force = false) => {
			if (id === refreshId || force) {
				files = data;
			}
		};

		const id1 = ++refreshId;
		refreshId++; // simulate a newer refresh being started

		// This is what upload-completion refresh does: force=true
		applyIfCurrent(id1, ['new-file.txt'], true);
		expect(files).toEqual(['new-file.txt']); // force bypasses stale guard
	});

	it('simulates the full race condition: upload completion wins', async () => {
		/**
		 * Simulate the timeline:
		 *   1.  User-triggered refresh during upload (id=1, slow)
		 *   2.  Upload-completion refresh (id=2, fast, force=true)
		 *
		 * The slow request's response arrives AFTER the fast one.
		 * Without force=true, the slow stale response would overwrite
		 * the fresh data.  With force=true, the fresh data sticks.
		 */
		let files: string[] = [];
		let refreshId = 0;

		const applyIfCurrent = async (
			id: number,
			data: string[],
			delay: number,
			force = false,
		) => {
			await new Promise((r) => setTimeout(r, delay));
			if (id === refreshId || force) {
				files = data;
			}
		};

		// User-triggered refresh (slow, 100ms)
		const p1 = applyIfCurrent(++refreshId, ['existing.txt'], 100, false);
		// Upload-completion refresh (fast, 10ms, force=true)
		const p2 = applyIfCurrent(++refreshId, ['existing.txt', 'uploaded.mp4'], 10, true);

		await p2; // fast one finishes first
		expect(files).toEqual(['existing.txt', 'uploaded.mp4']);

		await p1; // slow one finishes later, but stale guard blocks it
		expect(files).toEqual(['existing.txt', 'uploaded.mp4']); // fresh data preserved
	});

	it('loadingRequestId pattern: only the request that set loading can clear it', () => {
		let loading = false;
		let loadingRequestId = 0;
		let refreshId = 0;

		const setLoading = (show: boolean) => {
			if (show) {
				loadingRequestId = ++refreshId;
				loading = true;
			}
		};

		const maybeClearLoading = (id: number) => {
			if (id === loadingRequestId) {
				loading = false;
			}
		};

		// Request A: showLoading=true
		setLoading(true);
		expect(loadingRequestId).toBe(1);

		// Request B: showLoading=false (upload completion)
		// does NOT change loadingRequestId
		++refreshId;

		// Request A's finally runs
		maybeClearLoading(1);
		expect(loading).toBe(false); // cleared by the request that set it

		// Request B's finally runs
		maybeClearLoading(2);
		expect(loading).toBe(false); // already false, no-op
	});

	it('loadingRequestId prevents stale request from prematurely clearing loading', () => {
		let loading = false;
		let loadingRequestId = 0;
		let refreshId = 0;

		const setLoading = (show: boolean) => {
			if (show) {
				loadingRequestId = ++refreshId;
				loading = true;
			}
		};

		const maybeClearLoading = (id: number) => {
			if (id === loadingRequestId) {
				loading = false;
			}
		};

		// Request A: showLoading=true
		setLoading(true); // loadingRequestId = 1
		expect(loading).toBe(true);

		// Request B: showLoading=true (newer user action)
		setLoading(true); // loadingRequestId = 2
		expect(loadingRequestId).toBe(2);

		// Request A's finally runs (stale)
		maybeClearLoading(1); // id=1 !== loadingRequestId=2 → no-op
		expect(loading).toBe(true); // still loading, waiting for B

		// Request B's finally runs
		maybeClearLoading(2); // id=2 === loadingRequestId=2 → cleared
		expect(loading).toBe(false);
	});
});
