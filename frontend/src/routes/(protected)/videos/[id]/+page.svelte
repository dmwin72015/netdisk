<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import Hls from 'hls.js';
	import { Copy, Image as ImageIcon, Crosshair, Upload, X } from '@lucide/svelte';
	import { getVideo, uploadThumbnail, captureFrameThumbnail } from '$lib/api/videos';
	import { ApiError, getAccessToken } from '$lib/api/client';
	import type { Task } from '$lib/api/tasks';
	import { fmtDurationText, authedUrl, copyToClipboard, clipboardUnavailableReason } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let task = $state<Task | null>(null);
	let error = $state<string | null>(null);
	let copied = $state(false);
	let copyError = $state<string | null>(null);
	let video: HTMLVideoElement | undefined = $state();
	let hls: Hls | null = null;
	let thumbInput: HTMLInputElement | undefined = $state();
	let thumbBusy = $state<null | 'upload' | 'frame'>(null);
	let thumbMsg = $state<string | null>(null);
	let thumbIsError = $state(false);

	async function load() {
		try {
			task = await getVideo(page.params.id!);
		} catch (err) {
			error = err instanceof Error ? err.message : m.load_failed();
		}
	}

	$effect(() => {
		if (!task?.m3u8Url || !video || hls) return;
		attach(authedUrl(task.m3u8Url));
	});

	function attach(src: string) {
		if (!video) return;
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

	async function copyUrl() {
		if (!task?.m3u8Url) return;
		const ok = await copyToClipboard(task.m3u8Url);
		if (ok) {
			copied = true;
			copyError = null;
			setTimeout(() => (copied = false), 1500);
		} else {
			copyError = clipboardUnavailableReason();
			setTimeout(() => (copyError = null), 3000);
		}
	}

	function pickThumbnailFile() {
		thumbMsg = null;
		thumbInput?.click();
	}

	async function onThumbPicked(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		target.value = '';
		if (!file || !task) return;
		thumbBusy = 'upload';
		thumbMsg = null;
		thumbIsError = false;
		try {
			task = await uploadThumbnail(task.id, file);
			thumbMsg = m.cover_updated();
		} catch (err) {
			thumbMsg = err instanceof ApiError ? err.message : err instanceof Error ? err.message : m.upload_failed();
			thumbIsError = true;
		} finally {
			thumbBusy = null;
		}
	}

	async function captureCurrentFrame() {
		if (!task || !video) return;
		const at = video.currentTime;
		thumbBusy = 'frame';
		thumbMsg = null;
		thumbIsError = false;
		try {
			task = await captureFrameThumbnail(task.id, at);
			thumbMsg = m.used_frame_at({ at: at.toFixed(1) });
		} catch (err) {
			thumbMsg = err instanceof ApiError ? err.message : err instanceof Error ? err.message : m.capture_failed();
			thumbIsError = true;
			thumbMsg = err instanceof ApiError ? err.message : err instanceof Error ? err.message : m.capture_failed();
		} finally {
			thumbBusy = null;
		}
	}

	onMount(load);
	onDestroy(() => hls?.destroy());
</script>

{#if error}
	<p class="text-sm text-red-600">{error}</p>
{:else if !task}
	<p class="text-sm text-slate-500">{m.loading()}</p>
{:else}
	<h1 class="text-xl font-semibold">{task.originalName}</h1>
	<div class="mt-4 overflow-hidden rounded-lg bg-black">
		<video bind:this={video} controls class="w-full aspect-video"></video>
	</div>

	<dl class="mt-4 grid grid-cols-2 gap-2 text-sm sm:grid-cols-4">
		<div><dt class="text-slate-500">{m.detail_size()}</dt><dd>{(task.fileSize / 1024 / 1024).toFixed(1)} MB</dd></div>
		{#if task.durationSec}
			<div><dt class="text-slate-500">{m.duration()}</dt><dd>{fmtDurationText(task.durationSec)}</dd></div>
		{/if}
		<div><dt class="text-slate-500">{m.status()}</dt><dd>{task.status}</dd></div>
		<div><dt class="text-slate-500">{m.created()}</dt><dd>{new Date(task.createdAt * 1000).toLocaleString()}</dd></div>
	</dl>

	{#if task.m3u8Url}
		<div class="mt-4 flex items-center gap-2">
			<code class="flex-1 truncate rounded bg-slate-100 px-2 py-1 text-xs">{task.m3u8Url}</code>
			<button
				type="button"
				onclick={copyUrl}
				class="flex items-center gap-1 rounded border px-2 py-1 text-xs hover:bg-slate-50"
			>
				<Copy size={12} /> {copied ? m.copied() : m.copy_m3u8()}
			</button>
			{#if copyError}
				<span class="text-xs text-red-500">{copyError}</span>
			{/if}
		</div>
	{/if}

	<!-- Cover editor -->
	<section class="mt-6 rounded-lg border bg-white p-4">
		<header class="flex items-center gap-2">
			<ImageIcon size={16} class="text-slate-500" />
			<h2 class="text-sm font-medium">{m.cover_image()}</h2>
		</header>
		<div class="mt-3 flex flex-col gap-4 sm:flex-row">
			<div class="aspect-video w-full max-w-xs overflow-hidden rounded bg-slate-200 sm:w-64">
				{#if task.thumbnailUrl}
						<img
							src={authedUrl(task.thumbnailUrl)}
							alt={m.cover_preview()}
							loading="lazy"
							class="h-full w-full object-cover"
					/>
				{:else}
					<div class="flex h-full w-full items-center justify-center text-slate-400">
						<ImageIcon size={28} />
					</div>
				{/if}
			</div>
			<div class="flex flex-1 flex-col gap-2 text-sm">
				<button
					type="button"
					onclick={pickThumbnailFile}
					disabled={thumbBusy !== null}
					class="inline-flex items-center justify-center gap-2 rounded border border-slate-300 bg-white px-3 py-2 hover:bg-slate-50 disabled:opacity-60"
				>
					<Upload size={14} />
					{thumbBusy === 'upload' ? m.uploading() : m.upload_image()}
				</button>
				<button
					type="button"
					onclick={captureCurrentFrame}
					disabled={thumbBusy !== null}
					class="inline-flex items-center justify-center gap-2 rounded border border-slate-300 bg-white px-3 py-2 hover:bg-slate-50 disabled:opacity-60"
				>
					<Crosshair size={14} />
					{thumbBusy === 'frame' ? m.capturing() : m.use_current_frame()}
				</button>
				<p class="text-xs text-slate-500">{m.cover_hint()}</p>
				{#if thumbMsg}
					<p class="flex items-center gap-1 text-xs {thumbIsError ? 'text-red-600' : 'text-emerald-700'}">
						{thumbMsg}
						<button type="button" onclick={() => (thumbMsg = null)} aria-label={m.close()} class="text-slate-400 hover:text-slate-700">
							<X size={12} />
						</button>
					</p>
				{/if}
			</div>
		</div>
		<input
			bind:this={thumbInput}
			type="file"
			accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp"
			class="hidden"
			onchange={onThumbPicked}
		/>
	</section>
{/if}
