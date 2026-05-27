<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { Upload, Download, Trash2, AlertCircle, Eye, CheckCircle2, XCircle, Loader2, FolderPlus, ChevronRight, Pencil, LayoutGrid, LayoutList, Search, Home, ChevronUp, ChevronDown, X, Pause, Play } from '@lucide/svelte';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import { confirmDelete, promptInput, confirmAction } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';
	import {
		listDrive,
		deleteDriveFile,
		renameDriveFile,
		getDownloadUrl,
		getDriveAncestors,
		fmtSize,
		driveChunkedUpload,
		resumeDriveUpload,
		cancelDriveUpload,
		initDriveUpload,
		uploadChunks,
		computeSHA256,
		driveCheckUpload,
		driveClaimHash,
		createDriveDir,
		type DriveFile
	} from '$lib/api/drive';

	const PAGE_SIZE = 50;
	const MAX_CONCURRENCY = 3;

	type UploadItem = {
		uid: string;
		file: File;
		fileName: string;
		fileSize: number;
		phase: 'pending' | 'uploading' | 'paused' | 'completed' | 'failed';
		progress: number;
		uploadedBytes: number;
		speed: number;
		sessionId: string | null;
		abortCtrl: AbortController | null;
		errorMsg: string | null;
	};

	let files = $state<DriveFile[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let loadingMore = $state(false);
	let error = $state<string | null>(null);
	let input: HTMLInputElement | undefined = $state();
	let deleting = $state(false);
	let previewFile = $state<{ id: string; name: string; mimeType: string; size: number } | null>(null);

	let uploadItems = $state<UploadItem[]>([]);
	let activeCount = $state(0);
	let uploadPanelOpen = $state(true);

	let completedCount = $derived(uploadItems.filter((i) => i.phase === 'completed').length);
	let failedCount = $derived(uploadItems.filter((i) => i.phase === 'failed').length);
	let totalSpeed = $derived(uploadItems.reduce((s, i) => s + i.speed, 0));

	let uidCounter = 0;
	function nextUid() {
		return `drive-upload-${++uidCounter}`;
	}

	let searchQuery = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let refreshId = 0;
	let hashQueue = Promise.resolve();

	type ViewMode = 'list' | 'grid';
	let viewMode = $state<ViewMode>(
		browser ? (localStorage.getItem('vf.drive.view') as ViewMode) || 'list' : 'list'
	);
	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		if (browser) localStorage.setItem('vf.drive.view', mode);
	}

	let currentDir = $state<string | null>(null);
	type DirBreadcrumb = { id: string | null; name: string };
	let dirStack = $state<DirBreadcrumb[]>([{ id: null, name: m.all_files() }]);
	let breadExpanded = $state(false);

	function buildDirUrl(dirId: string | null): string {
		const url = new URL(page.url);
		if (dirId) {
			url.searchParams.set('dir', dirId);
		} else {
			url.searchParams.delete('dir');
		}
		return url.pathname + url.search;
	}

	function navigateToDir(id: string, name: string) {
		searchQuery = '';
		void goto(buildDirUrl(id), { keepFocus: true, noScroll: true });
	}

	afterNavigate(async () => {
		breadExpanded = false;
		const dirParam = page.url.searchParams.get('dir');
		if (dirParam) {
			currentDir = dirParam;
			const ancestors = await loadAncestors(dirParam);
			dirStack = ancestors.length > 0
				? [{ id: null, name: m.all_files() }, ...ancestors]
				: [{ id: null, name: m.all_files() }];
		} else {
			currentDir = null;
			dirStack = [{ id: null, name: m.all_files() }];
		}
		void refresh();
	});

	function navigateUp() {
		if (dirStack.length <= 1) return;
		const parentId = dirStack[dirStack.length - 2].id;
		searchQuery = '';
		void goto(buildDirUrl(parentId), { keepFocus: true, noScroll: true });
	}

	async function loadAncestors(dirId: string): Promise<DirBreadcrumb[]> {
		try {
			const ancestors = await getDriveAncestors(dirId);
			return ancestors.map((f) => ({ id: f.id, name: f.name }));
		} catch {
			return [];
		}
	}

	async function createDir() {
		const name = await promptInput(m.new_dir(), m.enter_dir_name());
		if (!name) return;
		try {
			await createDriveDir(name.trim(), currentDir ?? undefined);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.create_dir_failed();
		}
	}

	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void refresh(), 300);
	}

	async function refresh(query?: string) {
		if (!$user) return;
		const id = ++refreshId;
		loading = true;
		error = null;
		loadingMore = false;
		try {
			const data = await listDrive(PAGE_SIZE, 0, query ?? (searchQuery || undefined), currentDir ?? undefined);
			if (id !== refreshId) return;
			files = data.items;
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
		const id = ++refreshId;
		loadingMore = true;
		try {
			const data = await listDrive(PAGE_SIZE, files.length, searchQuery || undefined, currentDir ?? undefined);
			if (id !== refreshId) return;
			files = [...files, ...data.items];
		} catch (e) {
			if (id !== refreshId) return;
			error = e instanceof Error ? e.message : m.load_more_failed();
		} finally {
			if (id === refreshId) loadingMore = false;
		}
	}

	async function onPick(e: Event) {
		const el = e.currentTarget as HTMLInputElement;
		const fileList = el?.files;
		if (!fileList || fileList.length === 0) return;

		const filesArr = Array.from(fileList);
		el.value = '';

		const newItems: UploadItem[] = filesArr.map((f) => ({
			uid: nextUid(),
			file: f,
			fileName: f.name,
			fileSize: f.size,
			phase: 'pending',
			progress: 0,
			uploadedBytes: 0,
			speed: 0,
			sessionId: null,
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
				const idx = uploadItems.findIndex((i) => i.phase === 'pending');
				if (idx === -1) break;

				const item = uploadItems[idx];
				const file = item.file;
				item.phase = 'uploading';
				item.abortCtrl = new AbortController();
				activeCount++;

				try {
					const mime = file.type || 'application/octet-stream';
					const dirId = currentDir;

					// Compute SHA-256 and do unified pre-upload check
					let checkResult: Awaited<ReturnType<typeof driveCheckUpload>> | null = null;
					let fileHash = '';
					hashQueue = hashQueue.then(async () => {
						if (file.size <= 200 * 1024 * 1024) {
							try {
								fileHash = await computeSHA256(file);
								checkResult = await driveCheckUpload(fileHash, file.size, file.name, mime, dirId);
							} catch {
								// Check failed — fall through to normal upload
							}
						}
					});
					await hashQueue;

					if ((item.phase as string) === 'paused') continue;

					if ((checkResult as any)?.status === 'full') {
						// If the current user already has this file, ask for confirmation
						if ((checkResult as any).own_file) {
							const confirmed = await confirmAction(
								m.confirm_duplicate_title(),
								m.confirm_duplicate_message({ name: file.name }),
								m.confirm_duplicate_btn(),
							);
							if (!confirmed) {
								// User cancelled — remove from upload list
								uploadItems = uploadItems.filter((i) => i.uid !== item.uid);
								activeCount--;
								continue;
							}
						}
						// Scenario 1: File already fully uploaded — simulate progress + claim
						if (fileHash) {
							await driveClaimHash(fileHash, file.name, mime, file.size, dirId).catch(() => {});
						}
						// Simulate upload progress for UX
						for (let pct = 0; pct <= 100; pct += 5) {
							item.progress = pct;
							item.uploadedBytes = Math.round((file.size * pct) / 100);
							await new Promise((r) => setTimeout(r, 50));
							if ((item.phase as string) === 'paused') break;
						}
						if ((item.phase as string) === 'paused') continue;
						item.phase = 'completed';
						item.progress = 100;
						item.uploadedBytes = file.size;
						item.speed = 0;
					} else {
						// Scenario 2 (partial) or 3 (none) — init session with hash
						// Reuse hash if already computed, otherwise compute now
						let sha256ForInit = fileHash;
						if (!sha256ForInit && file.size <= 200 * 1024 * 1024) {
							try { sha256ForInit = await computeSHA256(file); } catch {}
						}
						const sessionResp = await initDriveUpload(file.name, mime, file.size, dirId, sha256ForInit || undefined);

						// Handle case where Init returns already_uploaded (race condition)
						if (sessionResp.status === 'already_uploaded' && sessionResp.file_id) {
							item.phase = 'completed';
							item.progress = 100;
							item.uploadedBytes = file.size;
							item.speed = 0;
							continue;
						}

						const session = sessionResp;
						item.sessionId = session.id;

						// For partial sessions, start progress from existing offset
						const startOffset = session.received_bytes || 0;
						if (startOffset > 0) {
							item.uploadedBytes = startOffset;
							item.progress = Math.round((startOffset / file.size) * 100);
						}

						let lastBytes = startOffset;
						let lastTime = Date.now();
						await uploadChunks(file, session.id, startOffset, {
							signal: item.abortCtrl.signal,
							onProgress: (p) => {
								item.progress = Math.round((p.uploaded / p.total) * 100);
								item.uploadedBytes = p.uploaded;
								const now = Date.now();
								const dt = (now - lastTime) / 1000;
								if (dt >= 0.5) {
									item.speed = Math.round((p.uploaded - lastBytes) / dt);
									lastBytes = p.uploaded;
									lastTime = now;
								}
							}
						});
						if ((item.phase as string) === 'paused') continue;
						item.phase = 'completed';
						item.progress = 100;
						item.uploadedBytes = file.size;
						item.speed = 0;
					}
				} catch (e) {
					if ((item.phase as string) === 'paused') continue;
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
				} finally {
					item.abortCtrl = null;
					activeCount--;
				}
			}
		} finally {
			uploadQueueRunning = false;
		}

		if (uploadItems.some((i) => i.phase === 'completed')) {
			error = null;
			await refresh(searchQuery || undefined);
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

	async function resumeUpload(uid: string) {
		const idx = uploadItems.findIndex((i) => i.uid === uid);
		if (idx === -1) return;
		const item = uploadItems[idx];
		if (item.phase !== 'paused' && item.phase !== 'failed') return;

		item.errorMsg = null;

		if (item.sessionId) {
			// Resume from existing session
			item.phase = 'uploading';
			item.abortCtrl = new AbortController();
			activeCount++;
			try {
				let lastBytes = item.uploadedBytes;
				let lastTime = Date.now();
				await resumeDriveUpload(item.file, item.sessionId, {
					signal: item.abortCtrl.signal,
					onProgress: (p) => {
						item.progress = Math.round((p.uploaded / p.total) * 100);
						item.uploadedBytes = p.uploaded;
						const now = Date.now();
						const dt = (now - lastTime) / 1000;
						if (dt >= 0.5) {
							item.speed = Math.round((p.uploaded - lastBytes) / dt);
							lastBytes = p.uploaded;
							lastTime = now;
						}
					}
				});
				if ((item.phase as string) !== 'paused') {
					item.phase = 'completed';
					item.progress = 100;
					item.uploadedBytes = item.fileSize;
					item.speed = 0;
				}
			} catch (e) {
				if ((item.phase as string) !== 'paused') {
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
				}
			} finally {
				item.abortCtrl = null;
				activeCount--;
			}
		} else {
			// No session yet — restart from beginning
			item.phase = 'pending';
			item.progress = 0;
			item.uploadedBytes = 0;
			void startUploadQueue();
			return;
		}

		if (uploadItems.some((i) => i.phase === 'completed')) {
			await refresh(searchQuery || undefined);
		}
	}

	async function deleteUpload(uid: string) {
		const idx = uploadItems.findIndex((i) => i.uid === uid);
		if (idx === -1) return;
		const item = uploadItems[idx];

		if (item.phase === 'uploading') {
			item.abortCtrl?.abort();
		}
		if (item.sessionId) {
			try { await cancelDriveUpload(item.sessionId); } catch { /* best effort */ }
		}

		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	function goToDir(f: DriveFile) {
		navigateToDir(f.id, f.name);
	}

	function clearCompleted() {
		uploadItems = [];
	}

	async function remove(id: string, name: string) {
		if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
		deleting = true;
		error = null;
		try {
			await deleteDriveFile(id);
			await refresh(searchQuery || undefined);
		} catch (e) {
			error = e instanceof Error ? e.message : m.delete_failed();
		} finally {
			deleting = false;
		}
	}

	async function rename(id: string, currentName: string) {
		const newName = await promptInput(m.rename(), m.enter_new_name(), currentName);
		if (!newName || newName === currentName) return;
		error = null;
		try {
			await renameDriveFile(id, newName);
			await refresh(searchQuery || undefined);
		} catch (e) {
			error = e instanceof Error ? e.message : m.rename_failed();
		}
	}

	function isDir(f: DriveFile): boolean {
		return f.is_dir;
	}

	let hasActiveUploads = $derived(uploadItems.length > 0 && uploadItems.some((i) => i.phase !== 'completed'));

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
	<!-- Wait for client-side auth check to avoid SSR flash -->
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
							title={m.show_full_path()}
						>...</button>
					{/if}
				{:else}
					{#if i > 0}
						<ChevronRight size={14} class="shrink-0 text-gray-300" />
					{/if}
					{#if i === dirStack.length - 1}
						<span class="max-w-48 truncate font-medium text-gray-900 sm:max-w-64 md:max-w-80" title={crumb.name}>{crumb.name}</span>
					{:else if i === 0}
						<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(crumb.id), { keepFocus: true, noScroll: true }); }} class="shrink-0 rounded p-1 text-gray-500 transition-colors hover:text-gray-900" title={crumb.name}>
							<Home size={16} />
						</button>
					{:else}
						<button type="button" onclick={() => { searchQuery = ''; void goto(buildDirUrl(crumb.id), { keepFocus: true, noScroll: true }); }} class="max-w-32 shrink truncate rounded px-1 text-gray-500 transition-colors hover:text-gray-900 sm:max-w-40" title={crumb.name}>{crumb.name}</button>
					{/if}
				{/if}
			{/each}
		</div>
		{/if}

		<!-- Actions -->
		<div class="flex items-center justify-end gap-2">
				<!-- Search -->
				<div class="relative">
					<Search size={15} class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
					<input
						type="search"
						placeholder={m.search_files()}
						bind:value={searchQuery}
						oninput={onSearchInput}
						class="h-8 w-48 rounded-lg border border-gray-200 bg-gray-50 pl-8 pr-2.5 text-sm text-gray-700 outline-none transition-colors placeholder:text-gray-400 hover:border-gray-300 focus:border-blue-400 focus:bg-white focus:ring-2 focus:ring-blue-50"
					/>
				</div>

				<!-- View toggle -->
				<div class="flex overflow-hidden rounded-lg border border-gray-200">
					<button
						type="button"
						onclick={() => setViewMode('list')}
						class="p-1.5 transition-colors {viewMode === 'list' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}"
						title={m.list_view()}
					>
						<LayoutList size={15} />
					</button>
					<button
						type="button"
						onclick={() => setViewMode('grid')}
						class="p-1.5 transition-colors {viewMode === 'grid' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}"
						title={m.grid_view()}
					>
						<LayoutGrid size={15} />
					</button>
				</div>

				<!-- New folder -->
				<button
					type="button"
					onclick={createDir}
					class="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50"
				>
					<FolderPlus size={15} /> {m.new_dir()}
				</button>

				<!-- Upload -->
				<button
					type="button"
					onclick={() => input?.click()}
					class="flex h-8 items-center gap-1.5 rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 active:bg-blue-800"
				>
					<Upload size={15} /> {m.upload_files()}
				</button>
				<input bind:this={input} type="file" multiple class="hidden" onchange={onPick} />
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
				<p class="text-sm text-gray-400">{currentDir ? m.dir_empty() : m.no_files()}</p>
			</div>
		{:else if viewMode === 'grid'}
			<!-- Grid view -->
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
				{#each files as f (f.id)}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						class="group relative flex flex-col items-center rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-200 hover:shadow-md {isDir(f) ? 'cursor-pointer' : ''}"
						onclick={isDir(f) ? () => goToDir(f) : undefined}
					>
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div class="absolute right-1.5 top-1.5 flex gap-0.5 rounded-lg bg-white/90 opacity-0 shadow-sm backdrop-blur transition-opacity group-hover:opacity-100" onclick={(e) => e.stopPropagation()}>
							{#if !isDir(f)}
								<button
									type="button"
									onclick={() => (previewFile = { id: f.id, name: f.name, mimeType: f.mime_type, size: f.size })}
									class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
								>
									<Eye size={14} />
								</button>
								<a
									href={getDownloadUrl(f.id)}
									download={f.name}
									class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
								>
									<Download size={14} />
								</a>
							{/if}
							<button
								type="button"
								onclick={() => rename(f.id, f.name)}
								class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
							>
								<Pencil size={14} />
							</button>
							<button
								type="button"
								onclick={() => remove(f.id, f.name)}
								disabled={deleting}
								class="rounded-md p-1 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 disabled:opacity-30"
							>
								<Trash2 size={14} />
							</button>
						</div>
						<MimeIcon mimeType={f.mime_type} isDir={isDir(f)} size={36} />
						<p class="mt-3 w-full truncate text-center text-sm font-medium text-gray-700" title={f.name}>{f.name}</p>
						<p class="mt-0.5 text-xs text-gray-400">{isDir(f) ? '' : fmtSize(f.size)}</p>
					</div>
				{/each}
			</div>
		{:else}
			<!-- List view -->
			<div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
				<table class="w-full table-fixed text-sm">
					<thead>
						<tr class="border-b border-gray-100 text-left text-xs text-gray-400">
							<th class="w-[45%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
							<th class="w-[15%] px-4 py-2.5 font-medium">{m.col_type()}</th>
							<th class="w-[10%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_upload_time()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody>
						{#each files as f (f.id)}
							<tr class="border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80 {isDir(f) ? 'cursor-pointer' : ''}"
								onclick={isDir(f) ? () => goToDir(f) : undefined}
							>
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-2.5">
										<span class="shrink-0"><MimeIcon mimeType={f.mime_type} isDir={isDir(f)} size={18} /></span>
										<span class="truncate text-gray-700" title={f.name}>{f.name}</span>
									</div>
								</td>
								<td class="truncate px-4 py-2.5 text-xs text-gray-400" title={isDir(f) ? m.directory() : f.mime_type}>{isDir(f) ? m.directory() : f.mime_type}</td>
								<td class="px-4 py-2.5 text-right text-gray-500">{isDir(f) ? '-' : fmtSize(f.size)}</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400">
									{new Date(f.created_at * 1000).toLocaleString()}
								</td>
								<!-- svelte-ignore a11y_no_static_element_interactions -->
								<td class="px-4 py-2.5 text-right" onclick={(e) => e.stopPropagation()}>
									<div class="flex items-center justify-end">
										{#if !isDir(f)}
											<button
												type="button"
												onclick={() => (previewFile = { id: f.id, name: f.name, mimeType: f.mime_type, size: f.size })}
												class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
											>
												<Eye size={15} />
											</button>
											<a
												href={getDownloadUrl(f.id)}
												download={f.name}
												class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
											>
												<Download size={15} />
											</a>
										{/if}
										<button
											type="button"
											onclick={() => rename(f.id, f.name)}
											class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
										>
											<Pencil size={15} />
										</button>
										<button
											type="button"
											onclick={() => remove(f.id, f.name)}
											disabled={deleting}
											class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 disabled:opacity-30"
										>
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
	<p class="text-gray-600">{@html m.please_login({ link: '<a href="/login" class="text-blue-600 underline hover:text-blue-700">' + m.login_link_text() + '</a>' })}</p>
{/if}

{#if previewFile}
	<DrivePreview
		id={previewFile.id}
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
			<!-- Expanded panel -->
			<div class="rounded-xl border border-gray-200 bg-white shadow-lg">
				<!-- Header -->
				<div class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5">
					<div class="flex items-center gap-2">
						<span class="text-sm font-medium text-gray-800">
							{hasActiveUploads ? m.uploading() : m.upload_done()}
						</span>
						<span class="text-xs text-gray-400">{completedCount}/{uploadItems.length}</span>
						{#if totalSpeed > 0}
							<span class="text-xs text-blue-500">{fmtSpeed(totalSpeed)}</span>
						{/if}
					</div>
					<div class="flex items-center gap-1">
						<button
							type="button"
							onclick={clearCompleted}
							class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600"
							title={m.clear()}
						>
							<X size={14} />
						</button>
						<button
							type="button"
							onclick={() => (uploadPanelOpen = false)}
							class="rounded-md p-1 text-gray-400 transition-colors hover:text-gray-600"
						>
							<ChevronDown size={14} />
						</button>
					</div>
				</div>
				<!-- Item list -->
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
									{#if item.phase === 'pending'}
										<Loader2 size={14} class="animate-spin text-gray-300" />
									{:else if item.phase === 'uploading'}
										<span class="mr-1 text-xs font-medium text-blue-600">{item.progress}%</span>
										<button
											type="button"
											onclick={() => pauseUpload(item.uid)}
											class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-amber-500"
											title={m.upload_paused()}
										>
											<Pause size={13} />
										</button>
									{:else if item.phase === 'paused'}
										<span class="mr-1 text-xs font-medium text-amber-500">{m.upload_paused()}</span>
										<button
											type="button"
											onclick={() => resumeUpload(item.uid)}
											class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500"
											title={m.upload_resume()}
										>
											<Play size={13} />
										</button>
									{:else if item.phase === 'completed'}
										<CheckCircle2 size={14} class="text-green-500" />
									{:else if item.phase === 'failed'}
										<XCircle size={14} class="text-red-500" />
										<button
											type="button"
											onclick={() => resumeUpload(item.uid)}
											class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-blue-500"
											title={m.upload_retry()}
										>
											<Play size={13} />
										</button>
									{/if}
									{#if item.phase !== 'completed'}
										<button
											type="button"
											onclick={() => deleteUpload(item.uid)}
											class="rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-red-500"
											title={m.remove()}
										>
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
							{:else if item.phase === 'pending'}
								<p class="mt-1 text-xs text-gray-300">{m.waiting()}</p>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<!-- Collapsed mini bar -->
			<button
				type="button"
				onclick={() => (uploadPanelOpen = true)}
				class="flex w-full items-center gap-3 rounded-xl border border-gray-200 bg-white px-4 py-2.5 shadow-lg transition-colors hover:bg-gray-50"
			>
				{#if hasActiveUploads}
					<Loader2 size={16} class="animate-spin text-blue-500" />
				{:else}
					<CheckCircle2 size={16} class="text-green-500" />
				{/if}
				<span class="flex-1 text-sm text-gray-700">{completedCount}/{uploadItems.length}</span>
				{#if totalSpeed > 0}
					<span class="text-xs text-blue-500">{fmtSpeed(totalSpeed)}</span>
				{/if}
				{#if failedCount > 0}
					<span class="text-xs text-red-500">{failedCount} failed</span>
				{/if}
				<ChevronUp size={14} class="text-gray-400" />
			</button>
		{/if}
	</div>
{/if}
