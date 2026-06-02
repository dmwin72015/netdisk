import type { UploadItem } from '$lib/types/upload';
import {
	preCheck, requestChallenge, verify as verifyUpload,
	initUpload, uploadChunk, completeUpload, getUploadStatus, updateHash
} from '$lib/api/upload';
import { importFile } from '$lib/api/files';
import { computeSHA256Chunked } from '$lib/upload-hash';
import { fmtSize } from '$lib/utils/format';
import { confirmAction } from '$lib/dialog';
import * as m from '$lib/paraglide/messages';
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
	errorMsg: string | null;
};

export function createUploadManager(opts: {
	getCurrentSlug?: () => string | null | Promise<string | null>;
	onCompleted?: () => void | Promise<void>;
	acceptFile?: (file: File) => boolean;
	onRejected?: (files: File[]) => void;
	onFileImported?: (result: ImportedUploadFile) => void | Promise<void>;
	storageKey?: string;
	maxConcurrent?: number;
}) {
	const STORAGE_KEY = opts.storageKey || '';

	let _getCurrentSlug = opts.getCurrentSlug ?? (() => null);
	let _onCompleted = opts.onCompleted ?? (() => {});
	let _acceptFile = opts.acceptFile;
	let _onRejected = opts.onRejected;
	let _onFileImported = opts.onFileImported;

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
				errorMsg: i.errorMsg,
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

	let uploadItems = $state<UploadItem[]>([]);

	$effect(() => {
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
			console.error(`[upload:${uid}]${s} ${label} FAILED: ${e.message}`, {
				name: e.name,
				stack: e.stack?.split('\n').slice(0, 4).join('\n'),
			});
		} else {
			console.error(`[upload:${uid}]${s} ${label} FAILED:`, e);
		}
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

	async function importPhysicalFile(item: UploadItem, physicalFileSlug: string) {
		item.phase = 'importing';
		const parentSlug = await _getCurrentSlug();
		const imported = await importFile(physicalFileSlug, item.fileName, parentSlug || undefined);
		await _onFileImported?.({
			fileSlug: imported.fileSlug,
			fileName: imported.fileName,
			physicalFileSlug,
			item
		});
	}

	// --- File picker handlers ---

	function onPick(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const fileList = el?.files;
		if (!fileList || fileList.length === 0) return;

		const pickedFiles = filterAcceptedFiles(Array.from(fileList));
		el.value = '';
		if (pickedFiles.length === 0) return;

		const newItems: UploadItem[] = pickedFiles.map((f) => ({
			uid: nextUid(),
			file: f,
			fileName: f.name,
			fileSize: f.size,
			fileHash: '',
			preHash: '',
			phase: 'hashing',
			progress: 0,
			hashProgress: 0,
			uploadedBytes: 0,
			speed: 0,
			uploadSlug: null,
			sessionId: null,
			abortCtrl: null,
			errorMsg: null
		}));

		uploadItems = [...uploadItems, ...newItems];
		for (const item of newItems) {
			log(item.uid, null, `selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
		}

		void startUploadQueue();
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

	function onFolderConfirm(selected: { file: File; relativePath: string }[]) {
		folderDialogOpen = false;
		if (selected.length === 0) return;

		const newItems: UploadItem[] = selected.map((f) => ({
			uid: nextUid(),
			file: f.file,
			fileName: f.relativePath,
			fileSize: f.file.size,
			fileHash: '',
			preHash: '',
			phase: 'hashing',
			progress: 0,
			hashProgress: 0,
			uploadedBytes: 0,
			speed: 0,
			uploadSlug: null,
			sessionId: null,
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
		while (runningCount < maxConcurrent) {
			const idx = uploadItems.findIndex((i) => (i.phase === 'hashing' || i.phase === 'pending') && !processing.has(i.uid));
			if (idx === -1) break;
			const item = uploadItems[idx];
			processing.add(item.uid);
			runningCount++;
			processFile(item).finally(() => {
				processing.delete(item.uid);
				runningCount--;
				pumpQueue();
			});
		}
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

			try {
				const t0 = Date.now();
				const cs = getChunkSize();
				if (cs === null) throw new Error('config unavailable');
				const chunkSize = cs;
				const totalChunks = getUploadChunkCount(item.fileSize, chunkSize);
				const useFastPath = shouldUseSmallFileFastPath(item.fileSize, chunkSize);
			let hashPromise: ReturnType<typeof computeSHA256Chunked> | null = null;

			const shouldPrepareHash = item.phase === 'hashing' || (item.phase === 'pending' && (!item.preHash || !item.fileHash));
			const shouldWaitHashBeforeInit = item.phase === 'pending' && !item.fileHash;

			if (shouldPrepareHash) {
				log(item.uid, null, `computing hash, totalChunks=${totalChunks}`);

				hashPromise = computeSHA256Chunked(item.file, {
					onPreHash: (preHash) => {
						item.preHash = preHash;
						log(item.uid, null, `preHash ready (${Date.now() - t0}ms): ${preHash}`);
					},
					onProgress: (percent) => {
						item.hashProgress = percent;
					}
				}, chunkSize);

				while (!item.preHash) {
					await new Promise(r => setTimeout(r, 10));
				}

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
								const sampleBytes = new Uint8Array(sampleBuffer);

								const tokenBytes = new TextEncoder().encode(challenge.challengeToken);
								const proofInput = new Uint8Array(sampleBytes.length + tokenBytes.length);
								proofInput.set(sampleBytes);
								proofInput.set(tokenBytes, sampleBytes.length);
								const proofHash = await crypto.subtle.digest('SHA-256', proofInput);
								const proofCode = Array.from(new Uint8Array(proofHash))
									.map(b => b.toString(16).padStart(2, '0'))
									.join('');

								const verifyResult = await verifyUpload(hash, proofCode);
								log(item.uid, null, `verify result: ${verifyResult.status}`);

								if (verifyResult.status === 'HIT' && verifyResult.physicalFileSlug) {
									log(item.uid, null, `dedup HIT, importing from ${verifyResult.physicalFileSlug}`);

									if (verifyResult.existingFiles && verifyResult.existingFiles.length > 0) {
										const paths = verifyResult.existingFiles.map(f => f.path).join('\n');
										const confirmed = await confirmAction(
											m.duplicate_detected(),
											m.duplicate_file_paths({ paths }),
											m.continue_upload()
										);
										if (!confirmed) {
											item.phase = 'failed';
											item.errorMsg = m.upload_skipped_duplicate();
											log(item.uid, null, 'skipped by user (duplicate)');
											return;
										}
									}

									item.phase = 'importing';
									item.progress = 100;
									await importPhysicalFile(item, verifyResult.physicalFileSlug);
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
			const parentSlug = await _getCurrentSlug();
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

			async function uploadChunkWorker() {
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
					if (i === null) return;

					const chunkStart = i * chunkSize;
					const chunkEnd = Math.min(chunkStart + chunkSize, item.fileSize);
					const chunkData = await item.file.slice(chunkStart, chunkEnd).arrayBuffer();

					let chunkOk = false;
					for (let attempt = 1; attempt <= MAX_CHUNK_RETRIES; attempt++) {
						if ((item.phase as string) === 'paused') return;
						try {
							await uploadChunk(task.uploadSlug, i, chunkData, item.abortCtrl?.signal);
							chunkOk = true;
							break;
						} catch (e) {
							if ((item.phase as string) === 'paused') return;
							if (attempt < MAX_CHUNK_RETRIES) {
								warn(item.uid, task.uploadSlug, `chunk ${i + 1}/${totalChunks} attempt ${attempt}/${MAX_CHUNK_RETRIES} failed, retrying...`, e);
								await new Promise(r => setTimeout(r, chunkRetryDelay(attempt)));
							} else {
								apiLogError(item.uid, task.uploadSlug, `chunk ${i + 1}/${totalChunks} (final attempt)`, e);
								throw e;
							}
						}
					}
					if (!chunkOk) throw new Error(`chunk ${i}/${totalChunks} failed after ${MAX_CHUNK_RETRIES} attempts`);

					log(item.uid, task.uploadSlug, `chunk ${i + 1}/${totalChunks} uploaded (${fmtSize(chunkEnd)}/${fmtSize(item.fileSize)})`);
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
				}
			}

			if (chunkWorkers > 0) {
				await Promise.all(Array.from({ length: chunkWorkers }, () => uploadChunkWorker().catch((e) => {
					chunkFailure ??= e;
				})));
			}
			if (chunkFailure) throw chunkFailure;
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
			await completeUpload(task.uploadSlug);

			const completedTask = await getUploadStatus(task.uploadSlug);
			if (completedTask.physicalFileSlug) {
				log(item.uid, task.uploadSlug, `importing physicalFile: ${completedTask.physicalFileSlug}`);
				await importPhysicalFile(item, completedTask.physicalFileSlug);
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

	function clearCompleted() {
		uploadItems = uploadItems.filter((i) => i.phase !== 'completed');
	}

	// --- Persistence ---

	function restore() {
		const snap = restoreSnapshot();
		if (snap.length === 0) return;
		const activePhases = new Set(['hashing', 'pending', 'verifying', 'uploading', 'paused']);
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
		item.phase = 'hashing';
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
		onPick,
		onPickFolder,
		onFolderConfirm,
		pauseUpload,
		resumeUpload,
		deleteUpload,
		clearCompleted,
		restore,
		retryItem,
		setGetCurrentSlug,
		setOnCompleted,
		setAcceptFile,
		setOnRejected,
		setOnFileImported,
		updateMaxConcurrent(n: number) {
			maxConcurrent = Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, n));
			pumpQueue();
		},
	};
}
