<script lang="ts">
	import { authedUrl, copyToClipboard } from '$lib/utils/format';
	import { ChevronLeft, ChevronRight, X, Image as ImageIcon, Star, Download, Link } from '@lucide/svelte';
	import { downloadUrl, setStarred } from '$lib/api/files';
	import { toast } from 'svelte-sonner';
	import type { PhotoItem } from '$lib/api/photos';
	import * as m from '$lib/paraglide/messages';

	let {
		slug,
		fileSlugs = $bindable([]),
		index,
		close,
		photos = [],
	}: {
		slug: string | null;
		fileSlugs: string[];
		index: number;
		close: () => void;
		photos: PhotoItem[];
	} = $props();

	let currentIndex = $state(index);
	let currentSlug = $derived(fileSlugs[currentIndex] ?? slug);
	let loading = $state(true);

	let currentPhoto = $derived(photos.find(p => p.slug === currentSlug));

	$effect(() => {
		currentIndex = index;
	});

	function prev() {
		if (currentIndex > 0) {
			currentIndex--;
			loading = true;
		}
	}

	function next() {
		if (currentIndex < fileSlugs.length - 1) {
			currentIndex++;
			loading = true;
		}
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
		if (e.key === 'ArrowLeft') prev();
		if (e.key === 'ArrowRight') next();
	}

	async function toggleStar(e: Event) {
		e.stopPropagation();
		if (!currentPhoto) return;
		const next = !currentPhoto.isStarred;
		currentPhoto.isStarred = next;
		try {
			await setStarred(currentSlug!, next);
		} catch {
			currentPhoto.isStarred = !next;
			toast.error(m.action_failed());
		}
	}

	function handleDownload(e: Event) {
		e.stopPropagation();
		const a = document.createElement('a');
		a.href = authedUrl(downloadUrl(currentSlug!));
		a.download = currentPhoto?.fileName || '';
		a.click();
	}

	async function handleCopyLink(e: Event) {
		e.stopPropagation();
		const url = new URL(authedUrl(downloadUrl(currentSlug!)), window.location.origin);
		if (await copyToClipboard(url.toString())) {
			toast.success(m.link_copied());
		} else {
			toast.error(m.copy_failed());
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<!-- Overlay -->
{#if currentSlug}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
		onclick={close}
	>
		<!-- Top-right actions -->
		<div class="absolute right-4 top-4 z-10 flex items-center gap-2">
			<button
				type="button"
				onclick={toggleStar}
				class="rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<Star size={20} class={currentPhoto?.isStarred ? 'fill-amber-400 text-amber-400' : ''} />
			</button>
			<button
				type="button"
				onclick={handleDownload}
				class="rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<Download size={20} />
			</button>
			<button
				type="button"
				onclick={handleCopyLink}
				class="rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<Link size={20} />
			</button>
			<button
				type="button"
				onclick={close}
				class="rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<X size={24} />
			</button>
		</div>

		<!-- Prev -->
		{#if currentIndex > 0}
			<button
				type="button"
				onclick={(e) => { e.stopPropagation(); prev(); }}
				class="absolute left-4 top-1/2 z-10 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<ChevronLeft size={28} />
			</button>
		{/if}

		<!-- Next -->
		{#if currentIndex < fileSlugs.length - 1}
			<button
				type="button"
				onclick={(e) => { e.stopPropagation(); next(); }}
				class="absolute right-4 top-1/2 z-10 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
			>
				<ChevronRight size={28} />
			</button>
		{/if}

		<!-- Image -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="flex max-h-full max-w-full items-center justify-center p-4"
			onclick={(e) => e.stopPropagation()}
		>
			{#if loading}
				<div class="flex items-center justify-center p-16">
					<ImageIcon size={40} class="animate-pulse text-gray-500" />
				</div>
			{/if}
			<img
				src={authedUrl(downloadUrl(currentSlug))}
				alt=""
				class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-2xl"
				class:hidden={loading}
				onload={() => (loading = false)}
			/>
		</div>

		<!-- Toolbar -->
		<div class="absolute bottom-4 left-1/2 z-10 -translate-x-1/2 rounded-full bg-black/50 px-4 py-2 text-white text-sm">
			{currentIndex + 1} / {fileSlugs.length}
		</div>
	</div>
{/if}
