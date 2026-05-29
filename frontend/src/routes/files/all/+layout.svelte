<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import {
		listFiles, mkdir, trashFile, renameFile, setStarred,
		downloadUrl, getBreadcrumb, type FileItem
	} from '$lib/api/files';
	import type { NormalizedFile } from '$lib/types/file';
	import { normalizeFileItem } from '$lib/types/adapters';
	import {
		preCheck, requestChallenge, verify as verifyUpload,
		initUpload, uploadChunk, completeUpload, getUploadStatus, updateHash
	} from '$lib/api/upload';
	import { importFile } from '$lib/api/files';
	import { addToLibrary } from '$lib/api/media';
	import {
		Upload, FolderPlus, FolderOpen, Plus,
		LayoutGrid, LayoutList, Search,
		ArrowUpDown, ArrowUp, ArrowDown, Check
	} from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { toast } from 'svelte-sonner';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import FileListView from '$lib/components/files/FileListView.svelte';
	import FolderUploadDialog from '$lib/components/files/FolderUploadDialog.svelte';
	import UploadPanel from '$lib/components/files/UploadPanel.svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import type { UploadItem } from '$lib/types/upload';
	import { computeSHA256Chunked } from '$lib/upload-hash';
	import { confirmDelete, promptInput } from '$lib/dialog';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let { children } = $props();

	const PAGE_SIZE = 50;
	const CHUNK_SIZE = 4 * 1024 * 1024;

	let files = $state<FileItem[]>([]);
	let normalizedFiles = $derived(files.map(normalizeFileItem));
	let total = $state(0);
	let loading = $state(true);
	let loadingMore = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();
	let folderInput: HTMLInputElement | undefined = $state();
	let previewFile = $state<{ slug: string; name: string; mimeType: string; size: number } | null>(null);
	let folderDialogFiles = $state<{ file: File; relativePath: string }[]>([]);
	let folderDialogOpen = $state(false);
	let folderDialogLoading = $state(false);

	let uploadItems = $state<UploadItem[]>([]);

	let uidCounter = 0;
	function nextUid() {
		return `upload-${++uidCounter}`;
	}

	function log(uid: string, msg: string, data?: unknown) {
		const prefix = `[upload:${uid}]`;
		if (data !== undefined) console.log(prefix, msg, data);
		else console.log(prefix, msg);
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

	type SortField = 'file_name' | 'file_size' | 'created_at' | 'updated_at';
	let sortBy = $state<SortField>(
		browser ? (localStorage.getItem('nd.files.sortBy') as SortField) || 'created_at' : 'created_at'
	);
	let sortDir = $state<'ASC' | 'DESC'>(
		browser ? (localStorage.getItem('nd.files.sortDir') as 'ASC' | 'DESC') || 'DESC' : 'DESC'
	);
	function setSort(field: SortField) {
		if (sortBy === field) {
			sortDir = sortDir === 'ASC' ? 'DESC' : 'ASC';
		} else {
			sortBy = field;
			sortDir = field === 'file_name' ? 'ASC' : 'DESC';
		}
		if (browser) {
			localStorage.setItem('nd.files.sortBy', sortBy);
			localStorage.setItem('nd.files.sortDir', sortDir);
		}
		void refresh(true);
	}

	const sortOptions: { field: SortField; label: () => string }[] = [
		{ field: 'file_name', label: () => m.sort_name() },
		{ field: 'file_size', label: () => m.sort_size() },
		{ field: 'created_at', label: () => m.sort_created() },
		{ field: 'updated_at', label: () => m.sort_updated() },
	];

	let currentSlug = $state('');
	let crumbs = $state<{ id: string; name: string }[]>([]);
	let breadcrumbRef: Breadcrumb | undefined = $state();

	function navigateToDir(slug: string) {
		searchQuery = '';
		loading = true;
		files = [];
		void goto('/files/all/' + slug, { keepFocus: true, noScroll: true });
	}

	async function fetchBreadcrumb(dirSlug: string) {
		if (!dirSlug) {
			crumbs = [];
			return;
		}
		try {
			const items = await getBreadcrumb(dirSlug);
			crumbs = items.map((b) => ({ id: b.slug, name: b.fileName }));
		} catch {
			crumbs = [{ id: dirSlug, name: dirSlug }];
		}
	}

	afterNavigate(() => {
		const slug = page.params.slug ?? '';
		if (slug !== currentSlug) {
			currentSlug = slug;
			breadcrumbRef?.collapse();
		}
		void fetchBreadcrumb(currentSlug);
		void refresh(true);
	});

	async function createDir() {
		const name = await promptInput(m.new_folder(), m.enter_folder_name(), undefined, 100);
		if (!name) return;
		const trimmed = name.trim();
		if (files.some((f) => f.isDir && f.fileName === trimmed)) {
			toast.error(m.dir_already_exists());
			return;
		}
		try {
			await mkdir(trimmed, currentSlug || undefined);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.create_dir_failed());
		}
	}

	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void refresh(true), 300);
	}

	async function refresh(showLoading = false) {
		if (!$user) return;
		const id = ++refreshId;
		if (showLoading) loading = true;
		loadingMore = false;
		try {
			const data = await listFiles(currentSlug || undefined, 1, PAGE_SIZE, undefined, undefined, sortBy, sortDir);
			if (id !== refreshId) return;
			files = data.files;
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
		if (loadingMore) return;
		loadingMore = true;
		const id = ++refreshId;
		try {
			const page_num = Math.floor(files.length / PAGE_SIZE) + 1;
			const data = await listFiles(currentSlug || undefined, page_num, PAGE_SIZE, undefined, undefined, sortBy, sortDir);
			if (id !== refreshId) return;
			files = [...files, ...data.files];
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

		// Show dialog immediately with loading state
		folderDialogFiles = [];
		folderDialogOpen = true;
		folderDialogLoading = true;

		// Defer file processing to allow dialog to render first
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
									item.phase = 'importing';
									item.progress = 100;
									await importFile(verifyResult.physicalFileSlug, item.fileName, currentSlug || undefined);
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
					const task = await initUpload(item.fileHash, item.preHash, item.fileSize, mimeType);
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
						await importFile(completedTask.physicalFileSlug, item.fileName, currentSlug || undefined);
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
			await refresh();
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

	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
		try {
			await trashFile(slug);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		}
	}

	async function rename(slug: string, currentName: string) {
		const newName = await promptInput(m.rename(), m.enter_new_name(), currentName, 100);
		if (!newName || newName === currentName) return;
		try {
			await renameFile(slug, newName);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.rename_failed());
		}
	}

	async function toggleStar(slug: string, currentlyStarred: boolean) {
		try {
			await setStarred(slug, !currentlyStarred);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.unstar_failed());
		}
	}

	function onPreview(file: NormalizedFile) {
		previewFile = { slug: file.id, name: file.name, mimeType: file.mimeType || '', size: file.size };
	}

	async function onAddToMedia(file: NormalizedFile) {
		try {
			const resp = await addToLibrary(file.id);
			if (resp.transcodeStatus === 'existing') {
				toast.info(m.media_already_in_library());
			} else {
				toast.success(m.media_add_success());
			}
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_add_failed());
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
			<Breadcrumb
				bind:this={breadcrumbRef}
				items={crumbs}
				onNavigate={(id) => { searchQuery = ''; void goto('/files/all/' + id, { keepFocus: true, noScroll: true }); }}
				onHome={() => { searchQuery = ''; void goto('/files/all', { keepFocus: true, noScroll: true }); }}
			/>
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
				<Dropdown
					triggerClass="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 text-sm text-gray-600 transition-colors hover:border-gray-300 hover:bg-gray-50"
					contentClass="min-w-[144px]"
				>
					{#snippet trigger()}
						<ArrowUpDown size={14} />
						<span class="hidden sm:inline">{sortOptions.find(o => o.field === sortBy)?.label()}</span>
						{#if sortDir === 'ASC'}
							<ArrowUp size={14} class="text-blue-500" />
						{:else}
							<ArrowDown size={14} class="text-blue-500" />
						{/if}
					{/snippet}

					{#each sortOptions as opt (opt.field)}
						<DropdownBase.Item onSelect={() => setSort(opt.field)}>
							<span class={sortBy === opt.field ? 'font-medium text-gray-900' : ''}>{opt.label()}</span>
							{#if sortBy === opt.field}
								{#if sortDir === 'ASC'}
									<ArrowUp size={14} class="ml-auto text-blue-500" />
								{:else}
									<ArrowDown size={14} class="ml-auto text-blue-500" />
								{/if}
							{/if}
						</DropdownBase.Item>
					{/each}
				</Dropdown>

				<div class="flex overflow-hidden rounded-lg border border-gray-200">
					<button type="button" onclick={() => setViewMode('list')} class="p-1.5 transition-colors {viewMode === 'list' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
						<LayoutList size={15} />
					</button>
					<button type="button" onclick={() => setViewMode('grid')} class="p-1.5 transition-colors {viewMode === 'grid' ? 'bg-blue-50 text-blue-600' : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}">
						<LayoutGrid size={15} />
					</button>
				</div>

				<Dropdown
					triggerClass="flex h-8 items-center gap-1.5 rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 active:bg-blue-800"
					contentClass="min-w-[180px]"
				>
					{#snippet trigger()}
						<Plus size={15} /> {m.new_folder()}
					{/snippet}

					<DropdownBase.Item
						class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50"
						onSelect={() => fileInput?.click()}
					>
						<Upload size={15} class="text-gray-500" />
						{m.upload_files()}
					</DropdownBase.Item>

					<DropdownBase.Item
						class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50"
						onSelect={() => folderInput?.click()}
					>
						<FolderOpen size={15} class="text-gray-500" />
						{m.upload_folder()}
					</DropdownBase.Item>

					<DropdownBase.Separator />

					<DropdownBase.Item
						class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50"
						onSelect={() => createDir()}
					>
						<FolderPlus size={15} class="text-gray-500" />
						{m.new_folder()}
					</DropdownBase.Item>
				</Dropdown>
				<input bind:this={fileInput} type="file" multiple class="hidden" onchange={onPick} />
				<input bind:this={folderInput} type="file" webkitdirectory class="hidden" onchange={onPickFolder} />
			</div>
		</div>

		<!-- File list -->
		<FileListView
			files={normalizedFiles}
			{viewMode}
			{loading}
			directoryId={currentSlug}
			downloadUrlFn={downloadUrl}
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

<FolderUploadDialog
	files={folderDialogFiles}
	open={folderDialogOpen}
	loading={folderDialogLoading}
	onConfirm={onFolderConfirm}
	onCancel={() => { folderDialogOpen = false; }}
/>

<UploadPanel
	items={uploadItems}
	onPause={pauseUpload}
	onResume={resumeUpload}
	onDelete={deleteUpload}
	onClear={clearCompleted}
/>

{@render children()}
