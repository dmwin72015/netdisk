<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { getAccessToken } from '$lib/api/client';
	import {
		listFiles, mkdir, trashFile, renameFile, moveFile, setStarred,
		downloadUrl, type FileItem
	} from '$lib/api/files';
	import {
		preCheck, requestChallenge, verify as verifyUpload,
		initUpload, uploadChunk, completeUpload, getUploadStatus,
		type UploadTask
	} from '$lib/api/upload';
	import { importFile } from '$lib/api/files';
	import {
		Upload, Download, Trash2, AlertCircle, Eye, CheckCircle2, XCircle,
		Loader2, FolderPlus, ChevronRight, Pencil, LayoutGrid, LayoutList,
		Search, Home, ChevronUp, ChevronDown, X, Pause, Play, Star
	} from '@lucide/svelte';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import { confirmDelete, promptInput, confirmAction } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 50;
	const CHUNK_SIZE = 8 * 1024 * 1024; // 8MB
	const MAX_CONCURRENCY = 3;

	type UploadItem = {
		uid: string;
		file: File;
		fileName: string;
		fileSize: number;
		fileHash: string;
		phase: 'hashing' | 'pending' | 'prechecking' | 'verifying' | 'uploading' | 'paused' | 'importing' | 'completed' | 'failed';
		progress: number;
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
	let fileInput: HTMLInputElement | undefined = $state();
	let deleting = $state(false);
	let previewFile = $state<{ slug: string; name: string; mimeType: string; size: number } | null>(null);

	let uploadItems = $state<UploadItem[]>([]);
	let uploadPanelOpen = $state(true);

	let completedCount = $derived(uploadItems.filter((i) => i.phase === 'completed').length);
	let failedCount = $derived(uploadItems.filter((i) => i.phase === 'failed').length);
	let totalSpeed = $derived(uploadItems.reduce((s, i) => s + i.speed, 0));
	let hasActiveUploads = $derived(uploadItems.length > 0 && uploadItems.some((i) => i.phase !== 'completed' && i.phase !== 'failed'));

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
	type DirBreadcrumb = { slug: string | undefined; name: string };
	let dirStack = $state<DirBreadcrumb[]>([{ slug: undefined, name: 'All Files' }]);
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

	function navigateToDir(slug: string, name: string) {
		searchQuery = '';
		void goto(buildDirUrl(slug), { keepFocus: true, noScroll: true });
	}

	afterNavigate(async () => {
		breadExpanded = false;
		const dirParam = page.url.searchParams.get('dir');
		if (dirParam) {
			currentSlug = dirParam;
			// We don't have an ancestors API, so build breadcrumb from navigation
			// For now, just show current dir
			dirStack = [{ slug: undefined, name: 'All Files' }, { slug: dirParam, name: dirParam }];
		} else {
			currentSlug = undefined;
			dirStack = [{ slug: undefined, name: 'All Files' }];
		}
		void refresh();
	});

	function navigateUp() {
		if (dirStack.length <= 1) return;
		const parentSlug = dirStack[dirStack.length - 2].slug;
		searchQuery = '';
		void goto(buildDirUrl(parentSlug), { keepFocus: true, noScroll: true });
	}

	async function createDir() {
		const name = await promptInput('New Folder', 'Enter folder name');
		if (!name) return;
		try {
			await mkdir(name.trim(), currentSlug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create folder';
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
			error = e instanceof Error ? e.message : 'Failed to load files';
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
			error = e instanceof Error ? e.message : 'Failed to load more';
		} finally {
			if (id === refreshId) loadingMore = false;
		}
	}

	async function computeSHA256(file: File): Promise<{ hash: string; totalChunks: number }> {
		return new Promise((resolve, reject) => {
			const worker = new Worker(
				new URL('$lib/workers/sha256.worker.ts', import.meta.url),
				{ type: 'module' }
			);
			worker.onmessage = (e: MessageEvent) => {
				if (e.data.type === 'complete') {
					resolve({ hash: e.data.hash, totalChunks: e.data.totalChunks });
				}
				worker.terminate();
			};
			worker.onerror = (e) => {
				reject(new Error('SHA-256 computation failed'));
				worker.terminate();
			};
			worker.postMessage(file);
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
			phase: 'hashing',
			progress: 0,
			uploadedBytes: 0,
			speed: 0,
			uploadSlug: null,
			abortCtrl: null,
			errorMsg: null
		}));

		uploadItems = [...uploadItems, ...newItems];
		uploadPanelOpen = true;

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
					// Step 1: Compute SHA-256
					if (item.phase === 'hashing') {
						const { hash, totalChunks } = await computeSHA256(item.file);
						item.fileHash = hash;
						item.phase = 'prechecking';

						// Step 2: Pre-check (does file exist in system?)
						const preResult = await preCheck(hash, item.fileSize);

						if (preResult.status === 'HIT' && preResult.physical_file_slug) {
							// File already exists in system — instant import
							item.phase = 'importing';
							item.progress = 100;
							await importFile(preResult.physical_file_slug, item.fileName, currentSlug);
							item.phase = 'completed';
							item.uploadedBytes = item.fileSize;
							continue;
						}

						// Step 3: Request challenge for verification
						item.phase = 'verifying';
						const challenge = await requestChallenge(hash, item.fileSize, totalChunks);

						// Read sample at offset
						const sampleStart = challenge.offset;
						const sampleEnd = Math.min(sampleStart + 1024, item.fileSize);
						const sampleBlob = item.file.slice(sampleStart, sampleEnd);
						const sampleBuffer = await sampleBlob.arrayBuffer();
						const sampleBytes = new Uint8Array(sampleBuffer);
						const sampleHex = Array.from(sampleBytes).map(b => b.toString(16).padStart(2, '0')).join('');

						// Verify
						const verifyResult = await verifyUpload(
							hash, item.fileSize, challenge.challenge_token, challenge.offset, sampleHex
						);

						if (verifyResult.status === 'OK' && verifyResult.physical_file_slug) {
							// Verified — instant import without upload
							item.phase = 'importing';
							item.progress = 100;
							await importFile(verifyResult.physical_file_slug, item.fileName, currentSlug);
							item.phase = 'completed';
							item.uploadedBytes = item.fileSize;
							continue;
						}

						// Step 4: Need to actually upload
						item.phase = 'uploading';
						const task = await initUpload(hash, item.fileSize, totalChunks, item.fileName);
						item.uploadSlug = task.upload_slug;

						// Upload chunks
						item.abortCtrl = new AbortController();
						const startChunk = task.chunks_uploaded || 0;
						let uploaded = startChunk * CHUNK_SIZE;

						for (let i = startChunk; i < totalChunks; i++) {
							if ((item.phase as string) === 'paused') break;

							const start = i * CHUNK_SIZE;
							const end = Math.min(start + CHUNK_SIZE, item.fileSize);
							const chunkData = await item.file.slice(start, end).arrayBuffer();

							await uploadChunk(task.upload_slug, i, chunkData);
							uploaded = end;
							item.uploadedBytes = uploaded;
							item.progress = Math.round((uploaded / item.fileSize) * 100);
						}

						if ((item.phase as string) === 'paused') continue;

						// Complete upload
						await completeUpload(task.upload_slug);

						// Get the physical file slug from completed task
						const completedTask = await getUploadStatus(task.upload_slug);
						if (completedTask.physical_file_slug) {
							item.phase = 'importing';
							await importFile(completedTask.physical_file_slug, item.fileName, currentSlug);
						}

						item.phase = 'completed';
						item.progress = 100;
						item.uploadedBytes = item.fileSize;
						item.speed = 0;
					}
				} catch (e) {
					if (item.phase === 'paused') continue;
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : 'Upload failed';
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
		if (!(await confirmDelete(`Delete "${name}"?`))) return;
		deleting = true;
		error = null;
		try {
			await trashFile(slug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
		} finally {
			deleting = false;
		}
	}

	async function rename(slug: string, currentName: string) {
		const newName = await promptInput('Rename', 'Enter new name', currentName);
		if (!newName || newName === currentName) return;
		error = null;
		try {
			await renameFile(slug, newName);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Rename failed';
		}
	}

	async function toggleStar(slug: string, currentlyStarred: boolean) {
		try {
			await setStarred(slug, !currentlyStarred);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update star';
		}
	}

	function getDownloadHref(slug: string): string {
		return downloadUrl(slug);
	}

	function isDir(f: FileItem): boolean {
		return f.is_dir;
	}

	function fmtSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return (bytes / Math.pow(k, i)).toFixed(i > 0 ? 1 : 0) + ' ' + sizes[i];
	}

	function fmtSpeed(bytesPerSec: number): string {
		if (bytesPerSec <= 0) return '';
		if (bytesPerSec >= 1024 * 1024) return (bytesPerSec / (1024 * 1024)).toFixed(1) + ' MB/s';
		if (bytesPerSec >= 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
		return bytesPerSec + ' B/s';
	}

	onMount(() => {
		if (!$user) void goto('/login');
	});
</script>

{#if !$authReady}
	<!-- Wait for client-side auth check -->
{:else if $user}
	<div class="space-y-4">
		<!-- Breadcrumb -->
		{#if dirStack.length > 1}
		<div class="flex items-center gap-1 overflow-hidden text-sm">
			{#each dirStack as crumb, i}
				{#if !breadExpanded && dirStack.length > 4 && i > 0 && i < dirStack.length - 2}
					{#if i === 1}
						<ChevronRight size={14} class="shrink-0 text-gray-300" />
						<button
							type="button"
							onclick={() => (breadExpanded = true)}
							class="shrink-0 rounded px-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
						>...</button>
					{/if}
				{:else}
					{#if i > 0}
						<ChevronRight size={14} class="shrink-0 text-gray-300" />
					{/if}
					{#if i === dirStack.length - 1}
						<span class="max-w-48 truncate font-medium text-gray-900 sm:max-w-64 md:max-w-80">{crumb.name}</span>
					{:else if i === 0}
						<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(crumb.slug), { keepFocus: true, noScroll: true }); }} class="shrink-0 rounded p-1 text-gray-500 transition-colors hover:text-gray-900">
							<Home size={16} />
						</button>
					{:else}
						<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(crumb.slug), { keepFocus: true, noScroll: true }); }} class="max-w-32 shrink truncate rounded px-1 text-gray-500 transition-colors hover:text-gray-900 sm:max-w-40">{crumb.name}</button>
					{/if}
				{/if}
			{/each}
		</div>
		{/if}

		<!-- Actions -->
		<div class="flex items-center justify-end gap-2">
			<div class="relative">
				<Search size={15} class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
				<input
					type="search"
					placeholder="Search files..."
					bind:value={searchQuery}
					oninput={onSearchInput}
					class="h-8 w-48 rounded-lg border border-gray-200 bg-gray-50 pl-8 pr-2.5 text-sm text-gray-700 outline-none transition-colors placeholder:text-gray-400 hover:border-gray-300 focus:border-blue-400 focus:bg-white focus:ring-2 focus:ring-blue-50"
				/>
			</div>

			<div class="flex overflow-hidden rounded-lg border border-gray-200">
				<button type="button" onclick={() => setViewMode('list')} class="p-1.5 transition-colors {viewMode === 'list' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
					<LayoutList size={15} />
				</button>
				<button type="button" onclick={() => setViewMode('grid')} class="p-1.5 transition-colors {viewMode === 'grid' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
					<LayoutGrid size={15} />
				</button>
			</div>

			<button type="button" onclick={createDir} class="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50">
				<FolderPlus size={15} /> New Folder
			</button>

			<button type="button" onclick={() => fileInput?.click()} class="flex h-8 items-center gap-1.5 rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 active:bg-blue-800">
				<Upload size={15} /> Upload
			</button>
			<input bind:this={fileInput} type="file" multiple class="hidden" onchange={onPick} />
		</div>

		<!-- Error -->
		{#if error}
			<div class="flex items-start gap-2.5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">
				<AlertCircle size={16} class="mt-0.5 shrink-0" />
				<span>{error}</span>
			</div>
		{/if}

		<!-- File list -->
		{#if loading && files.length === 0}
			<div class="flex items-center justify-center py-16">
				<Loader2 size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if files.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
				<FolderPlus size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{currentSlug ? 'Folder is empty' : 'No files yet'}</p>
			</div>
		{:else if viewMode === 'grid'}
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
				{#each files as f (f.slug)}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="group relative flex flex-col items-center rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-200 hover:shadow-md {isDir(f) ? 'cursor-pointer' : ''}"
						onclick={isDir(f) ? () => navigateToDir(f.slug, f.file_name) : undefined}
					>
						<div class="absolute right-1.5 top-1.5 flex gap-0.5 rounded-lg bg-white/90 opacity-0 shadow-sm backdrop-blur transition-opacity group-hover:opacity-100" onclick={(e) => e.stopPropagation()}>
							<button type="button" onclick={() => toggleStar(f.slug, f.is_starred)} class="rounded-md p-1 transition-colors {f.is_starred ? 'text-amber-400' : 'text-gray-400 hover:text-amber-400'}">
								<Star size={14} fill={f.is_starred ? 'currentColor' : 'none'} />
							</button>
							{#if !isDir(f)}
								<button type="button" onclick={() => (previewFile = { slug: f.slug, name: f.file_name, mimeType: f.mime_type || '', size: f.file_size })} class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700">
									<Eye size={14} />
								</button>
								<a href={getDownloadHref(f.slug)} download={f.file_name} class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700">
									<Download size={14} />
								</a>
							{/if}
							<button type="button" onclick={() => rename(f.slug, f.file_name)} class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700">
								<Pencil size={14} />
							</button>
							<button type="button" onclick={() => remove(f.slug, f.file_name)} disabled={deleting} class="rounded-md p-1 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 disabled:opacity-30">
								<Trash2 size={14} />
							</button>
						</div>
						<MimeIcon mimeType={f.mime_type} isDir={isDir(f)} size={36} />
						<p class="mt-3 w-full truncate text-center text-sm font-medium text-gray-700" title={f.file_name}>{f.file_name}</p>
						<p class="mt-0.5 text-xs text-gray-400">{isDir(f) ? '' : fmtSize(f.file_size)}</p>
					</div>
				{/each}
			</div>
		{:else}
			<div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
				<table class="w-full table-fixed text-sm">
					<thead>
						<tr class="border-b border-gray-100 text-left text-xs text-gray-400">
							<th class="w-[40%] px-4 py-2.5 font-medium">Name</th>
							<th class="w-[15%] px-4 py-2.5 font-medium">Type</th>
							<th class="w-[10%] px-4 py-2.5 text-right font-medium">Size</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">Modified</th>
							<th class="w-[20%] px-4 py-2.5 text-right font-medium">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each files as f (f.slug)}
							<tr class="border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80 {isDir(f) ? 'cursor-pointer' : ''}" onclick={isDir(f) ? () => navigateToDir(f.slug, f.file_name) : undefined}>
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-2.5">
										<span class="shrink-0"><MimeIcon mimeType={f.mime_type} isDir={isDir(f)} size={18} /></span>
										<span class="truncate text-gray-700" title={f.file_name}>{f.file_name}</span>
										{#if f.is_starred}
											<Star size={12} class="shrink-0 text-amber-400" fill="currentColor" />
										{/if}
									</div>
								</td>
								<td class="truncate px-4 py-2.5 text-xs text-gray-400">{isDir(f) ? 'Folder' : f.mime_type}</td>
								<td class="px-4 py-2.5 text-right text-gray-500">{isDir(f) ? '-' : fmtSize(f.file_size)}</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400">
									{new Date(f.updated_at).toLocaleDateString()}
								</td>
								<!-- svelte-ignore a11y_no_static_element_interactions -->
								<td class="px-4 py-2.5 text-right" onclick={(e) => e.stopPropagation()}>
									<div class="flex items-center justify-end">
										<button type="button" onclick={() => toggleStar(f.slug, f.is_starred)} class="rounded-md p-1.5 transition-colors {f.is_starred ? 'text-amber-400' : 'text-gray-400 hover:text-amber-400'}">
											<Star size={15} fill={f.is_starred ? 'currentColor' : 'none'} />
										</button>
										{#if !isDir(f)}
											<button type="button" onclick={() => (previewFile = { slug: f.slug, name: f.file_name, mimeType: f.mime_type || '', size: f.file_size })} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
												<Eye size={15} />
											</button>
											<a href={getDownloadHref(f.slug)} download={f.file_name} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
												<Download size={15} />
											</a>
										{/if}
										<button type="button" onclick={() => rename(f.slug, f.file_name)} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
											<Pencil size={15} />
										</button>
										<button type="button" onclick={() => remove(f.slug, f.file_name)} disabled={deleting} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 disabled:opacity-30">
											<Trash2 size={15} />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

		{#if files.length > 0}
			<div class="flex items-center justify-between text-xs text-gray-400">
				<span>{total} file{total !== 1 ? 's' : ''}</span>
				{#if files.length < total}
					<button type="button" onclick={loadMore} disabled={loadingMore} class="text-gray-500 transition-colors hover:text-gray-700 disabled:opacity-50">
						{loadingMore ? 'Loading...' : 'Load more'}
					</button>
				{/if}
			</div>
		{/if}
	</div>
{:else}
	<p class="text-gray-600">Please <a href="/login" class="text-blue-600 underline hover:text-blue-700">login</a> to continue.</p>
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

<!-- Upload panel -->
{#if uploadItems.length > 0}
	<div class="fixed bottom-4 right-4 z-40 w-80 sm:w-96">
		{#if uploadPanelOpen}
			<div class="rounded-xl border border-gray-200 bg-white shadow-lg">
				<div class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5">
					<div class="flex items-center gap-2">
						<span class="text-sm font-medium text-gray-800">
							{hasActiveUploads ? 'Uploading' : 'Upload complete'}
						</span>
						<span class="text-xs text-gray-400">{completedCount}/{uploadItems.length}</span>
						{#if totalSpeed > 0}
							<span class="text-xs text-blue-500">{fmtSpeed(totalSpeed)}</span>
						{/if}
					</div>
					<div class="flex items-center gap-1">
						<button type="button" onclick={clearCompleted} class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600">
							<X size={14} />
						</button>
						<button type="button" onclick={() => (uploadPanelOpen = false)} class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600">
							<ChevronDown size={14} />
						</button>
					</div>
				</div>
				<div class="max-h-72 overflow-y-auto">
					{#each uploadItems as item (item.uid)}
						<div class="border-b border-gray-50 px-4 py-2 last:border-0">
							<div class="flex items-center gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm text-gray-700" title={item.fileName}>{item.fileName}</p>
									<div class="flex items-center gap-2 text-xs text-gray-400">
										<span>{fmtSize(item.fileSize)}</span>
										{#if item.phase === 'uploading' && item.speed > 0}
											<span class="text-blue-500">{fmtSpeed(item.speed)}</span>
										{/if}
									</div>
								</div>
								<div class="flex shrink-0 items-center gap-0.5">
									{#if item.phase === 'hashing'}
										<span class="text-xs text-gray-400">Hashing...</span>
										<Loader2 size={14} class="animate-spin text-gray-300" />
									{:else if item.phase === 'prechecking' || item.phase === 'verifying'}
										<span class="text-xs text-gray-400">Checking...</span>
										<Loader2 size={14} class="animate-spin text-gray-300" />
									{:else if item.phase === 'importing'}
										<span class="text-xs text-blue-500">Importing...</span>
										<Loader2 size={14} class="animate-spin text-blue-400" />
									{:else if item.phase === 'uploading'}
										<span class="mr-1 text-xs font-medium text-blue-600">{item.progress}%</span>
										<button type="button" onclick={() => pauseUpload(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-amber-500">
											<Pause size={13} />
										</button>
									{:else if item.phase === 'paused'}
										<span class="mr-1 text-xs font-medium text-amber-500">Paused</span>
										<button type="button" onclick={() => resumeUpload(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500">
											<Play size={13} />
										</button>
									{:else if item.phase === 'completed'}
										<CheckCircle2 size={14} class="text-green-500" />
									{:else if item.phase === 'failed'}
										<XCircle size={14} class="text-red-500" />
										<button type="button" onclick={() => resumeUpload(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500">
											<Play size={13} />
										</button>
									{/if}
									{#if item.phase !== 'completed' && item.phase !== 'importing'}
										<button type="button" onclick={() => deleteUpload(item.uid)} class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-red-500">
											<X size={13} />
										</button>
									{/if}
								</div>
							</div>
							{#if item.phase === 'uploading'}
								<div class="mt-1.5 h-1 overflow-hidden rounded-full bg-gray-100">
									<div class="h-full rounded-full bg-blue-500 transition-all" style="width:{item.progress}%"></div>
								</div>
							{:else if item.phase === 'paused'}
								<div class="mt-1.5 h-1 overflow-hidden rounded-full bg-gray-100">
									<div class="h-full rounded-full bg-amber-400 transition-all" style="width:{item.progress}%"></div>
								</div>
							{:else if item.phase === 'failed' && item.errorMsg}
								<p class="mt-1 text-xs text-red-500">{item.errorMsg}</p>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<button type="button" onclick={() => (uploadPanelOpen = true)} class="flex w-full items-center gap-3 rounded-xl border border-gray-200 bg-white px-4 py-2.5 shadow-lg transition-colors hover:bg-gray-50">
				{#if hasActiveUploads}
					<Loader2 size={16} class="animate-spin text-blue-500" />
				{:else}
					<CheckCircle2 size={16} class="text-green-500" />
				{/if}
				<span class="flex-1 text-sm text-gray-700">{completedCount}/{uploadItems.length}</span>
				{#if failedCount > 0}
					<span class="text-xs text-red-500">{failedCount} failed</span>
				{/if}
				<ChevronUp size={14} class="text-gray-400" />
			</button>
		{/if}
	</div>
{/if}
