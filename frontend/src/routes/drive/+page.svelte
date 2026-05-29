<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { Upload, Loader2, FolderPlus, LayoutGrid, LayoutList, Search } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import FileListView from '$lib/components/files/FileListView.svelte';
	import UploadPanel from '$lib/components/files/UploadPanel.svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import type { NormalizedFile } from '$lib/types/file';
	import type { UploadItem } from '$lib/types/upload';
	import { normalizeDriveFile } from '$lib/types/adapters';
	import { confirmDelete, promptInput, confirmAction } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize } from '$lib/utils/format';
	import {
		listDrive,
		deleteDriveFile,
		renameDriveFile,
		getDownloadUrl,
		getDriveAncestors,
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

	let files = $state<DriveFile[]>([]);
	let normalizedFiles = $derived(files.map(normalizeDriveFile));
	let total = $state(0);
	let loading = $state(true);
	let loadingMore = $state(false);
	let input: HTMLInputElement | undefined = $state();
	let previewFile = $state<{ id: string; name: string; mimeType: string; size: number } | null>(null);

	let uploadItems = $state<UploadItem[]>([]);
	let activeCount = $state(0);

	let uidCounter = 0;
	function nextUid() {
		return `drive-upload-${++uidCounter}`;
	}

	function log(uid: string, msg: string, data?: unknown) {
		const prefix = `[drive-upload:${uid}]`;
		if (data !== undefined) console.log(prefix, msg, data);
		else console.log(prefix, msg);
	}

	let searchQuery = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let refreshId = 0;
	let hashQueue = Promise.resolve();

	type ViewMode = 'list' | 'grid';
	let viewMode = $state<ViewMode>(
		browser ? (localStorage.getItem('nd.drive.view') as ViewMode) || 'list' : 'list'
	);
	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		if (browser) localStorage.setItem('nd.drive.view', mode);
	}

	let currentDir = $state<string | null>(null);
	type DirBreadcrumb = { id: string | null; name: string };
	let dirStack = $state<DirBreadcrumb[]>([{ id: null, name: m.all_files() }]);
	let crumbs = $derived(dirStack.filter((d) => d.id !== null).map((d) => ({ id: d.id!, name: d.name })));
	let breadcrumbRef: Breadcrumb | undefined = $state();

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
		loading = true;
		files = [];
		void goto(buildDirUrl(id), { keepFocus: true, noScroll: true });
	}

	afterNavigate(async () => {
		breadcrumbRef?.collapse();
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
		const trimmed = name.trim();
		if (files.some((f) => f.isDir && f.name === trimmed)) {
			toast.error(m.dir_already_exists());
			return;
		}
		try {
			await createDriveDir(trimmed, currentDir ?? undefined);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.create_dir_failed());
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
		loadingMore = false;
		try {
			const data = await listDrive(PAGE_SIZE, 0, query ?? (searchQuery || undefined), currentDir ?? undefined);
			if (id !== refreshId) return;
			files = data.items;
			total = data.total;
		} catch (e) {
			if (id !== refreshId) return;
			toast.error(e instanceof Error ? e.message : m.load_failed());
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
			toast.error(e instanceof Error ? e.message : m.load_more_failed());
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
			fileHash: '',
			preHash: '',
			phase: 'pending',
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
				log(item.uid, `picked from queue: ${file.name} (${fmtSize(file.size)})`);
				item.phase = 'uploading';
				item.abortCtrl = new AbortController();
				activeCount++;

				try {
					const mime = file.type || 'application/octet-stream';
					const dirId = currentDir;

					// Compute SHA-256 and do unified pre-upload check
					let checkResult: Awaited<ReturnType<typeof driveCheckUpload>> | null = null;
					let fileHash = '';
					log(item.uid, `computing hash (size=${file.size}, limit=200MB)...`);
					hashQueue = hashQueue.then(async () => {
						if (file.size <= 200 * 1024 * 1024) {
							try {
								fileHash = await computeSHA256(file);
								log(item.uid, `hash done: ${fileHash}`);
								checkResult = await driveCheckUpload(fileHash, file.size, file.name, mime, dirId);
								log(item.uid, `checkUpload result: ${(checkResult as any)?.status}`);
							} catch (e) {
								// Check failed — fall through to normal upload
							}
						}
					});
					await hashQueue;

					if ((item.phase as string) === 'paused') continue;

					log(item.uid, `check result status: ${(checkResult as any)?.status ?? 'null (hash skipped or failed)'}`);

					if ((checkResult as any)?.status === 'full') {
						log(item.uid, 'scenario 1: file already exists, claiming hash...');
						// If the current user already has this file, ask for confirmation
						if ((checkResult as any).ownFile) {
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
						log(item.uid, 'completed (dedup claim)');
					} else {
						// Scenario 2 (partial) or 3 (none) — init session with hash
						// Reuse hash if already computed, otherwise compute now
						let sha256ForInit = fileHash;
						if (!sha256ForInit && file.size <= 200 * 1024 * 1024) {
							try { sha256ForInit = await computeSHA256(file); } catch {}
						}
						log(item.uid, `scenario 2/3: init upload session (hash=${sha256ForInit ? sha256ForInit.slice(0, 12) + '...' : 'none'})...`);
						const sessionResp = await initDriveUpload(file.name, mime, file.size, dirId, sha256ForInit || undefined);

						// Handle case where Init returns already_uploaded (race condition)
						if (sessionResp.status === 'already_uploaded' && sessionResp.fileId) {
							log(item.uid, 'init returned already_uploaded');
							item.phase = 'completed';
							item.progress = 100;
							item.uploadedBytes = file.size;
							item.speed = 0;
							continue;
						}

						const session = sessionResp;
						item.sessionId = session.id;
						log(item.uid, `session created: ${session.id}, receivedBytes=${session.receivedBytes || 0}`);

						// For partial sessions, start progress from existing offset
						const startOffset = session.receivedBytes || 0;
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
						log(item.uid, 'completed (chunked upload)');
					}
				} catch (e) {
					if ((item.phase as string) === 'paused') {
						log(item.uid, 'paused during operation, skipping');
						continue;
					}
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
					log(item.uid, `FAILED: ${item.errorMsg}`, e);
				} finally {
					item.abortCtrl = null;
					activeCount--;
				}
			}
		} finally {
			uploadQueueRunning = false;
		}

		if (uploadItems.some((i) => i.phase === 'completed')) {
			await refresh(searchQuery || undefined);
		}
	}

	function pauseUpload(uid: string) {
		const item = uploadItems.find((i) => i.uid === uid);
		if (!item || item.phase !== 'uploading') return;
		log(uid, 'paused by user');
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

		log(uid, `resumed by user (from ${item.phase}, session=${item.sessionId ?? 'none'})`);
		item.errorMsg = null;

		if (item.sessionId) {
			// Resume from existing session
			item.phase = 'uploading';
			log(uid, `resuming session ${item.sessionId} from byte ${item.uploadedBytes}`);
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
					log(uid, 'completed (resume)');
				}
			} catch (e) {
				if ((item.phase as string) !== 'paused') {
					item.phase = 'failed';
					item.errorMsg = e instanceof Error ? e.message : m.upload_failed();
					log(uid, `resume FAILED: ${item.errorMsg}`, e);
				}
			} finally {
				item.abortCtrl = null;
				activeCount--;
			}
		} else {
			// No session yet — restart from beginning
			log(uid, 'no session, restarting from beginning');
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
		log(uid, `deleted by user (phase=${item.phase}, session=${item.sessionId ?? 'none'})`);

		if (item.phase === 'uploading') {
			item.abortCtrl?.abort();
		}
		if (item.sessionId) {
			log(uid, `cancelling session ${item.sessionId}`);
			try { await cancelDriveUpload(item.sessionId); } catch { /* best effort */ }
		}

		uploadItems = uploadItems.filter((i) => i.uid !== uid);
	}

	function goToDir(f: DriveFile) {
		navigateToDir(f.id, f.name);
	}

	function onPreviewFile(file: NormalizedFile) {
		previewFile = { id: file.id, name: file.name, mimeType: file.mimeType || '', size: file.size };
	}

	async function onRenameFile(id: string, currentName: string) {
		await rename(id, currentName);
	}

	async function onDeleteFile(id: string, name: string) {
		await remove(id, name);
	}

	function clearCompleted() {
		uploadItems = [];
	}

	async function remove(id: string, name: string) {
		if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
		try {
			await deleteDriveFile(id);
			await refresh(searchQuery || undefined);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		}
	}

	async function rename(id: string, currentName: string) {
		const newName = await promptInput(m.rename(), m.enter_new_name(), currentName);
		if (!newName || newName === currentName) return;
		try {
			await renameDriveFile(id, newName);
			await refresh(searchQuery || undefined);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.rename_failed());
		}
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
			<Breadcrumb
				bind:this={breadcrumbRef}
				items={crumbs}
				onNavigate={(id) => { searchQuery = ''; void goto(buildDirUrl(id), { keepFocus: true, noScroll: true }); }}
				onHome={() => { searchQuery = ''; void goto(buildDirUrl(null), { keepFocus: true, noScroll: true }); }}
			/>
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

		<!-- File list -->
		<FileListView
			files={normalizedFiles}
			{viewMode}
			{loading}
			downloadUrlFn={getDownloadUrl}
			emptyMessage={currentDir ? m.dir_empty() : m.no_files()}
			onNavigateDir={(id) => { const f = files.find(x => x.id === id); if (f) goToDir(f); }}
			onPreview={onPreviewFile}
			onRename={onRenameFile}
			onDelete={onDeleteFile}
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

<UploadPanel
	items={uploadItems}
	onPause={pauseUpload}
	onResume={resumeUpload}
	onDelete={deleteUpload}
	onClear={clearCompleted}
/>
