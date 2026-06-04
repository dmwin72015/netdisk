<script lang="ts">
	import { onDestroy, getContext } from 'svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import {
		listFiles, mkdir, trashFile, batchTrashFiles, renameFile, setStarred,
		downloadUrl, getBreadcrumb, moveFile, type FileItem
	} from '$lib/api/files';
	import { ApiError } from '$lib/api/client';
	import type { NormalizedFile } from '$lib/types/file';
	import { normalizeFileItem } from '$lib/types/adapters';
	import { addToLibrary } from '$lib/api/media';
	import { toast } from 'svelte-sonner';
	import DrivePreview from '$lib/components/DrivePreview.svelte';
	import FileListView from '$lib/components/files/FileListView.svelte';
	import FolderUploadDialog from '$lib/components/files/FolderUploadDialog.svelte';
	import Breadcrumb from '$lib/components/Breadcrumb.svelte';
	import FilesToolbar, { type SortField, type ViewMode } from '$lib/components/files/FilesToolbar.svelte';
	import { confirmDelete, promptInput } from '$lib/dialog';
	import { FileQuestionMark, LoaderCircle } from '@lucide/svelte';
	import { UPLOAD_FILE_CONCURRENCY } from '$lib/upload-concurrency';
	import type { createUploadManager as UploadMgrFn } from '$lib/upload-manager.svelte';
	type UploadManager = ReturnType<typeof UploadMgrFn>;
	import * as m from '$lib/paraglide/messages';

	let { children } = $props();

	const PAGE_SIZE = 50;

	// --- File listing ---
	let files = $state<FileItem[]>([]);
	let normalizedFiles = $derived(files.map(normalizeFileItem));
	let total = $state(0);
	let loading = $state(true);
	let loadingMore = $state(false);
	let deleting = $state(false);
	let notFound = $state(false);
	let refreshId = 0;

	// --- Preferences ---
	let searchQuery = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;

	let viewMode = $state<ViewMode>(
		browser ? (localStorage.getItem('nd.files.view') as ViewMode) || 'list' : 'list'
	);
	function setViewMode(mode: ViewMode) {
		viewMode = mode;
		if (browser) localStorage.setItem('nd.files.view', mode);
	}

	let sortBy = $state<SortField>(
		browser ? (localStorage.getItem('nd.files.sortBy') as SortField) || 'created_at' : 'created_at'
	);
	let sortDir = $state<'ASC' | 'DESC'>(
		browser ? (localStorage.getItem('nd.files.sortDir') as 'ASC' | 'DESC') || 'DESC' : 'DESC'
	);
	let showSystemDirs = $state(
		browser ? localStorage.getItem('nd.files.showSystemDirs') !== 'false' : true
	);

	function setShowSystemDirs(value: boolean) {
		showSystemDirs = value;
		if (browser) localStorage.setItem('nd.files.showSystemDirs', String(value));
		void refresh(true);
	}

	function normalizeUploadConcurrency(value: number) {
		return Math.max(1, Math.min(UPLOAD_FILE_CONCURRENCY, value || UPLOAD_FILE_CONCURRENCY));
	}

		const initialUploadConcurrency =
			browser ? normalizeUploadConcurrency(parseInt(localStorage.getItem('nd.files.uploadConcurrency') || String(UPLOAD_FILE_CONCURRENCY), 10)) : UPLOAD_FILE_CONCURRENCY
		;
		let uploadConcurrency = $state(initialUploadConcurrency);

	// --- Upload manager (shared via context) ---
	const upload = getContext<UploadManager>('upload');

	$effect(() => {
		upload.updateMaxConcurrent(uploadConcurrency);
		upload.setGetCurrentSlug(() => currentSlug);
		upload.setOnCompleted(() => refresh());
	});

	let fileInput: HTMLInputElement | undefined = $state();
	let folderInput: HTMLInputElement | undefined = $state();

	function setUploadConcurrency(value: number) {
		uploadConcurrency = normalizeUploadConcurrency(value);
		if (browser) localStorage.setItem('nd.files.uploadConcurrency', String(uploadConcurrency));
	}

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

	// --- Breadcrumb / Navigation ---
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

	// --- Search ---
	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => void refresh(true), 300);
	}

	// --- File listing ---
	async function refresh(showLoading = false) {
		if (!$user) return;
		const id = ++refreshId;
		if (showLoading) loading = true;
		loadingMore = false;
		notFound = false;
		try {
			const data = await listFiles(
				currentSlug || undefined,
				1,
				PAGE_SIZE,
				undefined,
				undefined,
				sortBy,
				sortDir,
				false,
				showSystemDirs
			);
			if (id !== refreshId) return;
			files = data.files;
			total = data.total;
		} catch (e) {
			if (id !== refreshId) return;
			if (e instanceof ApiError && e.status === 404) {
				notFound = true;
			} else {
				toast.error(e instanceof Error ? e.message : m.load_failed());
			}
		} finally {
			if (id === refreshId) loading = false;
		}
	}

	async function loadMore() {
		if (loadingMore) return;
		loadingMore = true;
		const id = ++refreshId;
		try {
			const page_num = Math.floor(files.length / PAGE_SIZE) + 1;
			const data = await listFiles(
				currentSlug || undefined,
				page_num,
				PAGE_SIZE,
				undefined,
				undefined,
				sortBy,
				sortDir,
				false,
				showSystemDirs
			);
			if (id !== refreshId) return;
			files = [...files, ...data.files];
		} catch (e) {
			if (id !== refreshId) return;
			toast.error(e instanceof Error ? e.message : m.load_more_failed());
		} finally {
			if (id === refreshId) loadingMore = false;
		}
	}

	async function loadFolderSummary(slug: string) {
		const pageSize = 100;
		let pageNum = 1;
		let loaded = 0;
		let total = 0;
		const summary = { fileCount: 0, folderCount: 0, size: 0 };

		do {
			const data = await listFiles(slug, pageNum, pageSize);
			total = data.total;
			loaded += data.files.length;
			for (const file of data.files) {
				if (file.isDir) summary.folderCount += 1;
				else {
					summary.fileCount += 1;
					summary.size += file.fileSize;
				}
			}
			pageNum += 1;
		} while (loaded < total);

		return summary;
	}

	onDestroy(() => {
		clearTimeout(searchTimer);
	});

	// --- Create directory ---
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

	// --- File operations ---
	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
		deleting = true;
		try {
			await trashFile(slug);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		} finally {
			deleting = false;
		}
	}

	async function batchRemove(ids: string[]) {
		const files = ids.map(id => normalizedFiles.find(f => f.id === id)).filter(Boolean) as NormalizedFile[];
		if (files.length === 0) return;
		const names = files.map(f => f.name);
		if (!(await confirmDelete(m.confirm_delete_multiple({ count: String(files.length), names: names.join('\n') })))) return;
		deleting = true;
		try {
			await batchTrashFiles(ids);
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.delete_failed());
		} finally {
			deleting = false;
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

	async function move(ids: string[], targetParentSlug: string) {
		if (ids.length === 0) return;
		try {
			await Promise.all(ids.map((id) => moveFile(id, targetParentSlug)));
			toast.success(m.move_success({ count: ids.length }));
			await refresh();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.move_failed());
			throw e;
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

	// --- Preview ---
	let previewFile = $state<{ slug: string; name: string; mimeType: string; size: number } | null>(null);

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


</script>

{#if $authReady && $user}
	<div class="space-y-4">
		{#if notFound}
			<div class="flex flex-col items-center justify-center py-24 text-center">
				<FileQuestionMark size={48} class="mb-4 text-gray-300" />
				<p class="mb-4 text-lg text-gray-500">{m.file_not_found()}</p>
				<button
					type="button"
					onclick={() => { notFound = false; void goto('/files/all', { keepFocus: true }); }}
					class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
				>
					{m.back_to_root()}
				</button>
			</div>
		{:else if loading && files.length === 0}
			<div class="flex items-center justify-center py-24">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else}
			<!-- Breadcrumb -->
			{#if currentSlug}
			<Breadcrumb
				bind:this={breadcrumbRef}
				items={crumbs}
				onNavigate={(id) => { searchQuery = ''; void goto('/files/all/' + id, { keepFocus: true, noScroll: true }); }}
				onHome={() => { searchQuery = ''; void goto('/files/all', { keepFocus: true, noScroll: true }); }}
			/>
		{/if}

		<!-- Toolbar -->
		<FilesToolbar
			bind:searchQuery
			{sortBy}
			{sortDir}
			{viewMode}
			onSearch={onSearchInput}
			onSort={setSort}
			onViewModeChange={setViewMode}
			onUploadFiles={() => fileInput?.click()}
			onUploadFolder={() => folderInput?.click()}
			onCreateDir={createDir}
			{showSystemDirs}
			onShowSystemDirsChange={setShowSystemDirs}
			{uploadConcurrency}
			onUploadConcurrencyChange={setUploadConcurrency}
		/>
		<input bind:this={fileInput} type="file" multiple class="hidden" onchange={upload.onPick} />
		<input bind:this={folderInput} type="file" webkitdirectory class="hidden" onchange={upload.onPickFolder} />

		<!-- File list -->
		<div class="relative">
			{#if deleting}
				<div class="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-white/60 backdrop-blur-[1px]">
					<LoaderCircle size={24} class="animate-spin text-gray-400" />
				</div>
			{/if}
			<FileListView
				files={normalizedFiles}
				{viewMode}
				{loading}
				directoryId={currentSlug}
				currentPath={crumbs}
				includeSystemDirs={showSystemDirs}
				downloadUrlFn={downloadUrl}
				emptyMessage={currentSlug ? m.dir_empty() : m.no_files()}
				onNavigateDir={navigateToDir}
				onStar={toggleStar}
				onPreview={onPreview}
				onRename={rename}
				onDelete={remove}
				onBatchDelete={batchRemove}
				onMove={move}
				onAddToMedia={onAddToMedia}
				{loadFolderSummary}
			/>
		</div>

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
		{/if}
	</div>
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
	files={upload.folderDialogFiles}
	open={upload.folderDialogOpen}
	loading={upload.folderDialogLoading}
	onConfirm={upload.onFolderConfirm}
	onCancel={() => { upload.folderDialogOpen = false; }}
/>

{@render children()}
