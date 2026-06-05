<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { authedUrl } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';
	import { Image as ImageIcon, LoaderCircle, Plus, ChevronDown, ArrowUpDown } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';

	import { listPhotos, thumbnailUrl, photoDetailUrl, type PhotoItem } from '$lib/api/photos';
	import PhotoViewer from '$lib/components/PhotoViewer.svelte';
	import AlbumCreateDialog from '$lib/components/AlbumCreateDialog.svelte';

	const PAGE_SIZE = 50;

	let photos = $state<PhotoItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let loadingMore = $state(false);
	let page = $state(1);



	let viewMode = $state<'grid' | 'list'>('grid');
	let photoSize = $state<'small' | 'medium' | 'large'>('medium');
	let groupByDate = $state(true);
	let sizeLabel = $derived(photoSize === 'large' ? m.photos_size_large() : photoSize === 'medium' ? m.photos_size_medium() : m.photos_size_small());

	// Lightbox
	let viewerSlug = $state<string | null>(null);
	let viewerIndex = $state(0);
	let allSlugs = $derived(photos.map(p => p.slug));

	// Album assignment
	let assignAlbumSlug = $state<string | null>(null);
	let assignPhotoSlugs = $state<string[]>([]);

	let showCreateAlbum = $state(false);

	let hasMore = $derived(photos.length < total);
	let gridClass = $derived(
		photoSize === 'large'
			? 'grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5'
			: photoSize === 'small'
				? 'grid-cols-4 gap-1 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-7 xl:grid-cols-8'
				: 'grid-cols-3 gap-1 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6'
	);

	$effect(() => {
		const savedView = localStorage.getItem('nd.photos.view');
		if (savedView === 'grid' || savedView === 'list') viewMode = savedView;
		const savedSize = localStorage.getItem('nd.photos.size');
		if (savedSize === 'small' || savedSize === 'medium' || savedSize === 'large') photoSize = savedSize;
		const savedGroup = localStorage.getItem('nd.photos.group');
		if (savedGroup !== null) groupByDate = savedGroup === 'true';
	});
	$effect(() => {
		localStorage.setItem('nd.photos.view', viewMode);
	});
	$effect(() => {
		localStorage.setItem('nd.photos.size', photoSize);
	});
	$effect(() => {
		localStorage.setItem('nd.photos.group', String(groupByDate));
	});

	function isSameDay(a: Date, b: Date): boolean {
		return a.getFullYear() === b.getFullYear()
			&& a.getMonth() === b.getMonth()
			&& a.getDate() === b.getDate();
	}

	function getDayLabel(isoDate: string): string {
		const d = new Date(isoDate);
		const today = new Date();
		const yesterday = new Date(today);
		yesterday.setDate(yesterday.getDate() - 1);
		if (isSameDay(d, today)) return m.photos_today();
		if (isSameDay(d, yesterday)) return m.photos_yesterday();
		return d.toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
	}

	type DayGroup = { date: string; label: string; items: PhotoItem[] };

	let groupedPhotos = $derived.by(() => {
		const groups = new Map<string, PhotoItem[]>();
		for (const f of photos) {
			const day = f.createdAt.slice(0, 10);
			if (!groups.has(day)) groups.set(day, []);
			groups.get(day)!.push(f);
		}
		return Array.from(groups.entries())
			.sort(([a], [b]) => b.localeCompare(a))
			.map(([date, items]): DayGroup => ({ date, label: getDayLabel(date), items }));
	});

	async function fetchPhotos() {
		loading = true;
		page = 1;
		try {
			const data = await listPhotos(1, PAGE_SIZE);
			photos = data.items;
			total = data.total;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;
		loadingMore = true;
		const nextPage = page + 1;
		try {
			const data = await listPhotos(nextPage, PAGE_SIZE);
			photos = [...photos, ...data.items];
			page = nextPage;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loadingMore = false;
		}
	}

	function openViewer(slug: string) {
		viewerIndex = allSlugs.indexOf(slug);
		viewerSlug = slug;
	}

	onMount(() => {
		void fetchPhotos();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4">
			<!-- Header -->
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<ImageIcon size={20} class="text-gray-500" />
					<h1 class="text-lg font-semibold text-gray-900">{m.photos_title()}</h1>
					<span class="text-sm text-gray-400">{m.photos_total({ total })}</span>
				</div>
				<div class="flex items-center gap-2">
					<Dropdown
						triggerClass="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 text-sm text-gray-600 transition-colors hover:border-gray-300 hover:bg-gray-50"
						contentClass="min-w-[180px]"
					>
						{#snippet trigger()}
							<ArrowUpDown size={14} />
							<span class="hidden sm:inline">{sizeLabel}</span>
							<ChevronDown size={14} class="text-gray-400" />
						{/snippet}
						{#snippet checkmark(active: boolean)}
							<svg viewBox="0 0 24 24" class="h-3.5 w-3.5 shrink-0 {active ? 'text-blue-600' : 'text-transparent'}">
								<path d="M20 6L9 17l-5-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
							</svg>
						{/snippet}
						<DropdownBase.Separator />
						<DropdownBase.Item onSelect={() => (viewMode = 'grid')}>
							{@render checkmark(viewMode === 'grid')}
							<svg viewBox="0 0 24 24" class="h-3.5 w-3.5 shrink-0 text-gray-500">
								<rect x="3" y="3" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.5" />
								<rect x="14" y="3" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.5" />
								<rect x="3" y="14" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.5" />
								<rect x="14" y="14" width="7" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.5" />
							</svg>
							<span class={viewMode === 'grid' ? 'font-medium text-gray-900' : ''}>{m.photos_view_grid()}</span>
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => (viewMode = 'list')}>
							{@render checkmark(viewMode === 'list')}
							<svg viewBox="0 0 24 24" class="h-3.5 w-3.5 shrink-0 text-gray-500">
								<line x1="8" y1="6" x2="21" y2="6" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
								<line x1="8" y1="12" x2="21" y2="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
								<line x1="8" y1="18" x2="21" y2="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
								<line x1="3" y1="6" x2="3.01" y2="6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
								<line x1="3" y1="12" x2="3.01" y2="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
								<line x1="3" y1="18" x2="3.01" y2="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
							</svg>
							<span class={viewMode === 'list' ? 'font-medium text-gray-900' : ''}>{m.photos_view_list()}</span>
						</DropdownBase.Item>
						<DropdownBase.Separator />
						<DropdownBase.Item onSelect={() => (groupByDate = !groupByDate)}>
							{@render checkmark(groupByDate)}
							<span class={groupByDate ? 'font-medium text-gray-900' : ''}>{m.photos_group_by_date()}</span>
						</DropdownBase.Item>
						<DropdownBase.Separator />
						<DropdownBase.Item onSelect={() => (photoSize = 'large')}>
							{@render checkmark(photoSize === 'large')}
							<span class={photoSize === 'large' ? 'font-medium text-gray-900' : ''}>{m.photos_size_large()}</span>
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => (photoSize = 'medium')}>
							{@render checkmark(photoSize === 'medium')}
							<span class={photoSize === 'medium' ? 'font-medium text-gray-900' : ''}>{m.photos_size_medium()}</span>
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => (photoSize = 'small')}>
							{@render checkmark(photoSize === 'small')}
							<span class={photoSize === 'small' ? 'font-medium text-gray-900' : ''}>{m.photos_size_small()}</span>
						</DropdownBase.Item>
					</Dropdown>
					<button
						type="button"
						onclick={() => (showCreateAlbum = true)}
						class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
					>
						<Plus size={15} />
						{m.albums_create()}
					</button>
				</div>
			</div>

			<!-- Loading -->
			{#if loading}
				<div class="flex items-center justify-center py-16">
					<LoaderCircle size={24} class="animate-spin text-gray-300" />
				</div>
			<!-- Empty -->
			{:else if photos.length === 0}
				<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
					<ImageIcon size={40} class="mb-3 text-gray-300" />
					<p class="text-sm text-gray-400">{m.photos_empty()}</p>
				</div>
			<!-- Content -->
			{:else}
				{#if groupByDate}
					{#each groupedPhotos as group (group.date)}
						<div class="mb-8">
							<div class="sticky top-14 z-30 bg-white/90 py-2 backdrop-blur-sm">
								<h2 class="text-base font-semibold text-gray-900">{group.label}</h2>
								<p class="text-xs text-gray-400">{m.photos_total({ total: group.items.length })}</p>
							</div>
							{#if viewMode === 'grid'}
								<div class="grid {gridClass}">
									{#each group.items as photo (photo.slug)}
										<button
											type="button"
											onclick={() => openViewer(photo.slug)}
											class="group relative aspect-square overflow-hidden rounded-md bg-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
										>
											<img
												src={authedUrl(thumbnailUrl(photo.slug))}
												alt={photo.fileName}
												class="h-full w-full object-cover transition group-hover:scale-105"
												loading="lazy"
											/>
											<div class="pointer-events-none absolute inset-0 bg-black/0 transition group-hover:bg-black/20"></div>
											{#if photo.isStarred}
												<div class="pointer-events-none absolute right-1.5 top-1.5 rounded-full bg-white/80 p-1">
													<svg viewBox="0 0 24 24" class="h-3 w-3 fill-amber-400 text-amber-400"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
												</div>
											{/if}
										</button>
									{/each}
								</div>
							{:else}
								<div class="overflow-hidden rounded-xl border border-gray-100 bg-white">
									<table class="w-full text-sm">
										<tbody>
											{#each group.items as photo (photo.slug)}
												<tr
													class="cursor-pointer border-t border-gray-50 transition-colors hover:bg-gray-50"
													onclick={() => openViewer(photo.slug)}
												>
													<td class="w-12 p-2">
														<img
															src={authedUrl(thumbnailUrl(photo.slug))}
															alt=""
															class="h-10 w-10 rounded object-cover"
															loading="lazy"
														/>
													</td>
													<td class="max-w-0 truncate px-2 py-2 text-gray-700">
														<span class="truncate">{photo.fileName}</span>
													</td>
													<td class="hidden whitespace-nowrap px-2 py-2 text-right text-gray-500 sm:table-cell">
														{photo.fileSize > 0 ? (photo.fileSize / 1024).toFixed(0) + ' KB' : '-'}
													</td>
													<td class="whitespace-nowrap px-2 py-2 text-right text-gray-500">
														{new Date(photo.createdAt).toLocaleDateString()}
													</td>
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							{/if}
						</div>
					{/each}
				{:else}
					{#if viewMode === 'grid'}
						<div class="grid {gridClass}">
							{#each photos as photo (photo.slug)}
								<button
									type="button"
									onclick={() => openViewer(photo.slug)}
									class="group relative aspect-square overflow-hidden rounded-md bg-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
								>
									<img
										src={authedUrl(thumbnailUrl(photo.slug))}
										alt={photo.fileName}
										class="h-full w-full object-cover transition group-hover:scale-105"
										loading="lazy"
									/>
									<div class="pointer-events-none absolute inset-0 bg-black/0 transition group-hover:bg-black/20"></div>
									{#if photo.isStarred}
										<div class="pointer-events-none absolute right-1.5 top-1.5 rounded-full bg-white/80 p-1">
											<svg viewBox="0 0 24 24" class="h-3 w-3 fill-amber-400 text-amber-400"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
										</div>
									{/if}
								</button>
							{/each}
						</div>
					{:else}
						<div class="overflow-hidden rounded-xl border border-gray-100 bg-white">
							<table class="w-full text-sm">
								<tbody>
									{#each photos as photo (photo.slug)}
										<tr
											class="cursor-pointer border-t border-gray-50 transition-colors hover:bg-gray-50"
											onclick={() => openViewer(photo.slug)}
										>
											<td class="w-12 p-2">
												<img
													src={authedUrl(thumbnailUrl(photo.slug))}
													alt=""
													class="h-10 w-10 rounded object-cover"
													loading="lazy"
												/>
											</td>
											<td class="max-w-0 truncate px-2 py-2 text-gray-700">
												<span class="truncate">{photo.fileName}</span>
											</td>
											<td class="hidden whitespace-nowrap px-2 py-2 text-right text-gray-500 sm:table-cell">
												{photo.fileSize > 0 ? (photo.fileSize / 1024).toFixed(0) + ' KB' : '-'}
											</td>
											<td class="whitespace-nowrap px-2 py-2 text-right text-gray-500">
												{new Date(photo.createdAt).toLocaleDateString()}
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				{/if}

				{#if hasMore}
					<div class="flex justify-center pb-8">
						<button
							type="button"
							onclick={loadMore}
							disabled={loadingMore}
							class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-6 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{#if loadingMore}
								<LoaderCircle size={15} class="animate-spin" />
							{/if}
							{m.photos_load_more()}
						</button>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Lightbox -->
	{#if viewerSlug}
		<PhotoViewer
			slug={viewerSlug}
			bind:fileSlugs={allSlugs}
			index={viewerIndex}
			close={() => (viewerSlug = null)}
			{photos}
		/>
	{/if}

	<!-- Create Album dialog -->
	<AlbumCreateDialog bind:show={showCreateAlbum} onCreated={() => {}} />
{/if}
