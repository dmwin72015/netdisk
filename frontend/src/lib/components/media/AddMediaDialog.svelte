<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { Loader2, Film, Check, X, Search } from '@lucide/svelte';
	import { listFiles, type FileItem } from '$lib/api/files';
	import { addToLibrary } from '$lib/api/media';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let {
		open,
		onClose,
		onDone
	}: {
		open: boolean;
		onClose: () => void;
		onDone: () => void;
	} = $props();

	let videos = $state<FileItem[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let page = $state(1);
	let selected = $state<Set<string>>(new Set());
	let submitting = $state(false);
	let error = $state<string | null>(null);
	let searchQuery = $state('');
	const PAGE_SIZE = 20;

	async function fetchVideos(p: number) {
		loading = true;
		error = null;
		try {
			const data = await listFiles(undefined, p, PAGE_SIZE, 'video/');
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
			? videos.filter(v => v.file_name.toLowerCase().includes(searchQuery.trim().toLowerCase()))
			: videos
	);

	$effect(() => {
		if (open) {
			selected.clear();
			page = 1;
			searchQuery = '';
			fetchVideos(1);
		}
	});

	function toggle(slug: string) {
		const next = new Set(selected);
		if (next.has(slug)) next.delete(slug);
		else next.add(slug);
		selected = next;
	}

	function toggleAll() {
		if (selected.size === filteredVideos.length) {
			selected = new Set();
		} else {
			selected = new Set(filteredVideos.map(v => v.slug));
		}
	}

	async function submit() {
		if (selected.size === 0) return;
		submitting = true;
		error = null;
		try {
			for (const slug of selected) {
				await addToLibrary(slug);
			}
			onDone();
			onClose();
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_add_failed();
		} finally {
			submitting = false;
		}
	}

	let closing = $state(false);
	function closeWithAnimation() {
		closing = true;
		setTimeout(() => {
			closing = false;
			onClose();
		}, 150);
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm {closing ? 'animate-out fade-out-0' : 'animate-in fade-in-0'}" onclick={(e) => { if (e.target === e.currentTarget) closeWithAnimation(); }}>
		<div class="flex h-[70vh] w-full max-w-2xl flex-col rounded-xl border border-gray-100 bg-white shadow-2xl {closing ? 'animate-out fade-out-0 zoom-out-95' : 'animate-in fade-in-0 zoom-in-95'}">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5">
				<h2 class="text-base font-semibold text-gray-900">{m.select_videos()}</h2>
				<button type="button" onclick={closeWithAnimation} class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
					<X size={18} />
				</button>
			</div>

			<!-- Search -->
			<div class="border-b border-gray-100 px-5 py-2.5">
				<div class="relative">
					<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
					<input
						type="text"
						bind:value={searchQuery}
						placeholder={m.search_files()}
						class="w-full rounded-lg border border-gray-200 bg-gray-50 py-2 pl-9 pr-3 text-sm text-gray-700 outline-none transition-colors placeholder:text-gray-400 focus:border-blue-400 focus:bg-white"
					/>
				</div>
			</div>

			<!-- Content -->
			<div class="flex-1 overflow-y-auto p-4">
				{#if error}
					<div class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</div>
				{/if}

				{#if loading}
					<div class="flex items-center justify-center py-16">
						<Loader2 size={24} class="animate-spin text-gray-300" />
					</div>
				{:else if filteredVideos.length === 0}
					<div class="flex flex-col items-center justify-center py-16 text-center">
						<Film size={40} class="mb-3 text-gray-300" />
						<p class="text-sm text-gray-400">{searchQuery.trim() ? m.no_videos_found() : m.no_videos_found()}</p>
					</div>
				{:else}
					<!-- Select all -->
					<button type="button" onclick={toggleAll} class="mb-2 text-xs text-blue-600 transition-colors hover:text-blue-700">
						{selected.size === filteredVideos.length ? m.cancel() : m.add_selected({ count: filteredVideos.length })}
					</button>

					<div class="space-y-1">
						{#each filteredVideos as v (v.slug)}
							<button type="button" onclick={() => toggle(v.slug)} class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors hover:bg-gray-50 {selected.has(v.slug) ? 'bg-blue-50' : ''}">
								<div class="flex h-5 w-5 shrink-0 items-center justify-center rounded border {selected.has(v.slug) ? 'border-blue-500 bg-blue-500 text-white' : 'border-gray-300'}">
									{#if selected.has(v.slug)}
										<Check size={12} />
									{/if}
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm text-gray-700">{v.file_name}</p>
									<p class="text-xs text-gray-400">{fmtSize(v.file_size)}</p>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-between border-t border-gray-100 px-5 py-3">
				<span class="text-xs text-gray-400">
					{#if selected.size > 0}
						{m.add_selected({ count: selected.size })}
					{/if}
				</span>
				<div class="flex gap-2">
					<button type="button" onclick={closeWithAnimation} class="rounded-lg px-4 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100">
						{m.cancel()}
					</button>
					<button type="button" onclick={submit} disabled={selected.size === 0 || submitting} class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 disabled:opacity-50">
						{#if submitting}
							<Loader2 size={14} class="animate-spin" />
						{/if}
						{m.add_selected({ count: selected.size })}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
