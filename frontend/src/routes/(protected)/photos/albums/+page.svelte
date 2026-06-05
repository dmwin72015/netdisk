<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';
	import { Folder as AlbumIcon, Image, LoaderCircle, Trash2, Plus } from '@lucide/svelte';
	import { listAlbums, deleteAlbum, type Album } from '$lib/api/albums';
	import AlbumCreateDialog from '$lib/components/AlbumCreateDialog.svelte';

	let albums = $state<Album[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);

	async function fetch() {
		loading = true;
		try {
			const data = await listAlbums(1, 1000);
			albums = data.items;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loading = false;
		}
	}

	async function handleDelete(albumSlug: string) {
		if (!confirm(m.albums_delete_confirm())) return;
		try {
			await deleteAlbum(albumSlug);
			albums = albums.filter(a => a.slug !== albumSlug);
			toast.success(m.albums_delete_success());
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		}
	}

	function onCreated() {
		void fetch();
	}

	onMount(() => {
		void fetch();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-2">
				<AlbumIcon size={20} class="text-gray-500" />
				<h1 class="text-lg font-semibold text-gray-900">{m.albums_title()}</h1>
			</div>
			<button
				type="button"
				onclick={() => (showCreate = true)}
				class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
			>
				<Plus size={15} />
				{m.albums_create()}
			</button>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if albums.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
				<Image size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.albums_empty()}</p>
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
				{#each albums as album (album.slug)}
					<a
						href="/photos/albums/{album.slug}"
						class="group relative block overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm transition-shadow hover:shadow-md"
					>
						<div class="aspect-[4/3] bg-gradient-to-br from-gray-100 to-gray-200">
							{#if album.coverUrl}
								<img src={album.coverUrl} alt={album.title} class="h-full w-full object-cover" />
							{/if}
						</div>
						<div class="p-3">
							<h3 class="truncate text-sm font-medium text-gray-900">{album.title}</h3>
							<p class="text-xs text-gray-400">{m.albums_photos_count({ count: album.itemCount })}</p>
						</div>
						<button
							type="button"
							onclick={(e) => { e.preventDefault(); handleDelete(album.slug); }}
							class="absolute right-2 top-2 rounded-full bg-white/80 p-1.5 text-gray-400 opacity-0 shadow transition-opacity hover:text-red-500 group-hover:opacity-100"
						>
							<Trash2 size={14} />
						</button>
					</a>
				{/each}
			</div>
		{/if}
	</div>

	<AlbumCreateDialog bind:show={showCreate} onCreated={onCreated} />
{/if}
