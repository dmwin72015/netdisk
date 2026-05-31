<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Hls from 'hls.js';
	import { ArrowLeft, Loader2, AlertCircle, Clock, Info } from '@lucide/svelte';
	import { getMediaItem, type MediaItem } from '$lib/api/media';
	import { getAccessToken } from '$lib/api/client';
	import { authedUrl } from '$lib/utils/format';
	import VideoStats from '$lib/components/media/VideoStats.svelte';
	import * as m from '$lib/paraglide/messages';

	let item = $state<MediaItem | null>(null);
	let error = $state<string | null>(null);
	let video: HTMLVideoElement | undefined = $state();
	let hls: Hls | null = null;
	let hlsAttached = false;
	let showStats = $state(false);

	async function load() {
		try {
			item = await getMediaItem(page.params.slug!);
		} catch (err) {
			error = err instanceof Error ? err.message : m.load_failed();
		}
	}

	function attachHLS(src: string) {
		if (!video || hlsAttached) return;
		hlsAttached = true;
		if (Hls.isSupported()) {
			hls = new Hls({
				xhrSetup: (xhr) => {
					const token = getAccessToken();
					if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
				}
			});
			let bufferRetries = 0;
			hls.on(Hls.Events.ERROR, (_event, data) => {
				if (data.fatal) {
					switch (data.type) {
						case Hls.ErrorTypes.NETWORK_ERROR:
							hls?.startLoad();
							break;
						case Hls.ErrorTypes.MEDIA_ERROR:
							hls?.recoverMediaError();
							break;
						default:
							error = m.player_error({ details: data.details });
							break;
					}
				} else if (data.details === Hls.ErrorDetails.BUFFER_APPENDING_ERROR) {
					bufferRetries++;
					if (bufferRetries > 3) {
						hls?.destroy();
						error = m.player_error({ details: data.details });
					}
				}
			});
			hls.loadSource(src);
			hls.attachMedia(video);
		} else if (video.canPlayType('application/vnd.apple.mpegurl')) {
			video.src = src;
		} else {
			error = m.hls_not_supported();
		}
	}

	// Attach HLS once video element is bound and item is loaded
	$effect(() => {
		if (video && item?.status === 'done' && item.playUrl && !hlsAttached) {
			attachHLS(authedUrl(item.playUrl));
		}
	});

	// Poll for status updates while processing
	let pollTimer: ReturnType<typeof setInterval> | undefined;
	$effect(() => {
		if (item && (item.status === 'pending' || item.status === 'processing')) {
			pollTimer = setInterval(async () => {
				try {
					const updated = await getMediaItem(page.params.slug!);
					item = updated;
					if (updated.status === 'done' || updated.status === 'failed') {
						if (pollTimer) clearInterval(pollTimer);
						if (updated.status === 'failed') {
							error = updated.errorMsg || m.conversion_failed();
						}
					}
				} catch {
					// ignore poll errors
				}
			}, 3000);
			return () => {
				if (pollTimer) clearInterval(pollTimer);
			};
		}
	});

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 's' && !e.ctrlKey && !e.metaKey && !e.altKey) {
			const tag = (e.target as HTMLElement)?.tagName;
			if (tag === 'INPUT' || tag === 'TEXTAREA') return;
			showStats = !showStats;
		}
	}

	onMount(load);
	onDestroy(() => {
		hls?.destroy();
		if (pollTimer) clearInterval(pollTimer);
	});
</script>

<svelte:window onkeydown={onKeydown} />

<div class="px-4">
	<!-- Back button -->
	<button type="button" onclick={() => goto('/media')}
		class="mb-4 flex items-center gap-1.5 text-sm text-gray-500 transition-colors hover:text-gray-700">
		<ArrowLeft size={16} /> {m.back_to_media()}
	</button>

	{#if error}
		<div class="flex items-start gap-2.5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">
			<AlertCircle size={16} class="mt-0.5 shrink-0" />
			<span>{error}</span>
		</div>
	{:else if !item}
		<div class="flex items-center justify-center py-32">
			<Loader2 size={28} class="animate-spin text-gray-300" />
		</div>
	{:else}
		<!-- Video player -->
		{#if item.status === 'done'}
			<div class="group relative overflow-hidden rounded-2xl bg-black">
				<video bind:this={video} controls class="w-full aspect-video"></video>
				<VideoStats {hls} {video} bind:visible={showStats} />
				<button type="button" onclick={() => showStats = !showStats}
					class="absolute bottom-3 right-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-black/50 text-white opacity-0 backdrop-blur-sm transition-opacity hover:bg-black/70 group-hover:opacity-100"
					title="Stats (S)">
					<Info size={16} />
				</button>
			</div>
		{:else if item.status === 'processing'}
			<div class="flex flex-col items-center justify-center rounded-2xl bg-gray-900 py-24">
				<Loader2 size={36} class="animate-spin text-gray-400" />
				<p class="mt-4 text-sm font-medium text-gray-300">{m.converting()} {item.progress}%</p>
				<div class="mt-3 h-1.5 w-64 overflow-hidden rounded-full bg-gray-700">
					<div class="h-full rounded-full bg-blue-500 transition-all" style="width:{item.progress}%"></div>
				</div>
			</div>
		{:else if item.status === 'pending'}
			<div class="flex flex-col items-center justify-center rounded-2xl bg-gray-900 py-24">
				<Clock size={36} class="text-gray-500" />
				<p class="mt-4 text-sm text-gray-400">{m.in_queue()}</p>
			</div>
		{:else if item.status === 'failed'}
			<div class="flex flex-col items-center justify-center rounded-2xl bg-red-950 py-24">
				<AlertCircle size={36} class="text-red-400" />
				<p class="mt-4 text-sm text-red-300">{item.errorMsg || m.conversion_failed()}</p>
			</div>
		{/if}

		<!-- Video title -->
		<h1 class="mt-4 text-lg font-semibold text-gray-900">{item.fileName}</h1>
	{/if}
</div>
