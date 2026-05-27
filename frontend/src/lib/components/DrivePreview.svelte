<script lang="ts">
	import { X, Download } from '@lucide/svelte';
	import { downloadUrl } from '$lib/api/files';
	import { getAccessToken } from '$lib/api/client';
	import * as m from '$lib/paraglide/messages';

	let {
		id,
		name,
		mimeType,
		size,
		open,
		close
	}: {
		id: string;
		name: string;
		mimeType: string;
		size: number;
		open: boolean;
		close: () => void;
	} = $props();

	let closing = $state(false);
	let contentEl = $state<HTMLDivElement | undefined>();

	let textContent = $state<string | null>(null);
	let textError = $state<string | null>(null);
	let loadingText = $state(false);

	let blobUrl = $state<string | null>(null);
	let blobError = $state<string | null>(null);
	let loadingBlob = $state(false);

	let dlUrl = $derived(downloadUrl(id));
	let canPreview = $derived(
		mimeType.startsWith('image/') ||
			mimeType.startsWith('video/') ||
			mimeType.startsWith('audio/') ||
			mimeType === 'application/pdf' ||
			mimeType.startsWith('text/')
	);
	let isText = $derived(mimeType.startsWith('text/'));
	let isImage = $derived(mimeType.startsWith('image/'));
	let isVideo = $derived(mimeType.startsWith('video/'));
	let isAudio = $derived(mimeType.startsWith('audio/'));
	let isPdf = $derived(mimeType === 'application/pdf');

	$effect(() => {
		if (!open || !isText) return;
		textContent = null;
		textError = null;
		loadingText = true;
		const token = getAccessToken() ?? '';
		fetch(dlUrl, { headers: { Authorization: `Bearer ${token}` } })
			.then((r) => {
				if (!r.ok) throw new Error('Failed to load');
				return r.text();
			})
			.then((t) => (textContent = t))
			.catch((e) => (textError = e.message))
			.finally(() => (loadingText = false));
	});

	$effect(() => {
		if (!open || !(isImage || isVideo || isAudio || isPdf)) return;
		blobUrl = null;
		blobError = null;
		loadingBlob = true;
		const token = getAccessToken() ?? '';
		fetch(dlUrl, { headers: { Authorization: `Bearer ${token}` } })
			.then((r) => {
				if (!r.ok) throw new Error('Failed to load');
				return r.blob();
			})
			.then((blob) => {
				blobUrl = URL.createObjectURL(blob);
			})
			.catch((e) => (blobError = e.message))
			.finally(() => (loadingBlob = false));

		return () => {
			if (blobUrl) URL.revokeObjectURL(blobUrl);
		};
	});

	function closeWithAnimation() {
		if (closing) return;
		closing = true;
		const el = contentEl;
		if (el) {
			el.addEventListener('animationend', () => cleanup(), { once: true });
		}
		setTimeout(() => {
			if (closing) cleanup();
		}, 300);
	}

	function cleanup() {
		closing = false;
		close();
	}

	function onBackdrop(e: MouseEvent) {
		if (e.target === e.currentTarget) closeWithAnimation();
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeWithAnimation();
	}

	function fmtSize(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return (bytes / Math.pow(k, i)).toFixed(i > 0 ? 1 : 0) + ' ' + sizes[i];
	}
</script>

<svelte:window onkeydown={onKeydown} />

{#if open || closing}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm {closing ? 'animate-out fade-out-0' : 'animate-in fade-in-0'}"
		onclick={onBackdrop}
		role="dialog"
		aria-modal="true"
	>
		<div bind:this={contentEl} class="relative flex max-h-[90vh] w-full max-w-4xl flex-col rounded-xl border border-gray-100 bg-white shadow-2xl {closing ? 'animate-out fade-out-0 zoom-out-95' : 'animate-in fade-in-0 zoom-in-95'}">
			<!-- Header -->
			<div class="flex items-center gap-3 border-b border-gray-100 px-5 py-3.5">
				<div class="min-w-0 flex-1 truncate">
					<p class="truncate text-sm font-medium text-gray-800" title={name}>{name}</p>
					<p class="text-xs text-gray-400">{fmtSize(size)}</p>
				</div>
				<div class="flex shrink-0 items-center gap-1">
					<a
						href={dlUrl}
						download={name}
						class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
						title={m.download()}
					>
						<Download size={18} />
					</a>
					<button
						type="button"
						onclick={closeWithAnimation}
						class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
						title={m.close()}
					>
						<X size={18} />
					</button>
				</div>
			</div>

			<!-- Content -->
			<div class="flex-1 overflow-auto">
				{#if !canPreview}
					<div class="flex flex-col items-center gap-3 py-20 text-sm text-gray-400">
						<p>{m.unsupported_preview()}</p>
						<a
							href={dlUrl}
							download={name}
							class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
						>
							{m.download_file()}
						</a>
					</div>
				{:else if isImage}
					{#if loadingBlob}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-gray-400">{m.loading()}</p>
						</div>
					{:else if blobError}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-red-500">{blobError}</p>
						</div>
					{:else if blobUrl}
						<div class="flex items-center justify-center p-4">
							<img src={blobUrl} alt={name} class="max-h-[75vh] rounded-lg object-contain" />
						</div>
					{/if}
				{:else if isVideo}
					{#if loadingBlob}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-gray-400">{m.loading()}</p>
						</div>
					{:else if blobError}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-red-500">{blobError}</p>
						</div>
					{:else if blobUrl}
						<div class="flex items-center justify-center bg-black">
							<video controls autoplay class="max-h-[75vh] w-full">
								<source src={blobUrl} type={mimeType} />
							</video>
						</div>
					{/if}
				{:else if isAudio}
					{#if loadingBlob}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-gray-400">{m.loading()}</p>
						</div>
					{:else if blobError}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-red-500">{blobError}</p>
						</div>
					{:else if blobUrl}
						<div class="flex items-center justify-center py-16">
							<audio controls src={blobUrl} class="w-full max-w-md" />
						</div>
					{/if}
				{:else if isPdf}
					{#if loadingBlob}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-gray-400">{m.loading()}</p>
						</div>
					{:else if blobError}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-red-500">{blobError}</p>
						</div>
					{:else if blobUrl}
						<iframe src={blobUrl} class="h-[80vh] w-full" title={name}></iframe>
					{/if}
				{:else if isText}
					{#if loadingText}
						<div class="flex items-center justify-center py-16">
							<div class="text-sm text-gray-400">{m.loading()}</div>
						</div>
					{:else if textError}
						<div class="flex items-center justify-center py-16">
							<p class="text-sm text-red-500">{textError}</p>
						</div>
					{:else if textContent !== null}
						<pre class="overflow-x-auto p-5 text-sm leading-relaxed text-gray-700"><code>{textContent}</code></pre>
					{/if}
				{/if}
			</div>
		</div>
	</div>
{/if}
