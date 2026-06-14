import type { UploadItem } from '$lib/types/upload';
import {
	preCheck, requestChallenge, verify as verifyUpload,
	initUpload, uploadChunk, completeUpload, getUploadStatus, updateHash
} from '$lib/api/upload';
import type { ExistingFileRef } from '$lib/api/upload';
import { importFile, checkConflict, trashFile } from '$lib/api/files';
import { computeSHA256Chunked } from '$lib/upload-hash';
import { fmtSize } from '$lib/utils/format';
import { confirmAction } from '$lib/dialog';
import { openConfirm, getCheckboxValue } from '$lib/dialog-state.svelte';
import * as m from '$lib/paraglide/messages';
import { getDuplicateStrategy, setDuplicateStrategy } from '$lib/stores/file-preferences.svelte';
import { getChunkSize, configError } from '$lib/stores/config';
import { get } from 'svelte/store';
import { getUploadChunkCount, shouldUseSmallFileFastPath } from '$lib/upload-strategy';
import { getUploadedBytesFromCompletedChunks, getUploadProgress, normalizeCompletedChunks } from '$lib/upload-resume';
import { UPLOAD_CHUNK_CONCURRENCY_PER_FILE, UPLOAD_FILE_CONCURRENCY } from '$lib/upload-concurrency';

export type ImportedUploadFile = {
	fileSlug: string;
	fileName: string;
	physicalFileSlug: string;
	item: UploadItem;
};

export type ImportConflictContext = {
	physicalFileSlug: string;
	fileName: string;
	parentSlug: string | null;
	source: 'dedup' | 'upload';
	item: UploadItem;
	error: unknown;
};

export type DuplicateFileContext = {
	physicalFileSlug: string;
	existingFiles: ExistingFileRef[];
	item: UploadItem;
};

export type NameConflictInfo = {
	uid: string;
	fileName: string;
	fileSize: number;
	existingSlug: string;
};

export type NameConflictResult = {
	strategy: 'overwrite' | 'skip' | 'keep_both';
	applyToAll: boolean;
};

type SnapshotItem = {
	uid: string;
	fileName: string;
	fileSize: number;
	fileHash: string;
	preHash: string;
	phase: string;
	progress: number;
	uploadedBytes: number;
	uploadSlug: string | null;
	sessionId: string | null;
	parentSlug: string | null;
	errorMsg: string | null;
	conflictStrategy?: 'overwrite' | 'keep_both';
	conflictExistingSlug?: string;
};

export function createUploadManager(opts: {
	getCurrentSlug?: () => string | null | Promise<string | null>;
	onCompleted?: () => void | Promise<void>;
	acceptFile?: (file: File) => boolean;
	onRejected?: (files: File[]) => void;
	onFileImported?: (result: ImportedUploadFile) => void | Promise<void>;
	onImportConflict?: (context: ImportConflictContext) => boolean | Promise<boolean>;
	onDuplicateDetected?: (context: DuplicateFileContext) => boolean | Promise<boolean>;
	onNameConflicts?: (conflicts: NameConflictInfo[]) => Promise<Map<string, NameConflictResult>>;
	storageKey?: string;
	maxConcurrent?: number;
}) {
	const STORAGE_KEY = opts.storageKey || '';

	let _getCurrentSlug = opts.getCurrentSlug ?? (() => null);
	let _onCompleted = opts.onCompleted ?? (() => {});
	let _acceptFile = opts.acceptFile;
	let _onRejected = opts.onRejected;
	let _onFileImported = opts.onFileImported;
	let _onImportConflict = opts.onImportConflict;
	let _onDuplicateDetected = opts.onDuplicateDetected;
	let _onNameConflicts = opts.onNameConflicts;

	function setGetCurrentSlug(fn: () => string | null | Promise<string | null>) {
		_getCurrentSlug = fn;
	}
	function setOnCompleted(fn: () => void | Promise<void>) {
		_onCompleted = fn;
	}
	function setAcceptFile(fn?: (file: File) => boolean) {
		_acceptFile = fn;
	}
	function setOnRejected(fn?: (files: File[]) => void) {
		_onRejected = fn;
	}
	function setOnFileImported(fn?: (result: ImportedUploadFile) => void | Promise<void>) {
		_onFileImported = fn;
	}
	function setOnImportConflict(fn?: (context: ImportConflictContext) => boolean | Promise<boolean>) {
		_onImportConflict = fn;
	}
	function setOnDuplicateDetected(fn?: (context: DuplicateFileContext) => boolean | Promise<boolean>) {
		_onDuplicateDetected = fn;
	}
	function setOnNameConflicts(fn?: (conflicts: NameConflictInfo[]) => Promise<Map<string, NameConflictResult>>) {
		_onNameConflicts = fn;
	}

	function saveSnapshot() {
		if (!STORAGE_KEY) return;
		try {
			const snap: SnapshotItem[] = uploadItems.map((i) => ({
				uid: i.uid,
				fileName: i.fileName,
				fileSize: i.fileSize,
				fileHash: i.fileHash,
				preHash: i.preHash,
				phase: i.phase,
				progress: i.progress,
				uploadedBytes: i.uploadedBytes,
				uploadSlug: i.uploadSlug,
				sessionId: i.sessionId,
				parentSlug: i.parentSlug,
				errorMsg: i.errorMsg,
				conflictStrategy: i.conflictStrategy,
				conflictExistingSlug: i.conflictExistingSlug,
			}));
			localStorage.setItem(STORAGE_KEY, JSON.stringify(snap));
		} catch {
			// storage full or unavailable
		}
	}

	function restoreSnapshot() {
		if (!STORAGE_KEY) return [];
		try {
			const raw = localStorage.getItem(STORAGE_KEY);
			if (!raw) return [];
			const snap: SnapshotItem[] = JSON.parse(raw);
			return snap;
		} catch {
			return [];
		}
	}

	function clearSnapshot() {
		if (!STORAGE_KEY) return;
		try {
			localStorage.removeItem(STORAGE_KEY);
		} catch {
			// ignore
		}
	}

	let uploadItems = $state<UploadItem[]>([]);
	let _hydrated = false;

	$effect(() => {
		if (!_hydrated) return;
		const _ = uploadItems;
		saveSnapshot();
	});

	let folderDialogFiles = $state<{ file: File; relativePath: string }[]>([]);
	let folderDialogOpen = $state(false);
	let folderDialogLoading = $state(false);

	let uidCounter = 0;
	function nextUid() {
		return `upload-${++uidCounter}`;
	}

	const MAX_CHUNK_RETRIES = 3;
	const CHUNK_RETRY_BASE_DELAY_MS = 1000;

	function chunkRetryDelay(attempt: number) {
		return CHUNK_RETRY_BASE_DELAY_MS * 2 ** (attempt - 1) + Math.floor(Math.random() * 400);
	}

	function nextMissingChunk(completedChunks: Set<number>, totalChunks: number) {
		for (let i = 0; i < totalChunks; i++) {
			if (!completedChunks.has(i)) return i;
		}
		return null;
	}

	function log(uid: string, slug: string | null, msg: string, data?: unknown) {
		const s = slug ? `[${slug}]` : '';
		const prefix = `[upload:${uid}]${s}`;
		if (data !== undefined) console.log(prefix, msg, data);
		else console.log(prefix, msg);
	}

	function warn(uid: string, slug: string | null, msg: string, data?: unknown) {
		const s = slug ? `[${slug}]` : '';
		console.warn(`[upload:${uid}]${s} ${msg}`, data ?? '');
	}

	function err(uid: string, slug: string | null, msg: string, ...args: unknown[]) {
		const s = slug ? `[${slug}]` : '';
		console.error(`[upload:${uid}]${s} ${msg}`, ...args);
	}

	function apiLogError(uid: string, slug: string | null, label: string, e: unknown) {
		const s = slug ? `[${slug}]` : '';
		if (e instanceof Error) {
			const cause = (e as { cause?: unknown }).cause;
			console.error(`[upload:${uid}]${s} ${label} FAILED: ${e.message}`, {
				name: e.name,
				status: (e as { status?: unknown }).status,
				errCode: (e as { errCode?: unknown }).errCode,
				cause: cause instanceof Error ? cause.message : cause,
				stack: e.stack?.split('\n').slice(0, 6).join('\n'),
			});
		} else {
			console.error(`[upload:${uid}]${s} ${label} FAILED:`, e);
		}
	}

	function logChunkProgress(uid: string, slug: string | null, completedChunks: Set<number>, totalChunks: number, label: string) {
		const sorted = [...completedChunks].sort((a, b) => a - b);
		const missing: number[] = [];
		for (let i = 0; i < totalChunks; i++) {
			if (!completedChunks.has(i)) missing.push(i);
		}
		log(uid, slug, `${label} progress: ${completedChunks.size}/${totalChunks} completed, missing: [${missing.slice(0, 20).join(',')}${missing.length > 20 ? ',...' : ''}]`);
	}

	function filterAcceptedFiles(files: File[]) {
		if (!_acceptFile) return files;

		const accepted: File[] = [];
		const rejected: File[] = [];
		for (const file of files) {
			if (_acceptFile(file)) accepted.push(file);
			else rejected.push(file);
		}

		if (rejected.length > 0) {
			_onRejected?.(rejected);
		}
		return accepted;
	}

	function getFileExtension(name: string): string {
		const i = name.lastIndexOf('.');
		return i > 0 ? name.slice(i) : '';
	}

	function getFileNameWithoutExt(name: string): string {
		const i = name.lastIndexOf('.');
		return i > 0 ? name.slice(0, i) : name;
	}

	function windowsKeepBothName(name: string, counter: number): string {
		const ext = getFileExtension(name);
		const base = getFileNameWithoutExt(name);
		return `${base} (${counter})${ext}`;
	}

	async function resolveKeepBothName(originalName: string, parentSlug?: string): Promise<string> {
		let counter = 2;
		while (true) {
			const candidate = windowsKeepBothName(originalName, counter);
			try {
				const resp = await checkConflict(candidate, 0, '', parentSlug ?? undefined);
				if (resp.status !== 'NAME_CONFLICT') {
					return candidate;
				}
			} catch {
				return candidate;
			}
			counter++;
		}
	}

	async function resolveItemNameConflict(item: UploadItem): Promise<boolean> {
		// Returns true if the item was skipped/removed (caller should stop processing).
		if (!_onNameConflicts) return false;

		let resp: Awaited<ReturnType<typeof checkConflict>>;
		try {
			resp = await checkConflict(item.fileName, 0, '', item.parentSlug ?? undefined);
		} catch (e) {
			warn(item.uid, null, 'name-conflict check failed, proceeding without it', e);
			return false;
		}
		if (resp.status !== 'NAME_CONFLICT' || !resp.existing) return false;
		const existing = resp.existing;

		const autoStrategy = getDuplicateStrategy();
		if (autoStrategy === 'skip') {
			removeItemSilently(item.uid, `skipped due to name conflict: ${item.fileName}`);
			return true;
		}
		if (autoStrategy === 'overwrite') {
			item.conflictStrategy = 'overwrite';
			item.conflictExistingSlug = existing.slug;
			return false;
		}
		if (autoStrategy === 'keep_both') {
			item.fileName = await resolveKeepBothName(item.fileName, item.parentSlug ?? undefined);
			return false;
		}

		const results = await _onNameConflicts([
			{
				uid: item.uid,
				fileName: item.fileName,
				fileSize: item.fileSize,
				existingSlug: existing.slug
			}
		]);
		const result = results.get(item.uid);
		if (!result) return false;
		switch (result.strategy) {
			case 'skip':
				removeItemSilently(item.uid, `skipped due to name conflict: ${item.fileName}`);
				return true;
			case 'overwrite':
				item.conflictStrategy = 'overwrite';
				item.conflictExistingSlug = existing.slug;
				return false;
			case 'keep_both':
				item.fileName = await resolveKeepBothName(item.fileName, item.parentSlug ?? undefined);
				return false;
		}
		return false;
	}

	async function importPhysicalFile(item: UploadItem, physicalFileSlug: string, source: 'dedup' | 'upload') {
		item.phase = 'importing';
		const parentSlug = item.parentSlug;

		if (item.conflictStrategy === 'overwrite' && item.conflictExistingSlug) {
			try {
				log(item.uid, null, `overwrite: trashing existing file ${item.conflictExistingSlug}`);
				await trashFile(item.conflictExistingSlug);
			} catch (e) {
				warn(item.uid, null, 'overwrite: failed to trash existing file, import may still conflict', e);
			}
		}

		let imported: Awaited<ReturnType<typeof importFile>>;
		try {
			imported = await importFile(physicalFileSlug, item.fileName, parentSlug || undefined);
		} catch (error) {
			const handled = await _onImportConflict?.({
				physicalFileSlug,
				fileName: item.fileName,
				parentSlug,
				source,
				item,
				error
			});
			if (handled) return;
			throw error;
		}
		await _onFileImported?.({
			fileSlug: imported.fileSlug,
			fileName: imported.fileName,
			physicalFileSlug,
			item
		});
	}

	// --- File picker handlers ---

	async function enqueueFiles(files: File[]) {
		const pickedFiles = filterAcceptedFiles(files);
		if (pickedFiles.length === 0) return 0;

		const currentSlug = await _getCurrentSlug();
		const newItems: UploadItem[] = pickedFiles.map((f) => ({
			uid: nextUid(),
			file: f,
			fileName: f.name,
			fileSize: f.size,
			fileHash: '',
			preHash: '',
			phase: 'queued',
			progress: 0,
			hashProgress: 0,
			uploadedBytes: 0,
			speed: 0,
			uploadSlug: null,
			sessionId: null,
			parentSlug: currentSlug,
			abortCtrl: null,
			errorMsg: null
		}));

		uploadItems = [...uploadItems, ...newItems];
		for (const item of newItems) {
			log(item.uid, null, `selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
		}

		void startUploadQueue();
		return newItems.length;
	}

	function removeItemSilently(uid: string, reason: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (item) log(uid, null, reason);
		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	async function onPick(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const fileList = el?.files;
		if (!fileList || fileList.length === 0) return;

		const files = Array.from(fileList);
		el.value = '';
		await enqueueFiles(files);
	}

	function onPickFolder(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const fileList = el?.files;
		if (!fileList || fileList.length === 0) return;

		folderDialogFiles = [];
		folderDialogOpen = true;
		folderDialogLoading = true;

		setTimeout(() => {
			const pickedFiles = Array.from(fileList).map((f) => ({
				file: f,
				relativePath: ('webkitRelativePath' in f ? (f as { webkitRelativePath: string }).webkitRelativePath : '') || f.name,
			}));
			el.value = '';

			const acceptedFiles = filterAcceptedFiles(pickedFiles.map((f) => f.file));
			const acceptedSet = new Set(acceptedFiles);
			folderDialogFiles = pickedFiles.filter((f) => acceptedSet.has(f.file));
			folderDialogLoading = false;
			if (folderDialogFiles.length === 0) {
				folderDialogOpen = false;
			}
		}, 50);
	}

	async function onFolderConfirm(selected: { file: File; relativePath: string }[]) {
		folderDialogOpen = false;
		if (selected.length === 0) return;

		const currentSlug = await _getCurrentSlug();
		const newItems: UploadItem[] = selected.map((f) => ({
			uid: nextUid(),
			file: f.file,
			fileName: f.relativePath,
			fileSize: f.file.size,
			fileHash: '',
			preHash: '',
			phase: 'queued',
			progress: 0,
			hashProgress: 0,
			uploadedBytes: 0,
			speed: 0,
			uploadSlug: null,
			sessionId: null,
			parentSlug: currentSlug,
			abortCtrl: null,
			errorMsg: null
		}));

		uploadItems = [...uploadItems, ...newItems];
		for (const item of newItems) {
			log(item.uid, null, `selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
		}
		void startUploadQueue();
	}

	// --- Upload queue engine ---

	let runningCount = 0;
	let maxConcurrent = opts.maxConcurrent ?? UPLOAD_FILE_CONCURRENCY;
	const processing = new Set<string>();

	function startUploadQueue() {
		pumpQueue();
	}

	function pumpQueue() {
		const isReady = (i: UploadItem) =>
			(i.phase === 'queued' || i.phase === 'pending') && !processing.has(i.uid);
		const waiting = uploadItems.filter(isReady);
		if (waiting.length > 0 || runningCount > 0) {
			log('pump', null, `running=${runningCount} max=${maxConcurrent} waiting=${waiting.length} total=${uploadItems.length}`);
		}
		while (runningCount < maxConcurrent) {
			const idx = uploadItems.findIndex(isReady);
			if (idx === -1) break;
			const item = uploadItems[idx];
			processing.add(item.uid);
			runningCount++;
			if (item.phase === 'queued') {
				item.phase = 'hashing';
			}
			log(item.uid, item.uploadSlug, `pumping: starting processFile (phase=${item.phase})`);
			processFile(item).finally(() => {
				processing.delete(item.uid);
				runningCount--;
				log('pump', null, `processFile completed for ${item.uid}, running=${runningCount}`);
				pumpQueue();
			});
		}
	}

	async function computeProofCode(sampleBuffer: ArrayBuffer, challengeToken: string): Promise<string> {
		if (typeof crypto === 'undefined' || !crypto.subtle) {
			throw new Error('crypto.subtle not available (non-HTTPS context)');
		}
		const sampleBytes = new Uint8Array(sampleBuffer);
		const tokenBytes = new TextEncoder().encode(challengeToken);
		const proofInput = new Uint8Array(sampleBytes.length + tokenBytes.length);
		proofInput.set(sampleBytes);
		proofInput.set(tokenBytes, sampleBytes.length);
		const proofHash = await crypto.subtle.digest('SHA-256', proofInput);
		return Array.from(new Uint8Array(proofHash))
			.map(b => b.toString(16).padStart(2, '0'))
			.join('');
	}

	async function processFile(item: UploadItem) {
		const tStart = Date.now();
		log(item.uid, item.uploadSlug, `picked from queue, phase=${item.phase}`);

		if (get(configError)) {
			item.phase = 'failed';
			item.errorMsg = m.config_unavailable();
			log(item.uid, null, 'config unavailable, rejecting');
			return;
		}

		// Per-file name-conflict resolution: only happens when this file's slot opens.
		// Queued files in the back of the line do nothing until their turn.
		const skipped = await resolveItemNameConflict(item);
		if (skipped) return;

		try {
			const t0 = Date.now();
			const cs = getChunkSize();
			if (cs === null) throw new Error('config unavailable');
			const chunkSize = cs;
			const totalChunks = getUploadChunkCount(item.fileSize, chunkSize);
			const useFastPath = shouldUseSmallFileFastPath(item.fileSize, chunkSize);
			let hashPromise: ReturnType<typeof computeSHA256Chunked> | null = null;
			let preHashResolve: (() => void) | null = null;
			const preHashReady = new Promise<void>((resolve) => { preHashResolve = resolve; });

			const shouldPrepareHash = item.phase === 'hashing' || (item.phase === 'pending' && (!item.preHash || !item.fileHash));
			const shouldWaitHashBeforeInit = item.phase === 'pending' && !item.fileHash;

			if (shouldPrepareHash) {
				log(item.uid, null, `computing hash, totalChunks=${totalChunks}`);

				hashPromise = computeSHA256Chunked(item.file, {
					onPreHash: (preHash) => {
						item.preHash = preHash;
						preHashResolve?.();
						log(item.uid, null, `preHash ready (${Date.now() - t0}ms): ${preHash}`);
					},
					onProgress: (percent) => {
						item.hashProgress = percent;
					}
				}, chunkSize);

				const preHashTimeout = 60_000;
				await Promise.race([
					preHashReady,
					new Promise<never>((_, reject) =>
						setTimeout(() => reject(new Error(`preHash timeout after ${preHashTimeout}ms`)), preHashTimeout)
					),
				]);

				if (useFastPath) {
					const { hash } = await hashPromise!;
					item.fileHash = hash;
					log(item.uid, null, `small file fast path, skip preCheck/challenge (${Date.now() - t0}ms): ${hash}`);
				} else {
					const preResult = await preCheck(item.preHash, item.fileSize);
					log(item.uid, null, `preCheck result: ${preResult.status}`);

					if (preResult.status === 'SUSPECT_HIT') {
						const { hash } = await hashPromise!;
						item.fileHash = hash;
						log(item.uid, null, `fullHash done (${Date.now() - t0}ms): ${hash}`);

						item.phase = 'verifying';
						log(item.uid, null, 'phase → verifying, requesting challenge...');
						const challenge = await requestChallenge(hash);

						if (challenge.status === 'CHALLENGE') {
							const sampleStart = challenge.challengeOffset;
							const sampleEnd = Math.min(sampleStart + 1024, item.fileSize);
							const sampleBlob = item.file.slice(sampleStart, sampleEnd);
							const sampleBuffer = await sampleBlob.arrayBuffer();
							const proofCode = await computeProofCode(sampleBuffer, challenge.challengeToken);

							const verifyResult = await verifyUpload(hash, proofCode);
							log(item.uid, null, `verify result: ${verifyResult.status}`);

							if (verifyResult.status === 'HIT' && verifyResult.physicalFileSlug) {
								log(item.uid, null, `dedup HIT, importing from ${verifyResult.physicalFileSlug}`);

								const strategy = getDuplicateStrategy();
								let proceed = false;

								if (strategy === 'skip') {
									item.phase = 'failed';
									item.errorMsg = m.upload_skipped_duplicate();
									log(item.uid, null, 'skipped (duplicate strategy: skip)');
									return;
								} else if (strategy === 'overwrite' || strategy === 'keep_both') {
									proceed = true;
									if (strategy === 'overwrite') {
										item.conflictStrategy = 'overwrite';
									}
								} else {
									const existing = verifyResult.existingFiles!;
									if (_onDuplicateDetected) {
										proceed = await _onDuplicateDetected({
											physicalFileSlug: verifyResult.physicalFileSlug,
											existingFiles: existing,
											item
										});
									} else {
										proceed = await openConfirm({
											title: m.duplicate_detected(),
											message: m.duplicate_file_paths({ paths: existing.map(f => f.path).join('\n') }),
											confirmText: m.continue_upload(),
											cancelText: m.cancel(),
											checkboxLabel: m.upload_conflict_apply_all(),
										});
										if (getCheckboxValue()) {
											setDuplicateStrategy(proceed ? 'keep_both' : 'skip');
										}
									}
								}

								if (!proceed) {
									item.phase = 'failed';
									item.errorMsg = m.upload_skipped_duplicate();
									log(item.uid, null, 'skipped by user (duplicate)');
									return;
								}

								item.phase = 'importing';
								item.progress = 100;
								await importPhysicalFile(item, verifyResult.physicalFileSlug, 'dedup');
								item.phase = 'completed';
								item.uploadedBytes = item.fileSize;
								log(item.uid, null, 'completed (dedup)');
								try { await _onCompleted(); } catch (e) { warn(item.uid, null, 'onCompleted threw', e); }
								return;
							}
							log(item.uid, null, 'verify MISS, falling through to upload');
						} else {
							log(item.uid, null, `challenge status: ${challenge.status}, no challenge issued`);
						}
					} else if (shouldWaitHashBeforeInit) {
						const { hash } = await hashPromise!;
						item.fileHash = hash;
						log(item.uid, null, `resume fullHash done (${Date.now() - t0}ms): ${hash}`);
					} else {
						log(item.uid, null, 'preCheck NOT_FOUND, going straight to upload');
					}
				}
			}

			item.phase = 'uploading';
			log(item.uid, null, 'phase → uploading, init session...');
			const mimeType = item.file.type || 'application/octet-stream';
			const parentSlug = item.parentSlug;
			const task = await initUpload(item.fileHash, item.preHash, item.fileSize, mimeType, item.file.name, parentSlug || undefined);
			item.uploadSlug = task.uploadSlug;
			let hashSynced = Boolean(item.fileHash);
			let hashSyncPromise: Promise<void> | null = null;
			function syncHash(hash: string) {
				item.fileHash = hash;
				if (hashSynced) return Promise.resolve();
				hashSyncPromise ??= updateHash(task.uploadSlug, hash, item.preHash)
					.then(() => {
						hashSynced = true;
					})
					.finally(() => {
						hashSyncPromise = null;
					});
				return hashSyncPromise;
			}
			if (!item.fileHash && hashPromise) {
				hashPromise
					.then(({ hash }) => syncHash(hash))
					.catch((e) => warn(item.uid, task.uploadSlug, 'background hash sync failed', e));
			}
			const completedChunks = normalizeCompletedChunks(task.completedChunks, totalChunks);
			log(item.uid, task.uploadSlug, `session created, completed chunks: ${completedChunks.size}/${totalChunks}`);

			item.abortCtrl = new AbortController();
			let uploaded = getUploadedBytesFromCompletedChunks(completedChunks, totalChunks, chunkSize, item.fileSize);
			item.uploadedBytes = uploaded;
			item.progress = getUploadProgress(uploaded, item.fileSize);
			let lastTime = Date.now();
			let lastBytes = uploaded;
			let chunkCursor = 0;
			let chunkFailure: unknown = null;
			const chunkWorkers = Math.min(UPLOAD_CHUNK_CONCURRENCY_PER_FILE, Math.max(0, totalChunks - completedChunks.size));

			const CHUNK_WORKER_LOG_INTERVAL = 5; // log every N chunks
			async function uploadChunkWorker(workerId: number) {
				let workerChunksDone = 0;
				while (!chunkFailure) {
					if ((item.phase as string) === 'paused') return;
					let i: number | null = null;
					while (chunkCursor < totalChunks) {
						if (!completedChunks.has(chunkCursor)) {
							i = chunkCursor;
							chunkCursor++;
							break;
						}
						chunkCursor++;
					}
					if (i === null) {
						log(item.uid, task.uploadSlug, `worker#${workerId}: no more chunks to upload (processed ${workerChunksDone} chunks)`);
						return;
					}

					const chunkStart = i * chunkSize;
					const chunkEnd = Math.min(chunkStart + chunkSize, item.fileSize);
					const chunkData = await item.file.slice(chunkStart, chunkEnd).arrayBuffer();

					let chunkOk = false;
					for (let attempt = 1; attempt <= MAX_CHUNK_RETRIES; attempt++) {
						if ((item.phase as string) === 'paused') return;
						try {
							log(item.uid, task.uploadSlug, `worker#${workerId} chunk ${i + 1}/${totalChunks} attempt ${attempt}/${MAX_CHUNK_RETRIES} (size=${chunkData.byteLength})`);
							await uploadChunk(task.uploadSlug, i, chunkData, item.abortCtrl?.signal);
							chunkOk = true;
							break;
						} catch (e) {
							if ((item.phase as string) === 'paused') return;
							if (attempt < MAX_CHUNK_RETRIES) {
								warn(item.uid, task.uploadSlug, `worker#${workerId} chunk ${i + 1}/${totalChunks} attempt ${attempt}/${MAX_CHUNK_RETRIES} failed, retrying in ${chunkRetryDelay(attempt)}ms...`, e);
								await new Promise(r => setTimeout(r, chunkRetryDelay(attempt)));
							} else {
								apiLogError(item.uid, task.uploadSlug, `worker#${workerId} chunk ${i + 1}/${totalChunks} (final attempt)`, e);
								throw e;
							}
						}
					}
					if (!chunkOk) throw new Error(`worker#${workerId} chunk ${i}/${totalChunks} failed after ${MAX_CHUNK_RETRIES} attempts`);

					workerChunksDone++;
					completedChunks.add(i);
					uploaded = getUploadedBytesFromCompletedChunks(completedChunks, totalChunks, chunkSize, item.fileSize);
					item.uploadedBytes = uploaded;
					item.progress = getUploadProgress(uploaded, item.fileSize);

					const now = Date.now();
					const elapsed = (now - lastTime) / 1000;
					if (elapsed >= 0.5) {
						item.speed = (uploaded - lastBytes) / elapsed;
						lastBytes = uploaded;
						lastTime = now;
					}

					if (workerChunksDone % CHUNK_WORKER_LOG_INTERVAL === 0) {
						logChunkProgress(item.uid, task.uploadSlug, completedChunks, totalChunks, `worker#${workerId}`);
					}
				}
			}

			if (chunkWorkers > 0) {
				log(item.uid, task.uploadSlug, `starting ${chunkWorkers} chunk workers for ${totalChunks - completedChunks.size} remaining chunks`);
				const workerPromises = Array.from({ length: chunkWorkers }, (_, idx) =>
					uploadChunkWorker(idx + 1).catch((e) => {
						chunkFailure ??= e;
					})
				);
				await Promise.all(workerPromises);
			} else {
				log(item.uid, task.uploadSlug, `no chunk workers needed (${completedChunks.size}/${totalChunks} already done)`);
			}
			if (chunkFailure) {
				err(item.uid, task.uploadSlug, `chunk workers failed: ${chunkFailure instanceof Error ? chunkFailure.message : chunkFailure}`, chunkFailure);
				throw chunkFailure;
			}
			logChunkProgress(item.uid, task.uploadSlug, completedChunks, totalChunks, 'post-upload');
			const missingChunk = nextMissingChunk(completedChunks, totalChunks);
			if (missingChunk !== null) throw new Error(`chunk ${missingChunk + 1}/${totalChunks} missing after upload`);

			if ((item.phase as string) === 'paused') return;

			if (!item.fileHash) {
				log(item.uid, task.uploadSlug, 'hash not yet ready, waiting...');
				if (hashPromise) {
					const { hash } = await hashPromise;
					item.fileHash = hash;
				}
				log(item.uid, task.uploadSlug, `fullHash done: ${item.fileHash}, updating...`);
			}
			if (item.fileHash && !hashSynced) {
				await syncHash(item.fileHash);
			}

			log(item.uid, task.uploadSlug, 'all chunks uploaded, completing...');
			log(item.uid, task.uploadSlug, `calling completeUpload for slug=${task.uploadSlug}`);
			const completeResp = await completeUpload(task.uploadSlug);
			log(item.uid, task.uploadSlug, `completeUpload response:`, completeResp);

			let physicalFileSlug: string | undefined = completeResp?.physicalFileSlug;
			if (completeResp?.status !== 'DONE') {
				log(item.uid, task.uploadSlug, `complete status is "${completeResp?.status}", polling getUploadStatus every 2s for completion...`);
				for (let attempt = 0; attempt < 60; attempt++) {
					await new Promise(r => setTimeout(r, 2000));
					const statusResp = await getUploadStatus(task.uploadSlug);
					log(item.uid, task.uploadSlug, `poll attempt ${attempt + 1}/60: status=${statusResp.status} slug=${statusResp.physicalFileSlug} err=${statusResp.error ?? 'none'}`);
					if (statusResp.status === 'done' && statusResp.physicalFileSlug) {
						physicalFileSlug = statusResp.physicalFileSlug;
						log(item.uid, task.uploadSlug, `poll: got done with physicalFileSlug=${physicalFileSlug}`);
						break;
					}
					if (statusResp.status === 'failed') {
						err(item.uid, task.uploadSlug, `poll: upload task failed with error: ${statusResp.error ?? 'unknown'}`);
						throw new Error(statusResp.error ?? 'upload task failed');
					}
				}
				if (!physicalFileSlug) {
					// Try one final getUploadStatus to see what happened
					try {
						const finalStatus = await getUploadStatus(task.uploadSlug);
						err(item.uid, task.uploadSlug, `poll timed out after 120s, final status:`, finalStatus);
					} catch (finalErr) {
						err(item.uid, task.uploadSlug, `poll timed out, and final status query also failed:`, finalErr);
					}
					throw new Error('upload did not complete in time');
				}
			} else {
				log(item.uid, task.uploadSlug, `complete returned DONE immediately with physicalFileSlug=${physicalFileSlug}`);
			}

			if (physicalFileSlug) {
				log(item.uid, task.uploadSlug, `importing physicalFile: ${physicalFileSlug}`);
				await importPhysicalFile(item, physicalFileSlug, 'upload');
			}

			item.phase = 'completed';
			item.progress = 100;
			item.uploadedBytes = item.fileSize;
			item.speed = 0;
			log(item.uid, task.uploadSlug, `completed in ${Date.now() - tStart}ms`);
			try { await _onCompleted(); } catch (e) { warn(item.uid, task.uploadSlug, 'onCompleted threw', e); }
		} catch (e) {
			if (item.phase === 'paused') {
				log(item.uid, item.uploadSlug, 'paused during operation, skipping');
				return;
			}
			item.phase = 'failed';
			item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
			apiLogError(item.uid, item.uploadSlug, `FAILED after ${Date.now() - tStart}ms`, e);
		} finally {
			item.abortCtrl = null;
		}
	}

	// --- Upload controls ---

	function pauseUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || item.phase !== 'uploading') return;
		log(uid, item.uploadSlug, 'paused by user');
		item.phase = 'paused';
		item.speed = 0;
		item.abortCtrl?.abort();
		item.abortCtrl = null;
	}

	function resumeUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || (item.phase !== 'paused' && item.phase !== 'failed')) return;
		log(uid, item.uploadSlug, `resumed by user (from ${item.phase})`);
		item.errorMsg = null;
		item.phase = 'pending';
		item.progress = 0;
		item.uploadedBytes = 0;
		item.speed = 0;
		void startUploadQueue();
	}

	function deleteUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (item) log(uid, item.uploadSlug, `deleted by user (phase=${item.phase})`);
		if (item?.phase === 'uploading') {
			item.abortCtrl?.abort();
		}
		processing.delete(uid);
		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	function dismissAll() {
		clearSnapshot();
		uploadItems = [];
	}

	// --- Persistence ---

	function restore() {
		const snap = restoreSnapshot();
		_hydrated = true;
		if (snap.length === 0) return;
		const activePhases = new Set(['queued', 'hashing', 'pending', 'verifying', 'uploading', 'paused']);
		const restored: UploadItem[] = snap.map((s) => ({
			uid: s.uid,
			file: null as unknown as File,
			fileName: s.fileName,
			fileSize: s.fileSize,
			fileHash: s.fileHash,
			preHash: s.preHash,
			phase: activePhases.has(s.phase) ? 'interrupted' : (s.phase as UploadItem['phase']),
			progress: s.progress,
			hashProgress: 0,
			uploadedBytes: s.uploadedBytes,
			speed: 0,
			uploadSlug: s.uploadSlug,
			sessionId: s.sessionId,
			abortCtrl: null,
			parentSlug: s.parentSlug ?? null,
			errorMsg: s.errorMsg,
		}));
		if (uidCounter === 0) {
			uidCounter = Math.max(...restored.map((i) => {
				const n = parseInt(i.uid.replace('upload-', ''), 10);
				return isNaN(n) ? 0 : n;
			}), 0);
		}
		uploadItems = [...restored, ...uploadItems];
	}

	function retryItem(uid: string, file: File) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item) return;
		item.file = file;
		item.errorMsg = null;
		item.phase = 'queued';
		item.hashProgress = 0;
		item.progress = 0;
		item.uploadedBytes = 0;
		item.speed = 0;
		item.fileHash = '';
		item.preHash = '';
		item.uploadSlug = null;
		item.sessionId = null;
		void startUploadQueue();
	}

	return {
		get items() { return uploadItems; },
		set items(v: UploadItem[]) { uploadItems = v; },
		get folderDialogFiles() { return folderDialogFiles; },
		get folderDialogOpen() { return folderDialogOpen; },
		set folderDialogOpen(v: boolean) { folderDialogOpen = v; },
		get folderDialogLoading() { return folderDialogLoading; },
		enqueueFiles,
		onPick,
		onPickFolder,
		onFolderConfirm,
		pauseUpload,
		resumeUpload,
		deleteUpload,
		dismissAll,
		restore,
		retryItem,
		setGetCurrentSlug,
		setOnCompleted,
		setAcceptFile,
		setOnRejected,
		setOnFileImported,
		setOnImportConflict,
		setOnDuplicateDetected,
		setOnNameConflicts,
		updateMaxConcurrent(n: number) {
			maxConcurrent = Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, n));
			pumpQueue();
		},
	};
}
