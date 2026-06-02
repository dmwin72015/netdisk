<script lang="ts">
	import { Check, ChevronLeft, Folder, FolderOpen, LoaderCircle, MoveRight } from '@lucide/svelte';
	import { listFiles, type FileItem } from '$lib/api/files';
	import { Dialog } from '$lib/ui/dialog';
	import * as m from '$lib/paraglide/messages';

	type TargetDir = {
		slug: string;
		name: string;
	};

	let {
		open = $bindable(false),
		excludedIds = [],
		includeSystemDirs = true,
		onClose,
		onConfirm,
	}: {
		open?: boolean;
		excludedIds?: string[];
		includeSystemDirs?: boolean;
		onClose: () => void;
		onConfirm: (targetParentSlug: string) => void | Promise<void>;
	} = $props();

	const PAGE_SIZE = 100;
	let dirs = $state<FileItem[]>([]);
	let total = $state(0);
	let page = $state(1);
	let loading = $state(false);
	let loadingMore = $state(false);
	let submitting = $state(false);
	let error = $state<string | null>(null);
	let selectedSlug = $state('');
	let currentDir = $state<TargetDir | null>(null);
	let path = $state<TargetDir[]>([]);
	let prevOpen = $state(false);

	let selectableDirs = $derived(dirs.filter((dir) => !excludedIds.includes(dir.slug)));
	let canLoadMore = $derived(dirs.length < total);
	let targetName = $derived(
		selectedSlug
			? (currentDir?.slug === selectedSlug
					? currentDir.name
					: selectableDirs.find((dir) => dir.slug === selectedSlug)?.fileName) || m.all_files()
			: m.all_files()
	);

	$effect(() => {
		if (open && !prevOpen) {
			reset();
			void fetchDirs(null, 1, false);
		}
		prevOpen = open;
	});

	function reset() {
		dirs = [];
		total = 0;
		page = 1;
		loading = false;
		loadingMore = false;
		submitting = false;
		error = null;
		selectedSlug = '';
		currentDir = null;
		path = [];
	}

	async function fetchDirs(parent: TargetDir | null, nextPage = 1, append = false) {
		if (append) loadingMore = true;
		else loading = true;
		error = null;

		try {
			const data = await listFiles(
				parent?.slug,
				nextPage,
				PAGE_SIZE,
				undefined,
				undefined,
				'file_name',
				'ASC',
				true,
				includeSystemDirs
			);
			dirs = append ? [...dirs, ...data.files] : data.files;
			total = data.total;
			page = nextPage;
			currentDir = parent;
		} catch (e) {
			error = e instanceof Error ? e.message : m.load_failed();
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	function selectRoot() {
		selectedSlug = '';
	}

	async function enterDir(dir: FileItem) {
		const target = { slug: dir.slug, name: dir.fileName };
		path = [...path, target];
		selectedSlug = dir.slug;
		await fetchDirs(target, 1, false);
	}

	async function goBack() {
		if (path.length === 0) return;
		const nextPath = path.slice(0, -1);
		path = nextPath;
		const parent = nextPath.at(-1) ?? null;
		selectedSlug = parent?.slug ?? '';
		await fetchDirs(parent, 1, false);
	}

	async function loadMore() {
		if (!canLoadMore || loadingMore) return;
		await fetchDirs(currentDir, page + 1, true);
	}

	async function confirmMove() {
		submitting = true;
		error = null;
		try {
			await onConfirm(selectedSlug);
		} catch (e) {
			error = e instanceof Error ? e.message : m.move_failed();
		} finally {
			submitting = false;
		}
	}

	function handleOpenChange(v: boolean) {
		if (!v && !submitting) {
			open = false;
			onClose();
		}
	}
</script>

<Dialog
	bind:open
	onOpenChange={handleOpenChange}
	onCancel={onClose}
	title={m.select_move_target()}
	description={m.move_target_hint()}
	footer={false}
	closable={!submitting}
	class="h-[68vh] max-w-xl"
	bodyClass="!p-0 flex flex-col min-h-0 !overflow-hidden"
>
	<div class="flex items-center gap-2 border-b border-gray-100 px-5 py-2.5">
		<button
			type="button"
			onclick={goBack}
			disabled={path.length === 0 || loading || submitting}
			class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 disabled:opacity-40"
			aria-label={m.back()}
		>
			<ChevronLeft size={16} />
		</button>
		<div class="min-w-0 flex-1 truncate text-xs text-gray-500">
			<button type="button" class="text-gray-600 hover:text-blue-600" onclick={() => { path = []; selectedSlug = ''; void fetchDirs(null, 1, false); }}>
				{m.all_files()}
			</button>
			{#each path as item (item.slug)}
				<span class="px-1 text-gray-300">/</span>
				<span class="text-gray-600">{item.name}</span>
			{/each}
		</div>
	</div>

	<div class="border-b border-gray-100 px-4 py-2">
		<button
			type="button"
			onclick={selectRoot}
			class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors hover:bg-gray-50 {selectedSlug === '' ? 'bg-blue-50' : ''}"
		>
			<span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500">
				<FolderOpen size={18} />
			</span>
			<span class="min-w-0 flex-1 text-sm text-gray-700">{m.all_files()}</span>
			{#if selectedSlug === ''}
				<Check size={16} class="text-blue-600" />
			{/if}
		</button>
	</div>

	<div class="flex-1 overflow-y-auto p-4">
		{#if error}
			<div class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
		{/if}

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if selectableDirs.length === 0}
			<div class="flex flex-col items-center justify-center py-14 text-center">
				<Folder size={38} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.no_folders_found()}</p>
			</div>
		{:else}
			<div class="space-y-1">
				{#each selectableDirs as dir (dir.slug)}
					<div class="flex items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-gray-50 {selectedSlug === dir.slug ? 'bg-blue-50' : ''}">
						<button
							type="button"
							onclick={() => { selectedSlug = dir.slug; }}
							ondblclick={() => enterDir(dir)}
							class="flex min-w-0 flex-1 items-center gap-3 text-left"
						>
							<span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-500">
								<Folder size={18} />
							</span>
							<span class="min-w-0 flex-1 truncate text-sm text-gray-700">{dir.fileName}</span>
						</button>
						<button
							type="button"
							onclick={(e) => { e.stopPropagation(); void enterDir(dir); }}
							class="rounded-md px-2 py-1 text-xs text-gray-500 transition-colors hover:bg-white hover:text-blue-600"
						>
							{m.open_folder()}
						</button>
						{#if selectedSlug === dir.slug}
							<Check size={16} class="text-blue-600" />
						{/if}
					</div>
				{/each}
			</div>
			{#if canLoadMore}
				<div class="mt-3 flex justify-center">
					<button
						type="button"
						onclick={loadMore}
						disabled={loadingMore}
						class="rounded-lg px-3 py-1.5 text-sm text-gray-600 transition-colors hover:bg-gray-100 disabled:opacity-50"
					>
						{loadingMore ? m.loading() : m.load_more()}
					</button>
				</div>
			{/if}
		{/if}
	</div>

	<div class="flex items-center justify-between gap-3 border-t border-gray-100 px-5 py-3">
		<div class="min-w-0 truncate text-xs text-gray-400">
			{m.move_selected_target({ target: targetName })}
		</div>
		<div class="flex shrink-0 items-center gap-2">
			<button
				type="button"
				onclick={() => { open = false; onClose(); }}
				disabled={submitting}
				class="rounded-lg px-4 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100 disabled:opacity-50"
			>
				{m.cancel()}
			</button>
			<button
				type="button"
				onclick={confirmMove}
				disabled={submitting}
				class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 disabled:opacity-50"
			>
				{#if submitting}
					<LoaderCircle size={14} class="animate-spin" />
				{:else}
					<MoveRight size={14} />
				{/if}
				{m.move_here()}
			</button>
		</div>
	</div>
</Dialog>
