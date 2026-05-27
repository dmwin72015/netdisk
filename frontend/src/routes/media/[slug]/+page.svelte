<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import Hls from 'hls.js';
	import { ArrowLeft, Loader2, AlertCircle, Clock } from '@lucide/svelte';
	import { getMediaItem, getHLSUrl, type MediaItem } from '$lib/api/media';
	import { getAccessToken } from '$lib/api/client';
	import * as m from '$lib/paraglide/messages';

	let item = $state<MediaItem | null>(null);
	let error = $state<string | null>(null);
	let video: HTMLVideoElement | undefined = $state();
	let hls: Hls | null = null;

	function authedHLS(url: string): string {
		const token = getAccessToken();
		if (!token) return url;
		const u = new URL(url, window.location.origin);
		u.searchParams.set('access_token', token);
		return u.pathname + '?' + u.searchParams.toString();
	}

	async function load() {
		try {
			item = await getMediaItem(page.params.slug!);
			if (item.status === 'done') {
				const masterUrl = getHLSUrl(item.media_slug, 'master.m3u8');
				attachHLS(authedHLS(masterUrl));
			}
		} catch (err) {
			error = err instanceof Error ? err.message : m.load_failed();
		}
	}

	function attachHLS(src: string) {
		if (!video) return;
		if (Hls.isSupported()) {
			hls = new Hls({
				xhrSetup: (xhr) => {
					const token = getAccessToken();
					if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
				}
			});
			hls.on(Hls.Events.ERROR, (_event, data) => {
				console.error('hls error', data);
				if (data.fatal) error = m.player_error({ details: data.details });
			});
			hls.loadSource(src);
			hls.attachMedia(video);
		} else if (video.canPlayType('application/vnd.apple.mpegurl')) {
			video.src = src;
		} else {
			error = m.hls_not_supported();
		}
	}

	// Poll for status updates while processing
	let pollTimer: ReturnType<typeof setInterval> | undefined;
	$effect(() => {
		if (item && (item.status === 'pending' || item.status === 'processing')) {
			pollTimer = setInterval(async () => {
				try {
					const updated = await getMediaItem(page.params.slug!);
					item = updated;
					if (updated.status === 'done') {
						if (pollTimer) clearInterval(pollTimer);
						const masterUrl = getHLSUrl(updated.media_slug, 'master.m3u8');
						attachHLS(authedHLS(masterUrl));
					} else if (updated.status === 'failed') {
						if (pollTimer) clearInterval(pollTimer);
						error = updated.error_msg || m.conversion_failed();
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

	function fmtDuration(sec: number | null): string {
		if (!sec || sec <= 0) return '-';
		const s = Math.round(sec);
		const h = Math.floor(s / 3600);
		const m = Math.floor((s % 3600) / 60);
		const r = s % 60;
		if (h > 0) return `${h}h ${m}m ${r}s`;
		if (m > 0) return `${m}m ${r}s`;
		return `${r}s`;
	}

	onMount(load);
	onDestroy(() => {
		hls?.destroy();
		if (pollTimer) clearInterval(pollTimer);
	});
</script>

<div class="space-y-4">
	<button type="button" onclick={() => goto('/media')} class="flex items-center gap-1.5 text-sm text-gray-500 transition-colors hover:text-gray-700">
		<ArrowLeft size={16} /> {m.back_to_media()}
	</button>

	{#if error}
		<div class="flex items-start gap-2.5 rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">
			<AlertCircle size={16} class="mt-0.5 shrink-0" />
			<span>{error}</span>
		</div>
	{:else if !item}
		<div class="flex items-center justify-center py-16">
			<Loader2 size={24} class="animate-spin text-gray-300" />
		</div>
	{:else}
		<h1 class="text-xl font-semibold text-gray-900">{item.file_name}</h1>

		{#if item.status === 'done'}
			<div class="overflow-hidden rounded-xl bg-black">
				<video bind:this={video} controls class="w-full aspect-video"></video>
			</div>
		{:else if item.status === 'processing'}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-blue-200 bg-blue-50 py-20">
				<Loader2 size={32} class="animate-spin text-blue-400" />
				<p class="mt-3 text-sm font-medium text-blue-600">{m.converting()} {item.progress}%</p>
				<div class="mt-3 h-2 w-48 overflow-hidden rounded-full bg-blue-100">
					<div class="h-full rounded-full bg-blue-500 transition-all" style="width:{item.progress}%"></div>
				</div>
			</div>
		{:else if item.status === 'pending'}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-gray-200 bg-gray-50 py-20">
				<Clock size={32} class="text-gray-300" />
				<p class="mt-3 text-sm text-gray-500">{m.in_queue()}</p>
			</div>
		{:else if item.status === 'failed'}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-red-200 bg-red-50 py-20">
				<AlertCircle size={32} class="text-red-300" />
				<p class="mt-3 text-sm text-red-600">{item.error_msg || m.conversion_failed()}</p>
			</div>
		{/if}

		<dl class="grid grid-cols-2 gap-3 rounded-xl border border-gray-100 bg-white p-4 text-sm sm:grid-cols-4">
			<div>
				<dt class="text-gray-400">{m.status()}</dt>
				<dd class="mt-0.5 font-medium capitalize">{item.status}</dd>
			</div>
			{#if item.duration_sec}
				<div>
					<dt class="text-gray-400">{m.duration()}</dt>
					<dd class="mt-0.5 font-medium">{fmtDuration(item.duration_sec)}</dd>
				</div>
			{/if}
			<div>
				<dt class="text-gray-400">{m.created()}</dt>
				<dd class="mt-0.5">{new Date(item.created_at).toLocaleString()}</dd>
			</div>
			<div>
				<dt class="text-gray-400">{m.media_id()}</dt>
				<dd class="mt-0.5 font-mono text-xs">{item.media_slug}</dd>
			</div>
		</dl>
	{/if}
</div>
