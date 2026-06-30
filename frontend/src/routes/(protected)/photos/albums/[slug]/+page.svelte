<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { user, authReady } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import noFilesSvg from '$lib/assets/empty-states/no-files.svg';
	import { toast } from 'svelte-sonner';
	import { LoaderCircle, ArrowLeft, Trash2 } from '@lucide/svelte';
	import { authedUrl } from '$lib/utils/format';
	import { getAlbum, removePhotoFromAlbum, type AlbumDetail } from '$lib/api/albums';
	import { thumbnailUrl } from '$lib/api/photos';
	import PhotoViewer from '$lib/components/PhotoViewer.svelte';

	let album = $state<AlbumDetail | null>(null);
	let loading = $state(true);
	let viewerSlug = $state<string | null>(null);
	let viewerOpen = $state(false);
	let viewerIndex = $state(0);

	let allSlugs = $derived(album?.photos.map(p => p.slug) ?? []);
	let albumSlug = $derived($page.params.slug ?? '');

	async function fetch() {
		loading = true;
		try {
			album = await getAlbum(albumSlug);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loading = false;
		}
	}

	async function handleRemove(fileSlug: string) {
		try {
			await removePhotoFromAlbum(albumSlug, fileSlug);
			if (album) {
				album.photos = album.photos.filter(p => p.slug !== fileSlug);
				album.itemCount = album.photos.length;
			}
			toast.success(m.albums_remove_photo());
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		}
	}

	function openViewer(slug: string) {
		viewerIndex = allSlugs.indexOf(slug);
		viewerSlug = slug;
		viewerOpen = true;
	}

	onMount(() => {
		void fetch();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4 px-6 pt-4 pb-6">
		<!-- Header -->
		<div class="flex items-center gap-3">
			<a href="/photos/albums" class="rounded-md p-1 text-ink-4 hover:text-ink-3">
				<ArrowLeft size={20} />
			</a>
			{#if album}
				<div>
					<h1 class="text-lg font-semibold text-ink">{album.title}</h1>
					<p class="text-xs text-ink-4">
						{m.albums_photos_count({ count: album.itemCount })}
						{#if album.description}
							&middot; {album.description}
						{/if}
					</p>
				</div>
			{:else if !loading}
				<h1 class="text-lg font-semibold text-ink">Not Found</h1>
			{/if}
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if album && album.photos.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center">
				<img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
				<p class="text-sm text-ink-4">{m.photos_empty()}</p>
			</div>
		{:else if album}
			<div class="grid grid-cols-3 gap-1 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6">
				{#each album.photos as photo (photo.slug)}
					<div class="group relative">
						<button
							type="button"
							onclick={() => openViewer(photo.slug)}
							class="aspect-square w-full overflow-hidden rounded-md bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
						>
							<img
								src={authedUrl(thumbnailUrl(photo.slug))}
								alt={photo.fileName}
								class="h-full w-full object-cover transition group-hover:scale-105"
								loading="lazy"
							/>
						</button>
						<button
							type="button"
							onclick={() => handleRemove(photo.slug)}
							class="absolute right-1.5 top-1.5 rounded-full bg-surface/80 p-1.5 text-ink-4 opacity-0 shadow transition-opacity hover:text-danger group-hover:opacity-100"
							title={m.albums_remove_photo()}
						>
							<Trash2 size={12} />
						</button>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<PhotoViewer
		bind:open={viewerOpen}
		slug={viewerSlug}
		bind:fileSlugs={allSlugs}
		index={viewerIndex}
		onClose={() => (viewerSlug = null)}
		photos={album?.photos ?? []}
	/>
{/if}
