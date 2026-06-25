<script lang="ts">
	import { onMount, onDestroy, getContext } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { ApiError, getAccessToken } from '$lib/api/client';
	import {
		addToLibrary,
		ensureMediaUploadDir,
		listMedia,
		removeFromLibrary,
		batchRemoveFromLibrary,
		renameMediaItem,
		getMediaItem,
		readdExistingUploadToLibrary,
		type AddToLibraryResponse,
		type MediaItem
	} from '$lib/api/media';
	import { Trash2, LoaderCircle, Play, CircleAlert, Clock, Plus, Upload, ChevronDown, Check, X, Pencil } from '@lucide/svelte';
	import { fly } from 'svelte/transition';
	import { toast } from 'svelte-sonner';
	import { confirmAction, confirmDelete, promptInput } from '$lib/dialog';
	import { fmtDurationText, authedUrl } from '$lib/utils/format';
	import AddMediaDialog from '$lib/components/media/AddMediaDialog.svelte';
	import PasteUploadProvider from '$lib/components/files/PasteUploadProvider.svelte';
	import { Popover } from '$lib/ui/popover';
	import type { createUploadManager as UploadMgrFn } from '$lib/upload-manager.svelte';
	type UploadManager = ReturnType<typeof UploadMgrFn>;
	import * as m from '$lib/paraglide/messages';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';

	let items = $state<MediaItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let showAddDialog = $state(false);
	let showMenu = $state(false);
	let menuTimer: ReturnType<typeof setTimeout> | undefined;
	let groupWidth = $state(0);
	let videoInput: HTMLInputElement | undefined = $state();
	let es: EventSource | undefined;
	let refreshTimer: ReturnType<typeof setTimeout> | undefined;
	let selected = $state<Set<string>>(new Set());
	let allSelected = $derived(items.length > 0 && items.every(item => selected.has(item.mediaSlug)));
	let hasSelection = $derived(selected.size > 0);
	let selectedItems = $derived(items.filter(item => selected.has(item.mediaSlug)));
	const ERR_CODE_NAME_CONFLICT = 2004;

	function isVideoFile(file: File) {
		if (file.type.startsWith('video/')) return true;
		return /\.(mp4|mov|webm|mkv|avi|flv|wmv|ogv|ogg|mpeg|mpg|m4v)$/i.test(file.name);
	}

	function isNameConflict(error: unknown) {
		return error instanceof ApiError && error.errCode === ERR_CODE_NAME_CONFLICT;
	}

	function notifyMediaAdd(resp: AddToLibraryResponse) {
		if (resp.alreadyInLibrary) {
			toast.info(m.media_already_in_library());
		} else if (resp.transcodeReused) {
			toast.success(m.media_add_success());
		} else {
			toast.success(m.media_transcode_started());
		}
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
		upload.setOnDuplicateDetected(() => true);
		upload.setOnFileImported(async ({ fileSlug }) => {
			try {
				const resp = await addToLibrary(fileSlug);
				notifyMediaAdd(resp);
				scheduleRefresh();
			} catch (e) {
				toast.error(e instanceof Error ? e.message : m.media_add_failed());
				throw e;
			}
		});
		upload.setOnImportConflict(async ({ physicalFileSlug, fileName, source, error }) => {
			if (source !== 'dedup' || !isNameConflict(error)) return false;

			const confirmed = await confirmAction(
				m.media_existing_file_title(),
				m.media_existing_file_message({ name: fileName }),
				m.media_readd_existing_btn()
			);
			if (!confirmed) {
				throw new Error(m.upload_skipped_duplicate());
			}

			try {
				const resp = await readdExistingUploadToLibrary(physicalFileSlug, fileName);
				notifyMediaAdd(resp);
				scheduleRefresh();
				return true;
			} catch (e) {
				toast.error(e instanceof Error ? e.message : m.media_add_failed());
				throw e;
			}
		});
		upload.setOnCompleted(async () => {
			await refresh(false);
		});
	});

	function connectSSE() {
		if (es) return;
		const token = getAccessToken();
		if (!token) return;
		const url = new URL('/api/v1/media/events', window.location.origin);
		url.searchParams.set('access_token', token);
		es = new EventSource(url.toString());

		function updateItem(mediaSlug: string, update: Partial<MediaItem>) {
			const idx = items.findIndex(i => i.mediaSlug === mediaSlug);
			if (idx !== -1) {
				items[idx] = { ...items[idx], ...update };
			}
		}

		es.addEventListener('processing', (e) => {
			const data = JSON.parse(e.data);
			updateItem(data.mediaSlug, { status: data.status, progress: data.progress });
		});

		es.addEventListener('done', (e) => {
			const data = JSON.parse(e.data);
			updateItem(data.mediaSlug, { status: data.status, progress: 100 });
			void refreshItem(data.mediaSlug);
		});

		es.addEventListener('failed', (e) => {
			const data = JSON.parse(e.data);
			updateItem(data.mediaSlug, { status: data.status, progress: 0, errorMsg: data.errorMsg ?? null });
		});

		es.onerror = () => {
			// EventSource auto-reconnects
		};
	}

	function disconnectSSE() {
		if (es) {
			es.close();
			es = undefined;
		}
	}

	async function refreshItem(mediaSlug: string) {
		try {
			const updated = await getMediaItem(mediaSlug);
			const idx = items.findIndex(i => i.mediaSlug === mediaSlug);
			if (idx !== -1) {
				items[idx] = updated;
			}
		} catch {
			// ignore
		}
	}

	function scheduleRefresh() {
		clearTimeout(refreshTimer);
		refreshTimer = setTimeout(() => {
			disconnectSSE();
			void refresh(false);
		}, 250);
	}

	async function refresh(showLoading = true) {
		if (!$user) return;
		if (showLoading) loading = true;
		try {
			const data = await listMedia();
			items = data.items;
			total = data.total;
			const itemSlugs = new Set(items.map(item => item.mediaSlug));
			selected = new Set(Array.from(selected).filter(slug => itemSlugs.has(slug)));
			connectSSE();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_load_failed());
		} finally {
			if (showLoading) loading = false;
		}
	}

	onDestroy(() => {
		disconnectSSE();
		clearTimeout(refreshTimer);
		upload.setAcceptFile();
		upload.setOnRejected();
		upload.setOnFileImported();
		upload.setOnImportConflict();
		upload.setOnDuplicateDetected();
	});

	function toggleSelect(mediaSlug: string) {
		if (selected.has(mediaSlug)) selected.delete(mediaSlug);
		else selected.add(mediaSlug);
		selected = new Set(selected);
	}

	function toggleSelectAll() {
		selected = allSelected ? new Set() : new Set(items.map(item => item.mediaSlug));
	}

	function clearSelection() {
		selected = new Set();
	}

	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_remove_media({ name })))) return;
		try {
			await removeFromLibrary(slug);
			items = items.filter(i => i.mediaSlug !== slug);
			total--;
			selected.delete(slug);
			selected = new Set(selected);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_remove_failed());
		}
	}

	async function batchRemove() {
		const targets = selectedItems;
		if (targets.length === 0) return;
		const names = targets.map(item => item.fileName);
		if (!(await confirmDelete(m.confirm_delete_multiple({ count: String(targets.length), names: names.join('\n') })))) return;
		try {
			const slugs = targets.map(item => item.mediaSlug);
			await batchRemoveFromLibrary(slugs);
			items = items.filter(item => !selected.has(item.mediaSlug));
			total = Math.max(0, total - targets.length);
			clearSelection();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.media_remove_failed());
		}
	}

	async function rename(slug: string, currentName: string) {
		const newName = await promptInput(m.rename(), m.enter_new_name(), currentName, 512);
		const trimmed = newName?.trim();
		if (!trimmed || trimmed === currentName) return;
		try {
			const updated = await renameMediaItem(slug, trimmed);
			const idx = items.findIndex(item => item.mediaSlug === slug);
			if (idx !== -1) items[idx] = updated;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.rename_failed());
		}
	}

	onMount(() => {
		void refresh();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4 rounded-xl border border-line bg-white p-4">
		<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
			<div class="flex items-center gap-2">
				<span class="text-sm text-ink-4">{m.total_items({ total })}</span>
			</div>
				<div class="flex items-center gap-2">
					{#if items.length > 0}
						<button type="button" onclick={toggleSelectAll} class="flex h-8 items-center gap-1.5 rounded-lg border border-line px-3 text-sm text-ink-3 transition-colors hover:bg-surface-muted hover:text-ink">
							<Check size={15} /> {allSelected ? m.clear_selection() : m.select_all()}
						</button>
					{/if}
					<div class="relative" role="region"
						onmouseenter={() => { clearTimeout(menuTimer); showMenu = true; }}
						onmouseleave={() => { menuTimer = setTimeout(() => { showMenu = false; }, 200); }}
					>
						<div class="flex h-8 items-center overflow-hidden rounded-lg bg-primary text-sm font-medium text-white transition-colors" bind:clientWidth={groupWidth}>
							<button type="button" onclick={() => videoInput?.click()}
								class="flex h-full items-center gap-1.5 bg-primary px-3.5 hover:bg-primary-hover active:bg-primary-active"
							>
								<Upload size={15} /> {m.upload_video()}
							</button>
							<Popover
								bind:open={showMenu}
								triggerClass="flex h-full items-center px-1.5 bg-primary hover:bg-primary-hover active:bg-primary-active"
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
									<button type="button" class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-ink-2 outline-none hover:bg-surface-muted focus-visible:outline-none"
										onclick={() => { showMenu = false; showAddDialog = true; }}
									>
										<Plus size={15} class="shrink-0 text-ink-3" /> <span class="truncate">{m.add_to_media_library()}</span>
									</button>
								</div>
							</Popover>
						</div>
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

		<PasteUploadProvider
			targetLabel={m.media_title()}
			acceptFile={isVideoFile}
			onUpload={(files) => upload.enqueueFiles(files)}
		/>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if items.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center">
				<img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
				<p class="text-sm text-ink-4">{m.media_empty()}</p>
				<p class="mt-1 text-xs text-ink-4">{m.media_help()}</p>
			</div>
			{:else}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each items as item (item.mediaSlug)}
						{@const isSelected = selected.has(item.mediaSlug)}
						<div class="group relative overflow-hidden rounded-xl border border-line bg-white transition-all hover:border-line {isSelected ? 'border-primary ring-2 ring-primary-soft' : ''} {item.status === 'done' ? '' : 'cursor-default'}">
							<button
								type="button"
								aria-pressed={isSelected}
								onclick={(e) => { e.stopPropagation(); toggleSelect(item.mediaSlug); }}
								class="absolute left-2 top-2 z-10 flex h-6 w-6 items-center justify-center rounded-md border text-white transition-all {isSelected ? 'border-primary bg-primary opacity-100' : 'border-white/80 bg-black/30 opacity-0 hover:bg-black/45 group-hover:opacity-100'}"
							>
								{#if isSelected}<Check size={14} strokeWidth={3} />{/if}
							</button>
							<!-- Thumbnail / status area -->
							<div class="relative aspect-video overflow-hidden bg-surface-sunken">
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
										<LoaderCircle size={24} class="animate-spin text-primary" />
										<span class="text-xs text-primary">{item.progress}%</span>
									{:else if item.status === 'pending'}
										<Clock size={24} class="text-ink-4" />
										<span class="text-xs text-ink-4">{m.queued()}</span>
									{:else if item.status === 'failed'}
										<CircleAlert size={24} class="text-danger" />
										<span class="text-xs text-danger">{m.failed()}</span>
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
									<p class="truncate text-sm font-medium text-ink-2" title={item.fileName}>{item.fileName}</p>
									<p class="mt-0.5 text-xs text-ink-4">
										{new Date(item.createdAt).toLocaleDateString()}
									</p>
								</div>
									<div class="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100 {isSelected ? 'opacity-100' : ''}">
										<button type="button" onclick={() => rename(item.mediaSlug, item.fileName)} class="rounded-md p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-primary" title={m.rename()}>
											<Pencil size={14} />
										</button>
										<button type="button" onclick={() => remove(item.mediaSlug, item.fileName)} class="rounded-md p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-danger" title={m.remove()}>
											<Trash2 size={14} />
										</button>
									</div>
								</div>
							{#if item.status === 'failed' && item.errorMsg}
								<p class="mt-1 truncate text-xs text-danger" title={item.errorMsg}>{item.errorMsg}</p>
							{/if}
						</div>
					</div>
					{/each}
				</div>
			{/if}

			{#if hasSelection}
				<div class="fixed bottom-6 left-1/2 z-50 max-w-[calc(100vw-1rem)] -translate-x-1/2">
					<div
						class="flex items-center gap-2 overflow-x-auto rounded-full border border-line-soft bg-white/95 px-3 py-2 shadow-[0_12px_36px_rgba(15,23,42,0.16)] backdrop-blur"
						transition:fly={{ y: 16, duration: 180, opacity: 0 }}
					>
						<span class="shrink-0 px-3 text-sm font-medium text-ink-2">{m.selected_count({ count: String(selected.size) })}</span>
						<div class="h-7 w-px shrink-0 bg-surface-sunken"></div>
						<button type="button" onclick={batchRemove} class="flex h-8 w-8 items-center justify-center rounded-full text-ink-3 transition-colors hover:bg-danger-soft hover:text-danger" title={m.delete_label()}>
							<Trash2 size={16} />
						</button>
						<div class="mx-1 h-7 w-px bg-surface-sunken"></div>
						<button type="button" onclick={clearSelection} class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2" title={m.close()}>
							<X size={16} />
						</button>
					</div>
				</div>
			{/if}
		</div>

		<AddMediaDialog
		open={showAddDialog}
		onClose={() => (showAddDialog = false)}
		onDone={refresh}
	/>
{/if}
