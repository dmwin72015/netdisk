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

const CHUNK_SIZE = 4 * 1024 * 1024;

export function createUploadManager(opts: {
	getCurrentSlug: () => string | null;
	onCompleted: () => void | Promise<void>;
}) {
	let uploadItems = $state<UploadItem[]>([]);
	let folderDialogFiles = $state<{ file: File; relativePath: string }[]>([]);
	let folderDialogOpen = $state(false);
	let folderDialogLoading = $state(false);

	let uidCounter = 0;
	function nextUid() {
		return `upload-${++uidCounter}`;
	}

	function log(uid: string, msg: string, data?: unknown) {
		const prefix = `[upload:${uid}]`;
		if (data !== undefined) console.log(prefix, msg, data);
		else console.log(prefix, msg);
	}

	// --- File picker handlers ---

	function onPick(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const fileList = el?.files;
		if (!fileList || fileList.length === 0) return;

		const pickedFiles = Array.from(fileList);
		el.value = '';

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
			log(item.uid, `selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
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
				relativePath: (f as any).webkitRelativePath || f.name,
			}));
			el.value = '';

			folderDialogFiles = pickedFiles;
			folderDialogLoading = false;
		}, 50);
	}

	function onFolderConfirm(selected: { file: File; relativePath: string }[]) {
		folderDialogOpen = false;

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
			log(item.uid, `selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
		}
		void startUploadQueue();
	}

	// --- Upload queue engine ---

	let uploadQueueRunning = false;

	async function startUploadQueue() {
		if (uploadQueueRunning) return;
		uploadQueueRunning = true;

		try {
			while (true) {
				const idx = uploadItems.findIndex((i) => i.phase === 'hashing' || i.phase === 'pending');
				if (idx === -1) break;

				const item = uploadItems[idx];
				log(item.uid, `picked from queue, phase=${item.phase}`);

				try {
					const t0 = Date.now();
					const totalChunks = Math.ceil(item.fileSize / CHUNK_SIZE);
					let hashPromise: ReturnType<typeof computeSHA256Chunked> | null = null;

					if (item.phase === 'hashing') {
						log(item.uid, `computing hash, totalChunks=${totalChunks}`);

						hashPromise = computeSHA256Chunked(item.file, {
							onPreHash: (preHash) => {
								item.preHash = preHash;
								log(item.uid, `preHash ready (${Date.now() - t0}ms): ${preHash}`);
							},
							onProgress: (percent) => {
								item.hashProgress = percent;
							}
						});

						while (!item.preHash) {
							await new Promise(r => setTimeout(r, 10));
						}

						const preResult = await preCheck(item.preHash, item.fileSize);
						log(item.uid, `preCheck result: ${preResult.status}`);

						if (preResult.status === 'SUSPECT_HIT') {
							const { hash } = await hashPromise!;
							item.fileHash = hash;
							log(item.uid, `fullHash done (${Date.now() - t0}ms): ${hash}`);

							item.phase = 'verifying';
							log(item.uid, 'phase → verifying, requesting challenge...');
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
								log(item.uid, `verify result: ${verifyResult.status}`);

								if (verifyResult.status === 'HIT' && verifyResult.physicalFileSlug) {
									log(item.uid, `dedup HIT, importing from ${verifyResult.physicalFileSlug}`);

									// Show confirmation if duplicate files exist
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
											log(item.uid, 'skipped by user (duplicate)');
											continue;
										}
									}

									item.phase = 'importing';
									item.progress = 100;
									await importFile(verifyResult.physicalFileSlug, item.fileName, opts.getCurrentSlug() || undefined);
									item.phase = 'completed';
									item.uploadedBytes = item.fileSize;
									log(item.uid, 'completed (dedup)');
									continue;
								}
								log(item.uid, 'verify MISS, falling through to upload');
							} else {
								log(item.uid, `challenge status: ${challenge.status}, no challenge issued`);
							}
						} else {
							log(item.uid, 'preCheck NOT_FOUND, going straight to upload');
						}
					}

					item.phase = 'uploading';
					log(item.uid, 'phase → uploading, init session...');
					const mimeType = item.file.type || 'application/octet-stream';
					const task = await initUpload(item.fileHash, item.preHash, item.fileSize, mimeType, item.file.name);
					item.uploadSlug = task.uploadSlug;
					log(item.uid, `session created: ${task.uploadSlug}, resume from chunk ${task.completedChunks?.length || 0}`);

					item.abortCtrl = new AbortController();
					const startChunk = task.completedChunks?.length || 0;
					let uploaded = startChunk * CHUNK_SIZE;
					let lastTime = Date.now();
					let lastBytes = uploaded;

					for (let i = startChunk; i < totalChunks; i++) {
						if ((item.phase as string) === 'paused') break;

						const start = i * CHUNK_SIZE;
						const end = Math.min(start + CHUNK_SIZE, item.fileSize);
						const chunkData = await item.file.slice(start, end).arrayBuffer();

						await uploadChunk(task.uploadSlug, i, chunkData);
						log(item.uid, `chunk ${i + 1}/${totalChunks} uploaded (${fmtSize(end)}/${fmtSize(item.fileSize)})`);
						uploaded = end;
						item.uploadedBytes = uploaded;
						item.progress = Math.round((uploaded / item.fileSize) * 100);

						const now = Date.now();
						const elapsed = (now - lastTime) / 1000;
						if (elapsed >= 0.5) {
							item.speed = (uploaded - lastBytes) / elapsed;
							lastBytes = uploaded;
							lastTime = now;
						}
					}

					if ((item.phase as string) === 'paused') continue;

					if (!item.fileHash) {
						log(item.uid, 'hash not yet ready, waiting...');
						if (hashPromise) {
							const { hash } = await hashPromise;
							item.fileHash = hash;
						}
						log(item.uid, `fullHash done: ${item.fileHash}, updating...`);
						if (item.fileHash) await updateHash(task.uploadSlug, item.fileHash);
					}

					log(item.uid, 'all chunks uploaded, completing...');
					await completeUpload(task.uploadSlug);

					const completedTask = await getUploadStatus(task.uploadSlug);
					if (completedTask.physicalFileSlug) {
						log(item.uid, `importing physicalFile: ${completedTask.physicalFileSlug}`);
						item.phase = 'importing';
						await importFile(completedTask.physicalFileSlug, item.fileName, opts.getCurrentSlug() || undefined);
					}

					item.phase = 'completed';
					item.progress = 100;
					item.uploadedBytes = item.fileSize;
					item.speed = 0;
					log(item.uid, 'completed');
				} catch (e) {
					if (item.phase === 'paused') {
						log(item.uid, 'paused during operation, skipping');
						continue;
					}
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
					log(item.uid, `FAILED: ${item.errorMsg}`, e);
				} finally {
					item.abortCtrl = null;
				}
			}
		} finally {
			uploadQueueRunning = false;
		}

		if (uploadItems.some((i) => i.phase === 'completed')) {
			await opts.onCompleted();
		}
	}

	// --- Upload controls ---

	function pauseUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || item.phase !== 'uploading') return;
		log(uid, 'paused by user');
		item.phase = 'paused';
		item.speed = 0;
		item.abortCtrl?.abort();
		item.abortCtrl = null;
	}

	function resumeUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || (item.phase !== 'paused' && item.phase !== 'failed')) return;
		log(uid, `resumed by user (from ${item.phase})`);
		item.errorMsg = null;
		item.phase = 'pending';
		item.progress = 0;
		item.uploadedBytes = 0;
		item.speed = 0;
		void startUploadQueue();
	}

	function deleteUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (item) log(uid, `deleted by user (phase=${item.phase})`);
		if (item?.phase === 'uploading') {
			item.abortCtrl?.abort();
		}
		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	function clearCompleted() {
		uploadItems = uploadItems.filter((i) => i.phase !== 'completed');
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
	};
}
