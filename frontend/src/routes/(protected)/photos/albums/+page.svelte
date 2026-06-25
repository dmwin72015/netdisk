<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';
	import { toast } from 'svelte-sonner';
	import { Folder as AlbumIcon, LoaderCircle, Trash2, Plus } from '@lucide/svelte';
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
				<AlbumIcon size={20} class="text-ink-3" />
				<h1 class="text-lg font-semibold text-ink">{m.albums_title()}</h1>
			</div>
			<button
				type="button"
				onclick={() => (showCreate = true)}
				class="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
			>
				<Plus size={15} />
				{m.albums_create()}
			</button>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if albums.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center">
				<img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
				<p class="text-sm text-ink-4">{m.albums_empty()}</p>
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
				{#each albums as album (album.slug)}
					<a
						href="/photos/albums/{album.slug}"
						class="group border-line bg-surface relative block overflow-hidden rounded-lg border transition-colors duration-150 hover:bg-surface-muted"
					>
						<div class="aspect-[4/3] bg-surface-sunken">
							{#if album.coverUrl}
									<img src={album.coverUrl} alt={album.title} loading="lazy" class="h-full w-full object-cover" />
							{/if}
						</div>
						<div class="p-3">
							<h3 class="truncate text-sm font-medium text-ink">{album.title}</h3>
							<p class="text-xs text-ink-4">{m.albums_photos_count({ count: album.itemCount })}</p>
						</div>
						<button
							type="button"
							onclick={(e) => { e.preventDefault(); handleDelete(album.slug); }}
							class="bg-surface/85 text-ink-4 hover:text-danger absolute right-2 top-2 rounded-md p-1.5 opacity-0 transition-opacity duration-150 group-hover:opacity-100"
						>
							<Trash2 size={14} />
						</button>
					</a>
				{/each}
			</div>
		{/if}
	</div>

	<AlbumCreateDialog bind:open={showCreate} onCreated={onCreated} />
{/if}
