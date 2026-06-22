<script lang="ts">
	import { authedUrl, copyToClipboard } from '$lib/utils/format';
	import {
		ChevronLeft,
		ChevronRight,
		X,
		Image as ImageIcon,
		Star,
		Download,
		Link,
		ZoomIn,
		ZoomOut,
		RotateCcw,
		RotateCw,
		RefreshCcw,
		FolderOpen
	} from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { downloadUrl, setStarred } from '$lib/api/files';
	import { toast } from 'svelte-sonner';
	import type { PhotoItem } from '$lib/api/photos';
	import * as m from '$lib/paraglide/messages';

	const MIN_SCALE = 0.25;
	const MAX_SCALE = 5;
	const SCALE_STEP = 0.25;
	const WHEEL_ZOOM_SENSITIVITY = 0.0015;

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

	// svelte-ignore state_referenced_locally
	let currentIndex = $state(index);
	let currentSlug = $derived(fileSlugs[currentIndex] ?? slug);
	let visibleSlug = $state<string | null>(null);
	let loading = $state(true);
	let isSwitchingImage = $state(false);
	let closing = $state(false);
	let scale = $state(1);
	let rotation = $state(0);
	let offsetX = $state(0);
	let offsetY = $state(0);
	let isPanning = $state(false);
	let isWheelZooming = $state(false);
	let imageFrame: HTMLDivElement | undefined = $state();
	let wheelZoomTimer: ReturnType<typeof setTimeout> | undefined;
	let panStartX = 0;
	let panStartY = 0;
	let panOriginX = 0;
	let panOriginY = 0;
	let imageTransform = $derived(`scale(${scale}) rotate(${rotation}deg)`);
	let panTransform = $derived(`translate3d(${offsetX}px, ${offsetY}px, 0)`);
	let zoomLabel = $derived(`${Math.round(scale * 100)}%`);

	let currentPhoto = $derived(photos.find(p => p.slug === currentSlug));
	let visiblePhoto = $derived(photos.find(p => p.slug === visibleSlug));
	let requestedSlug: string | null = null;
	let imageLoadToken = 0;

	$effect(() => {
		currentIndex = index;
	});

	$effect(() => {
		const nextSlug = currentSlug;
		if (!nextSlug || nextSlug === requestedSlug) return;

		requestedSlug = nextSlug;
		loading = true;
		isSwitchingImage = true;
		resetTransform();

		const token = ++imageLoadToken;
		const image = new Image();
		let cancelled = false;

		const revealImage = async () => {
			try {
				await image.decode();
			} catch {
				// Some browsers reject decode() for cached or partially decoded images.
			}

			if (cancelled || token !== imageLoadToken) return;
			visibleSlug = nextSlug;
			loading = false;
			requestAnimationFrame(() => {
				if (!cancelled && token === imageLoadToken) {
					isSwitchingImage = false;
				}
			});
		};

		image.onload = () => {
			void revealImage();
		};
		image.onerror = () => {
			if (cancelled || token !== imageLoadToken) return;
			visibleSlug = nextSlug;
			loading = false;
			isSwitchingImage = false;
		};
		image.decoding = 'async';
		image.src = authedUrl(downloadUrl(nextSlug));
		if (image.complete) {
			void revealImage();
		}

		return () => {
			cancelled = true;
			image.onload = null;
			image.onerror = null;
		};
	});

	function clampScale(value: number) {
		return Math.min(MAX_SCALE, Math.max(MIN_SCALE, value));
	}

	function resetTransform() {
		scale = 1;
		rotation = 0;
		offsetX = 0;
		offsetY = 0;
		isPanning = false;
		isWheelZooming = false;
		clearTimeout(wheelZoomTimer);
	}

	function setScale(nextScale: number, focalPoint?: { x: number; y: number }) {
		const previousScale = scale;
		const next = clampScale(nextScale);
		if (next === previousScale) return;

		if (focalPoint && imageFrame && next > 1) {
			const rect = imageFrame.getBoundingClientRect();
			const centerX = rect.left + rect.width / 2;
			const centerY = rect.top + rect.height / 2;
			const ratio = next / previousScale;
			offsetX += (1 - ratio) * (focalPoint.x - centerX);
			offsetY += (1 - ratio) * (focalPoint.y - centerY);
		}

		scale = next;
		if (scale <= 1) {
			offsetX = 0;
			offsetY = 0;
		}
	}

	function setCurrentIndex(nextIndex: number) {
		currentIndex = nextIndex;
		loading = true;
		resetTransform();
	}

	function prev() {
		if (currentIndex > 0) {
			setCurrentIndex(currentIndex - 1);
		}
	}

	function next() {
		if (currentIndex < fileSlugs.length - 1) {
			setCurrentIndex(currentIndex + 1);
		}
	}

	function zoomIn(e?: Event) {
		e?.stopPropagation();
		setScale(scale + SCALE_STEP);
	}

	function zoomOut(e?: Event) {
		e?.stopPropagation();
		setScale(scale - SCALE_STEP);
	}

	function rotateLeft(e?: Event) {
		e?.stopPropagation();
		rotation -= 90;
	}

	function rotateRight(e?: Event) {
		e?.stopPropagation();
		rotation += 90;
	}

	function resetView(e?: Event) {
		e?.stopPropagation();
		resetTransform();
	}

	function handleWheel(e: WheelEvent) {
		e.preventDefault();
		e.stopPropagation();
		isWheelZooming = true;
		clearTimeout(wheelZoomTimer);
		wheelZoomTimer = setTimeout(() => {
			isWheelZooming = false;
		}, 120);

		const nextScale = scale * Math.exp(-e.deltaY * WHEEL_ZOOM_SENSITIVITY);
		setScale(nextScale, { x: e.clientX, y: e.clientY });
	}

	function startPan(e: PointerEvent) {
		e.stopPropagation();
		if (scale <= 1) return;
		isPanning = true;
		panStartX = e.clientX;
		panStartY = e.clientY;
		panOriginX = offsetX;
		panOriginY = offsetY;
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
	}

	function movePan(e: PointerEvent) {
		if (!isPanning) return;
		e.stopPropagation();
		offsetX = panOriginX + e.clientX - panStartX;
		offsetY = panOriginY + e.clientY - panStartY;
	}

	function endPan(e: PointerEvent) {
		if (!isPanning) return;
		e.stopPropagation();
		isPanning = false;
		const target = e.currentTarget as HTMLElement;
		if (target.hasPointerCapture(e.pointerId)) {
			target.releasePointerCapture(e.pointerId);
		}
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
		if (e.key === 'ArrowLeft') prev();
		if (e.key === 'ArrowRight') next();
		if (e.key === '+' || e.key === '=') zoomIn();
		if (e.key === '-') zoomOut();
		if (e.key === '0') resetView();
		if (e.key.toLowerCase() === 'r') rotateRight();
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

	function handleOpenFolder(e: Event) {
		e.stopPropagation();
		const parentSlug = currentPhoto?.parentSlug;
		if (parentSlug) {
			void goto('/files/all/' + parentSlug);
		} else {
			void goto('/files/all');
		}
		close();
	}

	function closeViewer() {
		if (closing) return;
		closing = true;
	}

	function onOverlayEnd() {
		if (closing) {
			close();
		}
	}

	function stopPropagation(e: Event) {
		e.stopPropagation();
	}
</script>

<svelte:window onkeydown={onKeydown} />

<!-- Overlay -->
{#if currentSlug}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		role="dialog"
		aria-modal="true"
		tabindex="-1"
		class="photo-viewer-overlay fixed inset-0 z-50 flex items-center justify-center bg-black/90"
		class:photo-viewer-overlay-closing={closing}
		onclick={closeViewer}
		onanimationend={onOverlayEnd}
	>
		<!-- Top-right actions -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div role="toolbar" aria-label={m.preview()} tabindex="-1" class="photo-viewer-actions absolute right-4 top-4 z-10 flex items-center gap-2" onclick={stopPropagation}>
			<button
				type="button"
				onclick={toggleStar}
				class="rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<Star size={20} class={currentPhoto?.isStarred ? 'fill-star text-warning' : ''} />
			</button>
			<button
				type="button"
				onclick={handleOpenFolder}
				title="打开所在目录"
				class="rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<FolderOpen size={20} />
			</button>
			<button
				type="button"
				onclick={handleDownload}
				class="rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<Download size={20} />
			</button>
			<button
				type="button"
				onclick={handleCopyLink}
				class="rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<Link size={20} />
			</button>
			<button
				type="button"
				onclick={closeViewer}
				class="rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<X size={24} />
			</button>
		</div>

		<!-- Prev -->
		{#if currentIndex > 0}
			<button
				type="button"
				onclick={(e) => { e.stopPropagation(); prev(); }}
				class="absolute left-4 top-1/2 z-10 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<ChevronLeft size={28} />
			</button>
		{/if}

		<!-- Next -->
		{#if currentIndex < fileSlugs.length - 1}
			<button
				type="button"
				onclick={(e) => { e.stopPropagation(); next(); }}
				class="absolute right-4 top-1/2 z-10 -translate-y-1/2 rounded-full bg-black/50 p-2 text-white transition-all duration-150 ease-out hover:scale-110 hover:bg-black/70 active:scale-95"
			>
				<ChevronRight size={28} />
			</button>
			{/if}

			<!-- Image -->
			<div class="relative flex h-full w-full items-center justify-center overflow-hidden p-4" onwheel={handleWheel}>
				{#if loading}
					<div class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center">
						<ImageIcon size={40} class="animate-pulse text-gray-500" />
					</div>
				{/if}
				{#if visibleSlug}
					<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
					<div
						bind:this={imageFrame}
						role="presentation"
						class="photo-viewer-image-wrap inline-flex touch-none select-none will-change-transform {scale > 1 ? (isPanning ? 'cursor-grabbing' : 'cursor-grab') : 'cursor-default'}"
						style:transform={panTransform}
						onclick={stopPropagation}
						onpointerdown={startPan}
						onpointermove={movePan}
						onpointerup={endPan}
						onpointercancel={endPan}
					>
						{#key visibleSlug}
							<img
								src={authedUrl(downloadUrl(visibleSlug))}
								alt={visiblePhoto?.fileName ?? ''}
								loading="eager"
								class="photo-viewer-image max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-dialog will-change-transform {isWheelZooming || isPanning || isSwitchingImage ? '' : 'transition-transform duration-150 ease-out'}"
								style:transform={imageTransform}
								draggable="false"
							/>
						{/key}
					</div>
				{/if}
			</div>

		<!-- Toolbar -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div role="toolbar" aria-label={m.preview()} tabindex="-1" class="photo-viewer-toolbar absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 items-center gap-2 rounded-full bg-black/50 px-3 py-2 text-sm text-white shadow-pop backdrop-blur" onclick={stopPropagation}>
			<span class="min-w-14 px-2 text-center text-white/80">{currentIndex + 1} / {fileSlugs.length}</span>
			<div class="h-5 w-px bg-white/20"></div>
			<button type="button" onclick={zoomOut} class="rounded-full p-1.5 transition-colors hover:bg-white/15 disabled:opacity-40" disabled={scale <= MIN_SCALE} title={m.zoom_out()} aria-label={m.zoom_out()}>
				<ZoomOut size={18} />
			</button>
			<span class="min-w-12 text-center text-xs text-white/80">{zoomLabel}</span>
			<button type="button" onclick={zoomIn} class="rounded-full p-1.5 transition-colors hover:bg-white/15 disabled:opacity-40" disabled={scale >= MAX_SCALE} title={m.zoom_in()} aria-label={m.zoom_in()}>
				<ZoomIn size={18} />
			</button>
			<div class="h-5 w-px bg-white/20"></div>
			<button type="button" onclick={rotateLeft} class="rounded-full p-1.5 transition-colors hover:bg-white/15" title={m.rotate_left()} aria-label={m.rotate_left()}>
				<RotateCcw size={18} />
			</button>
			<button type="button" onclick={rotateRight} class="rounded-full p-1.5 transition-colors hover:bg-white/15" title={m.rotate_right()} aria-label={m.rotate_right()}>
				<RotateCw size={18} />
			</button>
			<button type="button" onclick={resetView} class="rounded-full p-1.5 transition-colors hover:bg-white/15" title={m.reset_view()} aria-label={m.reset_view()}>
				<RefreshCcw size={18} />
			</button>
		</div>
	</div>
{/if}
