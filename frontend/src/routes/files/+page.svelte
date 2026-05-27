<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { getAccessToken } from '$lib/api/client';
	import {
		listFiles, mkdir, trashFile, renameFile, moveFile, setStarred,
		downloadUrl, getBreadcrumb, type FileItem, type BreadcrumbItem
	} from '$lib/api/files';
	import {
		preCheck, requestChallenge, verify as verifyUpload,
		initUpload, uploadChunk, completeUpload, getUploadStatus, updateHash
	} from '$lib/api/upload';
	import { importFile } from '$lib/api/files';
	import { addToLibrary } from '$lib/api/media';
	import {
		Upload, AlertCircle, FolderPlus, ChevronRight,
		LayoutGrid, LayoutList, Search, Home, CheckCircle
	} from '@lucide/svelte';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import FileListView from '$lib/components/files/FileListView.svelte';
	import UploadPanel from '$lib/components/files/UploadPanel.svelte';
	import { confirmDelete, promptInput, confirmAction } from '$lib/dialog';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 50;
	const CHUNK_SIZE = 4 * 1024 * 1024; // 4MB — must match backend chunkSize
	const MAX_CONCURRENCY = 3;

	type UploadItem = {
		uid: string;
		file: File;
		fileName: string;
		fileSize: number;
		fileHash: string;
		preHash: string;
		phase: 'hashing' | 'pending' | 'verifying' | 'uploading' | 'paused' | 'importing' | 'completed' | 'failed';
		progress: number;
		hashProgress: number;
		uploadedBytes: number;
		speed: number;
		uploadSlug: string | null;
		abortCtrl: AbortController | null;
		errorMsg: string | null;
	};

	let files = $state<FileItem[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let loadingMore = $state(false);
	let error = $state<string | null>(null);
	let successMsg = $state<string | null>(null);
	let fileInput: HTMLInputElement | undefined = $state();
	let deleting = $state(false);
	let previewFile = $state<{ slug: string; name: string; mimeType: string; size: number } | null>(null);

	let uploadItems = $state<UploadItem[]>([]);

	let uidCounter = 0;
	function nextUid() {
		return `upload-${++uidCounter}`;
	}

	let searchQuery = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let refreshId = 0;

	type ViewMode = 'list' | 'grid';
	let viewMode = $state<ViewMode>(
		browser ? (localStorage.getItem('nd.files.view') as ViewMode) || 'list' : 'list'
	);
	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		if (browser) localStorage.setItem('nd.files.view', mode);
	}

	let currentSlug = $state<string | undefined>(undefined);
	let dirStack = $state<BreadcrumbItem[]>([]);
	let breadExpanded = $state(false);

	function buildDirUrl(slug: string | undefined): string {
		const url = new URL(page.url);
		if (slug) {
			url.searchParams.set('dir', slug);
		} else {
			url.searchParams.delete('dir');
		}
		return url.pathname + url.search;
	}

	function navigateToDir(slug: string) {
		searchQuery = '';
		void goto(buildDirUrl(slug), { keepFocus: true, noScroll: true });
	}

	async function fetchBreadcrumb(dirSlug: string | undefined) {
		if (!dirSlug) {
			dirStack = [];
			return;
		}
		try {
			dirStack = await getBreadcrumb(dirSlug);
		} catch {
			dirStack = [{ slug: dirSlug, file_name: dirSlug }];
		}
	}

	afterNavigate(async () => {
		breadExpanded = false;
		const dirParam = page.url.searchParams.get('dir');
		currentSlug = dirParam || undefined;
		await fetchBreadcrumb(currentSlug);
		void refresh();
	});

	function navigateUp() {
		if (dirStack.length <= 1) return;
		const parentSlug = dirStack[dirStack.length - 2].slug;
		searchQuery = '';
		void goto(buildDirUrl(parentSlug), { keepFocus: true, noScroll: true });
	}

	async function createDir() {
		const name = await promptInput(m.new_folder(), m.enter_folder_name(), undefined, 100);
		if (!name) return;
		try {
			await mkdir(name.trim(), currentSlug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.create_dir_failed();
		}
	}

	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void refresh(), 300);
	}

	async function refresh() {
		if (!$user) return;
		const id = ++refreshId;
		loading = true;
		error = null;
		loadingMore = false;
		try {
			const data = await listFiles(currentSlug, 1, PAGE_SIZE);
			if (id !== refreshId) return;
			files = data.files;
			total = data.total;
		} catch (e) {
			if (id !== refreshId) return;
			error = e instanceof Error ? e.message : m.load_failed();
		} finally {
			if (id === refreshId) loading = false;
		}
	}

	onDestroy(() => {
		clearTimeout(searchTimer);
	});

	async function loadMore() {
		if (loadingMore) return;
		loadingMore = true;
		const id = ++refreshId;
		try {
			const page_num = Math.floor(files.length / PAGE_SIZE) + 1;
			const data = await listFiles(currentSlug, page_num, PAGE_SIZE);
			if (id !== refreshId) return;
			files = [...files, ...data.files];
		} catch (e) {
			if (id !== refreshId) return;
			error = e instanceof Error ? e.message : m.load_more_failed();
		} finally {
			if (id === refreshId) loadingMore = false;
		}
	}

	async function computeSHA256(
		file: File,
		callbacks: { onPreHash?: (hash: string) => void; onProgress?: (percent: number) => void }
	): Promise<{ preHash: string; hash: string; totalChunks: number }> {
		const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

		return new Promise((resolve, reject) => {
			let preHash = '';
			const worker = new Worker(
				new URL('$lib/workers/sha256.worker.ts', import.meta.url),
				{ type: 'module' }
			);
			worker.onmessage = (e: MessageEvent) => {
				if (e.data.type === 'pre_hash') {
					preHash = e.data.hash;
					callbacks.onPreHash?.(preHash);
				} else if (e.data.type === 'progress') {
					callbacks.onProgress?.(e.data.percent);
				} else if (e.data.type === 'complete') {
					resolve({ preHash, hash: e.data.hash, totalChunks });
					worker.terminate();
				}
			};
			worker.onerror = () => {
				reject(new Error('SHA-256 computation failed'));
				worker.terminate();
			};

			(async () => {
				for (let i = 0; i < totalChunks; i++) {
					const start = i * CHUNK_SIZE;
					const end = Math.min(start + CHUNK_SIZE, file.size);
					const buf = await file.slice(start, end).arrayBuffer();
					worker.postMessage({ type: 'chunk', index: i, data: buf }, [buf]);
				}
				worker.postMessage({ type: 'done', totalChunks });
			})();
		});
	}

	async function onPick(e: Event) {
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
			abortCtrl: null,
			errorMsg: null
		}));

		uploadItems = [...uploadItems, ...newItems];
		for (const item of newItems) {
			console.log(`[upload:${item.uid}] selected: ${item.fileName} (${fmtSize(item.fileSize)})`);
		}

		void startUploadQueue();
	}

	let uploadQueueRunning = false;

	async function startUploadQueue() {
		if (uploadQueueRunning) return;
		uploadQueueRunning = true;

		try {
			while (true) {
				const idx = uploadItems.findIndex((i) => i.phase === 'hashing' || i.phase === 'pending');
				if (idx === -1) break;

				const item = uploadItems[idx];

				try {
					if (item.phase === 'hashing') {
						console.log(`[upload:${item.uid}] computing hash (pre_hash + full)...`);
						const t0 = Date.now();
						const totalChunks = Math.ceil(item.fileSize / CHUNK_SIZE);

						// Start computing hash — pre_hash arrives quickly, full hash runs in background
						const hashPromise = computeSHA256(item.file, {
							onPreHash: (preHash) => {
								item.preHash = preHash;
								console.log(`[upload:${item.uid}] pre_hash ready (${Date.now() - t0}ms): ${preHash}`);
							},
							onProgress: (percent) => {
								item.hashProgress = percent;
							}
						});

						// Wait for pre_hash, then immediately start upload flow
						// We need to wait just enough for pre_hash to be set by the callback
						// Since pre_hash comes from the first 512KB (first chunk), wait for it
						while (!item.preHash) {
							await new Promise(r => setTimeout(r, 10));
						}

						// PreCheck with pre_hash — fast dedup check
						console.log(`[upload:${item.uid}] pre-check with pre_hash...`);
						const preResult = await preCheck(item.preHash, item.fileSize);

						if (preResult.status === 'SUSPECT_HIT') {
							// Potential duplicate — wait for full hash and verify
							console.log(`[upload:${item.uid}] pre-check SUSPECT_HIT, waiting for full hash...`);
							const { hash } = await hashPromise;
							item.fileHash = hash;
							console.log(`[upload:${item.uid}] full hash done (${Date.now() - t0}ms): ${hash}`);

							item.phase = 'verifying';
							const challenge = await requestChallenge(hash);

							if (challenge.status === 'CHALLENGE') {
								const sampleStart = challenge.challenge_offset;
								const sampleEnd = Math.min(sampleStart + 1024, item.fileSize);
								const sampleBlob = item.file.slice(sampleStart, sampleEnd);
								const sampleBuffer = await sampleBlob.arrayBuffer();
								const sampleBytes = new Uint8Array(sampleBuffer);

								const tokenBytes = new TextEncoder().encode(challenge.challenge_token);
								const proofInput = new Uint8Array(sampleBytes.length + tokenBytes.length);
								proofInput.set(sampleBytes);
								proofInput.set(tokenBytes, sampleBytes.length);
								const proofHash = await crypto.subtle.digest('SHA-256', proofInput);
								const proofCode = Array.from(new Uint8Array(proofHash))
									.map(b => b.toString(16).padStart(2, '0'))
									.join('');

								const verifyResult = await verifyUpload(hash, proofCode);
								console.log(`[upload:${item.uid}] verify result: ${verifyResult.status}`);

								if (verifyResult.status === 'HIT' && verifyResult.physical_file_slug) {
									console.log(`[upload:${item.uid}] dedup HIT, importing...`);
									item.phase = 'importing';
									item.progress = 100;
									await importFile(verifyResult.physical_file_slug, item.fileName, currentSlug);
									item.phase = 'completed';
									item.uploadedBytes = item.fileSize;
									console.log(`[upload:${item.uid}] import done (rapid)`);
									continue;
								}
							}
							// MISS — fall through to upload
						}

						// NOT_FOUND or MISS — start uploading immediately
						item.phase = 'uploading';
						const mimeType = item.file.type || 'application/octet-stream';
						console.log(`[upload:${item.uid}] init upload task (hash=${item.fileHash || '(pending)'})...`);
						const task = await initUpload(item.fileHash, item.preHash, item.fileSize, mimeType);
						item.uploadSlug = task.upload_slug;

						item.abortCtrl = new AbortController();
						const startChunk = task.completed_chunks?.length || 0;
						let uploaded = startChunk * CHUNK_SIZE;
						console.log(`[upload:${item.uid}] uploading ${totalChunks} chunks (resume from chunk ${startChunk})`);
						let lastTime = Date.now();
						let lastBytes = uploaded;

						// Upload chunks
						for (let i = startChunk; i < totalChunks; i++) {
							if ((item.phase as string) === 'paused') break;

							const start = i * CHUNK_SIZE;
							const end = Math.min(start + CHUNK_SIZE, item.fileSize);
							const chunkData = await item.file.slice(start, end).arrayBuffer();

							await uploadChunk(task.upload_slug, i, chunkData);
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

						// If hash wasn't set yet (NOT_FOUND path), wait for it and update
						if (!item.fileHash) {
							console.log(`[upload:${item.uid}] waiting for full hash before complete...`);
							const { hash } = await hashPromise;
							item.fileHash = hash;
							console.log(`[upload:${item.uid}] full hash done: ${hash}`);
							await updateHash(task.upload_slug, hash);
						}

						console.log(`[upload:${item.uid}] all chunks uploaded, completing...`);
						await completeUpload(task.upload_slug);

						const completedTask = await getUploadStatus(task.upload_slug);
						if (completedTask.physical_file_slug) {
							console.log(`[upload:${item.uid}] importing file...`);
							item.phase = 'importing';
							await importFile(completedTask.physical_file_slug, item.fileName, currentSlug);
						}

						item.phase = 'completed';
						item.progress = 100;
						item.uploadedBytes = item.fileSize;
						item.speed = 0;
						console.log(`[upload:${item.uid}] completed`);
					}
				} catch (e) {
					if (item.phase === 'paused') continue;
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
					console.error(`[upload:${item.uid}] failed:`, item.errorMsg, e);
				} finally {
					item.abortCtrl = null;
				}
			}
		} finally {
			uploadQueueRunning = false;
		}

		if (uploadItems.some((i) => i.phase === 'completed')) {
			error = null;
			await refresh();
		}
	}

	function pauseUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || item.phase !== 'uploading') return;
		item.phase = 'paused';
		item.speed = 0;
		item.abortCtrl?.abort();
		item.abortCtrl = null;
	}

	function resumeUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || (item.phase !== 'paused' && item.phase !== 'failed')) return;
		item.errorMsg = null;
		item.phase = 'pending';
		item.progress = 0;
		item.uploadedBytes = 0;
		item.speed = 0;
		void startUploadQueue();
	}

	function deleteUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (item?.phase === 'uploading') {
			item.abortCtrl?.abort();
		}
		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	function clearCompleted() {
		uploadItems = uploadItems.filter((i) => i.phase !== 'completed');
	}

	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
		deleting = true;
		error = null;
		try {
			await trashFile(slug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.delete_failed();
		} finally {
			deleting = false;
		}
	}

	async function rename(slug: string, currentName: string) {
		const newName = await promptInput(m.rename(), m.enter_new_name(), currentName, 100);
		if (!newName || newName === currentName) return;
		error = null;
		try {
			await renameFile(slug, newName);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.rename_failed();
		}
	}

	async function toggleStar(slug: string, currentlyStarred: boolean) {
		try {
			await setStarred(slug, !currentlyStarred);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.unstar_failed();
		}
	}

	function onPreview(file: FileItem) {
		previewFile = { slug: file.slug, name: file.file_name, mimeType: file.mime_type || '', size: file.file_size };
	}

	async function onAddToMedia(file: FileItem) {
		try {
			const resp = await addToLibrary(file.slug);
			if (resp.transcode_status === 'existing') {
				successMsg = m.media_already_in_library();
			} else if (resp.transcode_reused) {
				successMsg = m.media_add_success();
			} else {
				successMsg = m.media_add_success();
			}
			setTimeout(() => { successMsg = null; }, 5000);
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_add_failed();
		}
	}

	onMount(() => {
		if (!$user) void goto('/login');
	});
</script>

{#if !$authReady}
{:else if $user}
	<div class="space-y-4">
		<!-- Breadcrumb -->
		{#if currentSlug}
		<div class="flex items-center gap-1 overflow-hidden text-sm">
			<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(undefined), { keepFocus: true, noScroll: true }); }} class="shrink-0 rounded p-1 text-gray-500 transition-colors hover:text-gray-900">
				<Home size={16} />
			</button>
			{#each dirStack as crumb, i}
				{#if !breadExpanded && dirStack.length > 3 && i > 0 && i < dirStack.length - 1}
					{#if i === 1}
						<ChevronRight size={14} class="shrink-0 text-gray-300" />
						<button
							type="button"
							onclick={() => (breadExpanded = true)}
							class="shrink-0 rounded px-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
						>...</button>
					{/if}
				{:else}
					<ChevronRight size={14} class="shrink-0 text-gray-300" />
					{#if i === dirStack.length - 1}
						<span class="max-w-48 truncate font-medium text-gray-900 sm:max-w-64 md:max-w-80">{crumb.file_name}</span>
					{:else}
						<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(crumb.slug), { keepFocus: true, noScroll: true }); }} class="max-w-32 shrink truncate rounded px-1 text-gray-500 transition-colors hover:text-gray-900 sm:max-w-40">{crumb.file_name}</button>
					{/if}
				{/if}
			{/each}
		</div>
		{/if}

		<!-- Actions -->
		<div class="flex items-center justify-between gap-2">
			<div class="relative">
				<Search size={15} class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
				<input
					type="search"
					placeholder="{m.search_files()}..."
					bind:value={searchQuery}
					oninput={onSearchInput}
					class="h-8 w-48 rounded-lg border border-gray-200 bg-gray-50 pl-8 pr-2.5 text-sm text-gray-700 outline-none transition-colors placeholder:text-gray-400 hover:border-gray-300 focus:border-blue-400 focus:bg-white focus:ring-2 focus:ring-blue-50"
				/>
			</div>

			<div class="flex items-center gap-2">
				<div class="flex overflow-hidden rounded-lg border border-gray-200">
					<button type="button" onclick={() => setViewMode('list')} class="p-1.5 transition-colors {viewMode === 'list' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
						<LayoutList size={15} />
					</button>
					<button type="button" onclick={() => setViewMode('grid')} class="p-1.5 transition-colors {viewMode === 'grid' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
						<LayoutGrid size={15} />
					</button>
				</div>

				<button type="button" onclick={createDir} class="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50">
					<FolderPlus size={15} /> {m.new_folder()}
				</button>

				<button type="button" onclick={() => fileInput?.click()} class="flex h-8 items-center gap-1.5 rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 active:bg-blue-800">
					<Upload size={15} /> {m.upload_files()}
				</button>
				<input bind:this={fileInput} type="file" multiple class="hidden" onchange={onPick} />
			</div>
		</div>

		<!-- Error -->
		{#if error}
			<div class="flex items-start gap-2.5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">
				<AlertCircle size={16} class="mt-0.5 shrink-0" />
				<span>{error}</span>
			</div>
		{/if}

		<!-- Success -->
		{#if successMsg}
			<div class="flex items-start gap-2.5 rounded-lg border border-green-200 bg-green-50 px-3.5 py-2.5 text-sm text-green-700">
				<CheckCircle size={16} class="mt-0.5 shrink-0" />
				<span>{successMsg}</span>
				<a href="/media" class="ml-auto shrink-0 text-green-600 underline hover:text-green-700">{m.back_to_media()}</a>
			</div>
		{/if}

		<!-- File list -->
		<FileListView
			{files}
			{viewMode}
			{loading}
			{deleting}
			emptyMessage={currentSlug ? m.dir_empty() : m.no_files()}
			onNavigateDir={navigateToDir}
			onStar={toggleStar}
			onPreview={onPreview}
			onRename={rename}
			onDelete={remove}
			onAddToMedia={onAddToMedia}
		/>

		{#if files.length > 0}
			<div class="flex items-center justify-between text-xs text-gray-400">
				<span>{m.total_files({ total })}</span>
				{#if files.length < total}
					<button type="button" onclick={loadMore} disabled={loadingMore} class="text-gray-500 transition-colors hover:text-gray-700 disabled:opacity-50">
						{loadingMore ? m.loading() : m.load_more()}
					</button>
				{/if}
			</div>
		{/if}
	</div>
{:else}
	<p class="text-gray-600">{@html m.please_login({ link: `<a href="/login" class="text-blue-600 underline hover:text-blue-700">${m.login_link_text()}</a>` })}</p>
{/if}

{#if previewFile}
	<DrivePreview
		id={previewFile.slug}
		name={previewFile.name}
		mimeType={previewFile.mimeType}
		size={previewFile.size}
		open={true}
		close={() => (previewFile = null)}
	/>
{/if}

<UploadPanel
	items={uploadItems}
	onPause={pauseUpload}
	onResume={resumeUpload}
	onDelete={deleteUpload}
	onClear={clearCompleted}
/>
