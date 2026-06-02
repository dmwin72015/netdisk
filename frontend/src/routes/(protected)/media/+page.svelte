<script lang="ts">
	import { onMount, onDestroy, getContext } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { addToLibrary, ensureMediaUploadDir, listMedia, removeFromLibrary, type MediaItem } from '$lib/api/media';
	import { Film, Trash2, LoaderCircle, Play, CircleAlert, Clock, Plus, Upload, ChevronDown } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { confirmDelete } from '$lib/dialog';
	import { fmtDurationText, authedUrl } from '$lib/utils/format';
	import AddMediaDialog from '$lib/components/media/AddMediaDialog.svelte';
	import { Popover } from '$lib/ui/popover';
	import type { createUploadManager as UploadMgrFn } from '$lib/upload-manager.svelte';
	type UploadManager = ReturnType<typeof UploadMgrFn>;
	import * as m from '$lib/paraglide/messages';

	let items = $state<MediaItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let showAddDialog = $state(false);
	let showMenu = $state(false);
	let menuTimer: ReturnType<typeof setTimeout> | undefined;
	let groupWidth = $state(0);
	let videoInput: HTMLInputElement | undefined = $state();
	let pollTimer: ReturnType<typeof setInterval> | undefined;
	let refreshTimer: ReturnType<typeof setTimeout> | undefined;

	function isVideoFile(file: File) {
		if (file.type.startsWith('video/')) return true;
		return /\.(mp4|mov|webm|mkv|avi|flv|wmv|ogv|ogg|mpeg|mpg|m4v)$/i.test(file.name);
	}

	const upload = getContext<UploadManager>('upload');

	$effect(() => {
		upload.setAcceptFile(isVideoFile);
		upload.setGetCurrentSlug(async () => {
			const dir = await ensureMediaUploadDir();
			return dir.slug;
		});
		upload.setOnRejected((files) => {
			toast.error(m.media_upload_rejected({ count: files.length }));
		});
		upload.setOnFileImported(async ({ fileSlug }) => {
			try {
				const resp = await addToLibrary(fileSlug);
				if (resp.transcodeStatus === 'existing') {
					toast.info(m.media_already_in_library());
				} else {
					toast.success(m.media_transcode_started());
				}
				scheduleRefresh();
				startPolling();
			} catch (e) {
				toast.error(e instanceof Error ? e.message : m.media_add_failed());
				throw e;
			}
		});
		upload.setOnCompleted(async () => {
			await refresh(false);
		});
	});

	function startPolling() {
		if (pollTimer) return;
		pollTimer = setInterval(async () => {
			if (!$user) return;
			try {
				const data = await listMedia();
				items = data.items;
				total = data.total;
				if (!items.some(i => i.status === 'processing' || i.status === 'pending')) {
					stopPolling();
				}
			} catch {
				// ignore poll errors
			}
		}, 3000);
	}

	function stopPolling() {
		if (pollTimer) {
			clearInterval(pollTimer);
			pollTimer = undefined;
		}
	}

	function scheduleRefresh() {
		clearTimeout(refreshTimer);
		refreshTimer = setTimeout(() => void refresh(false), 250);
	}

	async function refresh(showLoading = true) {
		if (!$user) return;
		if (showLoading) loading = true;
		try {
			const data = await listMedia();
			items = data.items;
			total = data.total;
			if (items.some(i => i.status === 'processing' || i.status === 'pending')) {
				startPolling();
			}
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_load_failed());
		} finally {
			if (showLoading) loading = false;
		}
	}

	onDestroy(() => {
		stopPolling();
		clearTimeout(refreshTimer);
	});

	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_remove_media({ name })))) return;
		try {
			await removeFromLibrary(slug);
			items = items.filter(i => i.mediaSlug !== slug);
			total--;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_remove_failed());
		}
	}

	onMount(() => {
		void refresh();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4">
		<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
			<div class="flex items-center gap-2">
				<Film size={20} class="text-gray-500" />
				<h1 class="text-lg font-semibold text-gray-900">{m.media_title()}</h1>
				<span class="text-sm text-gray-400">{m.total_items({ total })}</span>
			</div>
			<div class="relative" role="region"
				onmouseenter={() => { clearTimeout(menuTimer); showMenu = true; }}
				onmouseleave={() => { menuTimer = setTimeout(() => { showMenu = false; }, 200); }}
			>
				<div class="flex h-8 items-center overflow-hidden rounded-lg bg-blue-600 text-sm font-medium text-white shadow-sm transition-colors" bind:clientWidth={groupWidth}>
					<button type="button" onclick={() => videoInput?.click()}
						class="flex h-full items-center gap-1.5 bg-blue-600 px-3.5 hover:bg-blue-700 active:bg-blue-800"
					>
						<Upload size={15} /> {m.upload_video()}
					</button>
					<Popover
						bind:open={showMenu}
						triggerClass="flex h-full items-center px-1.5 bg-blue-600 hover:bg-blue-700 active:bg-blue-800"
						contentClass="w-auto max-w-44 p-1.5"
						contentStyle="min-width: {groupWidth}px"
						sideOffset={4}
						align="end"
						onOpenChange={(o) => { if (!o) showMenu = false; }}
					>
						{#snippet trigger()}
							<ChevronDown size={14} />
						{/snippet}
						<div role="region"
							onmouseenter={() => clearTimeout(menuTimer)}
							onmouseleave={() => { menuTimer = setTimeout(() => { showMenu = false; }, 200); }}
						>
							<button type="button" class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-gray-700 outline-none hover:bg-gray-50 focus-visible:outline-none"
								onclick={() => { showMenu = false; showAddDialog = true; }}
							>
								<Plus size={15} class="shrink-0 text-gray-500" /> <span class="truncate">{m.add_to_media_library()}</span>
							</button>
						</div>
					</Popover>
				</div>
			</div>
		</div>
		<input
			bind:this={videoInput}
			type="file"
			accept="video/*,.mkv,.avi,.flv,.wmv,.ogv,.ogg,.mpeg,.mpg,.m4v"
			multiple
			class="hidden"
			onchange={upload.onPick}
		/>
		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if items.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
				<Film size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.media_empty()}</p>
				<p class="mt-1 text-xs text-gray-300">{m.media_help()}</p>
			</div>
		{:else}
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each items as item (item.mediaSlug)}
					<div class="group relative overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm transition-all hover:border-gray-200 hover:shadow-md {item.status === 'done' ? '' : 'cursor-default'}">
						<!-- Thumbnail / status area -->
						<div class="relative aspect-video bg-gray-100">
							{#if item.status === 'done'}
								<a href="/media/{item.mediaSlug}" class="block h-full">
									{#if item.posterUrl}
										<img src={authedUrl(item.posterUrl)} alt={item.fileName} class="h-full w-full object-cover transition group-hover:scale-105" loading="lazy" />
										<div class="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/0 transition group-hover:bg-black/30">
											<Play size={40} class="text-white opacity-0 transition group-hover:opacity-100" fill="currentColor" />
										</div>
									{:else}
										<div class="flex h-full items-center justify-center">
											<div class="flex h-12 w-12 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur transition-transform group-hover:scale-110">
												<Play size={20} fill="currentColor" />
											</div>
										</div>
									{/if}
								</a>
							{:else}
								<div class="flex h-full flex-col items-center justify-center gap-2">
									{#if item.status === 'processing'}
										<LoaderCircle size={24} class="animate-spin text-blue-400" />
										<span class="text-xs text-blue-500">{item.progress}%</span>
									{:else if item.status === 'pending'}
										<Clock size={24} class="text-gray-300" />
										<span class="text-xs text-gray-400">{m.queued()}</span>
									{:else if item.status === 'failed'}
										<CircleAlert size={24} class="text-red-300" />
										<span class="text-xs text-red-400">{m.failed()}</span>
									{/if}
								</div>
							{/if}

							<!-- Duration -->
							{#if item.durationSec}
								<div class="absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
									{fmtDurationText(item.durationSec)}
								</div>
							{/if}
						</div>

						<!-- Info -->
						<div class="px-3 py-2.5">
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-gray-700" title={item.fileName}>{item.fileName}</p>
									<p class="mt-0.5 text-xs text-gray-400">
										{new Date(item.createdAt).toLocaleDateString()}
									</p>
								</div>
								<button type="button" onclick={() => remove(item.mediaSlug, item.fileName)} class="shrink-0 rounded-md p-1 text-gray-400 opacity-0 transition-all hover:text-red-500 group-hover:opacity-100" title={m.remove()}>
									<Trash2 size={14} />
								</button>
							</div>
							{#if item.status === 'failed' && item.errorMsg}
								<p class="mt-1 truncate text-xs text-red-500" title={item.errorMsg}>{item.errorMsg}</p>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<AddMediaDialog
		open={showAddDialog}
		onClose={() => (showAddDialog = false)}
		onDone={refresh}
	/>
{/if}
