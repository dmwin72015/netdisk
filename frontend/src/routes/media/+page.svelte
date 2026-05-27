<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { listMedia, removeFromLibrary, type MediaItem } from '$lib/api/media';
	import { Film, Trash2, Loader2, Play, AlertCircle, Clock, Plus } from '@lucide/svelte';
	import { confirmDelete } from '$lib/dialog';
	import AddMediaDialog from '$lib/components/media/AddMediaDialog.svelte';
	import * as m from '$lib/paraglide/messages';

	let items = $state<MediaItem[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let showAddDialog = $state(false);

	async function refresh() {
		if (!$user) return;
		loading = true;
		error = null;
		try {
			const data = await listMedia();
			items = data.items;
			total = data.total;
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_load_failed();
		} finally {
			loading = false;
		}
	}

	async function remove(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_remove_media({ name })))) return;
		try {
			await removeFromLibrary(slug);
			items = items.filter(i => i.media_slug !== slug);
			total--;
		} catch (e) {
			error = e instanceof Error ? e.message : m.media_remove_failed();
		}
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'done': return 'text-green-600 bg-green-50';
			case 'processing': return 'text-blue-600 bg-blue-50';
			case 'pending': return 'text-gray-500 bg-gray-50';
			case 'failed': return 'text-red-600 bg-red-50';
			default: return 'text-gray-500 bg-gray-50';
		}
	}

	function statusIcon(status: string) {
		switch (status) {
			case 'done': return Play;
			case 'processing': return Loader2;
			case 'pending': return Clock;
			case 'failed': return AlertCircle;
			default: return Clock;
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

		{#if error}
			<div class="flex items-start gap-2.5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">
				<AlertCircle size={16} class="mt-0.5 shrink-0" />
				<span>{error}</span>
			</div>
		{/if}

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
				{#each items as item (item.media_slug)}
					<div class="group relative overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm transition-all hover:border-gray-200 hover:shadow-md">
						<!-- Thumbnail / status area -->
						<div class="relative aspect-video bg-gray-100">
							{#if item.status === 'done'}
								<a href="/media/{item.media_slug}" class="flex h-full items-center justify-center">
									<div class="flex h-12 w-12 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur transition-transform group-hover:scale-110">
										<Play size={20} fill="currentColor" />
									</div>
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

							<!-- Status badge -->
							<div class="absolute left-2 top-2">
								<span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium {statusColor(item.status)}">
									{#if item.status === 'processing'}
										<Loader2 size={10} class="animate-spin" />
									{/if}
									{item.status}
								</span>
							</div>

							<!-- Duration -->
							{#if item.duration_sec}
								<div class="absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
									{fmtDuration(item.duration_sec)}
								</div>
							{/if}
						</div>

						<!-- Info -->
						<div class="px-3 py-2.5">
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium text-gray-700" title={item.file_name}>{item.file_name}</p>
									<p class="mt-0.5 text-xs text-gray-400">
										{new Date(item.created_at).toLocaleDateString()}
									</p>
								</div>
								<button type="button" onclick={() => remove(item.media_slug, item.file_name)} class="shrink-0 rounded-md p-1 text-gray-400 opacity-0 transition-all hover:text-red-500 group-hover:opacity-100" title={m.remove()}>
									<Trash2 size={14} />
								</button>
							</div>
							{#if item.status === 'failed' && item.error_msg}
								<p class="mt-1 truncate text-xs text-red-500" title={item.error_msg}>{item.error_msg}</p>
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
