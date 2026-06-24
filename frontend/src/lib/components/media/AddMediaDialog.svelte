<script lang="ts">
	import { LoaderCircle, Film, Check, Search, Folder } from '@lucide/svelte';
	import { listFiles, type FileItem } from '$lib/api/files';
	import { addToLibrary } from '$lib/api/media';
	import { fmtSize } from '$lib/utils/format';
	import { Dialog } from '$lib/ui/dialog';
	import * as m from '$lib/paraglide/messages';

	let {
		open = $bindable(false),
		onClose,
		onDone,
		onNavigateDir
	}: {
		open?: boolean;
		onClose: () => void;
		onDone: () => void;
		onNavigateDir?: (slug: string) => void;
	} = $props();

	let videos = $state<FileItem[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let page = $state(1);
	let selected = $state<Record<string, boolean>>({});
	let submitting = $state(false);
	let error = $state<string | null>(null);
	let searchQuery = $state('');
	let prevOpen = $state(false);
	const PAGE_SIZE = 20;

	function selectedCount() {
		return Object.values(selected).filter(Boolean).length;
	}

	async function fetchVideos(p: number) {
		loading = true;
		error = null;
		try {
			const data = await listFiles(undefined, p, PAGE_SIZE, undefined, 'video');
			videos = data.files;
			total = data.total;
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_load_failed();
		} finally {
			loading = false;
		}
	}

	let filteredVideos = $derived(
		searchQuery.trim()
			? videos.filter(v => v.fileName.toLowerCase().includes(searchQuery.trim().toLowerCase()))
			: videos
	);

	$effect(() => {
		if (open && !prevOpen) {
			selected = {};
			page = 1;
			searchQuery = '';
			fetchVideos(1);
		}
		prevOpen = open;
	});

	function toggle(slug: string) {
		selected = { ...selected, [slug]: !selected[slug] };
	}

	function toggleAll() {
		const allSelected = filteredVideos.every(v => selected[v.slug]);
		const next: Record<string, boolean> = {};
		if (!allSelected) {
			for (const v of filteredVideos) next[v.slug] = true;
		}
		selected = next;
	}

	async function submit() {
		if (selectedCount() === 0) return;
		submitting = true;
		error = null;
		try {
			for (const [slug, isSelected] of Object.entries(selected)) {
				if (isSelected) await addToLibrary(slug);
			}
			onDone();
			open = false;
			onClose();
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_add_failed();
		} finally {
			submitting = false;
		}
	}

	function handleClose(v: boolean) {
		if (!v) onClose();
	}
</script>

<Dialog
	bind:open
	onOpenChangeComplete={handleClose}
	title={m.select_videos()}
	footer={false}
	class="h-[70vh]"
	contentStyle="max-width: 56rem"
	bodyClass="p-0 flex flex-col min-h-0 overflow-hidden"
>
		<!-- Search -->
		<div class="border-b border-line-soft px-5 py-2.5">
			<div class="relative">
				<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-4" />
				<input
					type="text"
					bind:value={searchQuery}
					placeholder={m.search_files()}
					class="w-full rounded-lg border border-line bg-surface-muted py-2 pl-9 pr-3 text-sm text-ink-2 outline-none transition-colors placeholder:text-ink-4 focus:border-primary focus:bg-white"
				/>
			</div>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-y-auto p-4">
			{#if error}
				<div class="mb-3 rounded-lg border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">{error}</div>
			{/if}

			{#if loading}
				<div class="flex items-center justify-center py-16">
					<LoaderCircle size={24} class="animate-spin text-ink-4" />
				</div>
			{:else if filteredVideos.length === 0}
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<Film size={40} class="mb-3 text-ink-4" />
					<p class="text-sm text-ink-4">{searchQuery.trim() ? m.no_videos_found() : m.no_videos_found()}</p>
				</div>
			{:else}
				<!-- Select all -->
				<button type="button" onclick={toggleAll} class="mb-2 text-xs text-primary transition-colors hover:text-primary">
					{filteredVideos.every(v => selected[v.slug]) ? m.cancel() : m.add_selected({ count: filteredVideos.length })}
				</button>

				<div class="space-y-1">
					{#each filteredVideos as v (v.slug)}
						{@const isSelected = !!selected[v.slug]}
						<button type="button" onclick={() => toggle(v.slug)} class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors hover:bg-surface-muted {isSelected ? 'bg-primary-soft' : ''}">
							<div class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {isSelected ? 'border-primary bg-primary text-white' : 'border-line'}">
								{#if isSelected}
									<Check size={12} />
								{/if}
							</div>
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm text-ink-2">{v.fileName}</p>
								<div class="flex items-center gap-2 text-xs text-ink-4">
									{#if v.parentName && v.parentSlug && onNavigateDir}
										<span
											role="button"
											tabindex="0"
											class="flex cursor-pointer items-center gap-1 text-primary hover:text-primary hover:underline"
											onclick={(e) => { e.stopPropagation(); onNavigateDir(v.parentSlug!); }}
											onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.stopPropagation(); onNavigateDir(v.parentSlug!); } }}
										>
											<Folder size={12} />
											{v.parentName}
										</span>
									{:else if v.parentName}
										<span class="flex items-center gap-1">
											<Folder size={12} />
											{v.parentName}
										</span>
									{/if}
									<span>{fmtSize(v.fileSize)}</span>
								</div>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex items-center justify-between border-t border-line-soft px-5 py-3">
			<span class="text-xs text-ink-4">
				{#if selectedCount() > 0}
					{m.add_selected({ count: selectedCount() })}
				{/if}
			</span>
			<div class="flex gap-2">
				<button type="button" onclick={() => { open = false; onClose(); }} class="rounded-lg px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken">
					{m.cancel()}
				</button>
				<button type="button" onclick={submit} disabled={selectedCount() === 0 || submitting} class="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:opacity-50">
					{#if submitting}
						<LoaderCircle size={14} class="animate-spin" />
					{/if}
					{m.add_selected({ count: selectedCount() })}
				</button>
			</div>
		</div>
</Dialog>
