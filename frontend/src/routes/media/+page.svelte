<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { listMedia, removeFromLibrary, type MediaItem } from '$lib/api/media';
	import { getAccessToken } from '$lib/api/client';
	import { Film, Trash2, Loader2, Play, AlertCircle, Clock, Plus } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { confirmDelete } from '$lib/dialog';
	import AddMediaDialog from '$lib/components/media/AddMediaDialog.svelte';
	import * as m from '$lib/paraglide/messages';

	let items = $state<MediaItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let showAddDialog = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | undefined;

	function authedPoster(url: string): string {
		const token = getAccessToken();
		if (!token) return url;
		const u = new URL(url, window.location.origin);
		u.searchParams.set('access_token', token);
		return u.pathname + '?' + u.searchParams.toString();
	}

	function startPolling() {
		stopPolling();
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

	async function refresh() {
		if (!$user) return;
		loading = true;
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
			loading = false;
		}
	}

	onDestroy(stopPolling);

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

	function fmtDuration(sec: number | null): string {
		if (!sec || sec <= 0) return '';
		const s = Math.round(sec);
		const h = Math.floor(s / 3600);
		const m = Math.floor((s % 3600) / 60);
		const r = s % 60;
		if (h > 0) return `${h}h ${m}m`;
		if (m > 0) return `${m}m ${r}s`;
		return `${r}s`;
	}

	onMount(() => {
		if (!$user) void goto('/login');
		else void refresh();
	});
</script>

{#if !$authReady}
{:else if $user}
	<div class="space-y-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-2">
				<Film size={20} class="text-gray-500" />
				<h1 class="text-lg font-semibold text-gray-900">{m.media_title()}</h1>
				<span class="text-sm text-gray-400">{m.total_items({ total })}</span>
			</div>
			<button type="button" onclick={() => (showAddDialog = true)} class="flex h-8 items-center gap-1.5 rounded-lg bg-blue-600 px-3.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 active:bg-blue-800">
				<Plus size={15} /> {m.add_to_media_library()}
			</button>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<Loader2 size={24} class="animate-spin text-gray-300" />
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
										<img src={authedPoster(item.posterUrl)} alt={item.fileName} class="h-full w-full object-cover transition group-hover:scale-105" loading="lazy" />
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
										<Loader2 size={24} class="animate-spin text-blue-400" />
										<span class="text-xs text-blue-500">{item.progress}%</span>
									{:else if item.status === 'pending'}
										<Clock size={24} class="text-gray-300" />
										<span class="text-xs text-gray-400">{m.queued()}</span>
									{:else if item.status === 'failed'}
										<AlertCircle size={24} class="text-red-300" />
										<span class="text-xs text-red-400">{m.failed()}</span>
									{/if}
								</div>
							{/if}

							<!-- Duration -->
							{#if item.durationSec}
								<div class="absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
									{fmtDuration(item.durationSec)}
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
{:else}
	<p class="text-gray-600">Please <a href="/login" class="text-blue-600 underline hover:text-blue-700">login</a> to continue.</p>
{/if}
