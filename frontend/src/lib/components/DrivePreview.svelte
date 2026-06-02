<script lang="ts">
	import { Copy, Download } from '@lucide/svelte';
	import { downloadUrl } from '$lib/api/files';
	import { getAccessToken } from '$lib/api/client';
	import { fmtSize } from '$lib/utils/format';
	import { isCodeLikeFile, isJsonFile, isTextPreviewFile } from '$lib/utils/code-files';
	import { Dialog } from '$lib/ui/dialog';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';

	let {
		id,
		name,
		mimeType,
		size,
		open = $bindable(false),
		close
	}: {
		id: string;
		name: string;
		mimeType: string;
		size: number;
		open?: boolean;
		close: () => void;
	} = $props();

	let textContent = $state<string | null>(null);
	let textError = $state<string | null>(null);
	let loadingText = $state(false);

	let blobUrl = $state<string | null>(null);
	let blobError = $state<string | null>(null);
	let loadingBlob = $state(false);

	let dlUrl = $derived(downloadUrl(id));
	let authedDlUrl = $derived.by(() => {
		const token = getAccessToken();
		if (!token) return dlUrl;
		const url = new URL(dlUrl, window.location.origin);
		url.searchParams.set('access_token', token);
		return `${url.pathname}?${url.searchParams.toString()}`;
	});
	let isJson = $derived(isJsonFile(name, mimeType));
	let isCode = $derived(isCodeLikeFile(name, mimeType));
	let isText = $derived(isTextPreviewFile(name, mimeType));
	let canPreview = $derived(
		mimeType.startsWith('image/') ||
			mimeType.startsWith('video/') ||
			mimeType.startsWith('audio/') ||
			mimeType === 'application/pdf' ||
			isText
	);
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

	function handleOpenChange(v: boolean) {
		if (!v) {
			open = false;
			close();
		}
	}

	async function copyLink() {
		try {
			const url = new URL(authedDlUrl, window.location.origin);
			await navigator.clipboard.writeText(url.toString());
			toast.success(m.copied());
		} catch {
			toast.error(m.copy_failed());
		}
	}

	function formatTextContent(content: string) {
		if (!isJson) return content;
		try {
			return JSON.stringify(JSON.parse(content), null, 2);
		} catch {
			return content;
		}
	}
</script>

<Dialog
	bind:open
	onOpenChange={handleOpenChange}
	onCancel={close}
	title={name}
	description={fmtSize(size)}
	footer={false}
	class="!w-[70vw] !max-w-[70vw] max-sm:!w-[calc(100vw-1rem)] max-sm:!max-w-[calc(100vw-1rem)]"
	bodyClass="!p-0"
>
	{#snippet headerExtra()}
		<button
			type="button"
			onclick={copyLink}
			class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
			title={m.copy_url()}
			aria-label={m.copy_url()}
		>
			<Copy size={18} />
		</button>
		<a
			href={authedDlUrl}
			download={name}
			class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
			title={m.download()}
			aria-label={m.download()}
		>
			<Download size={18} />
		</a>
	{/snippet}

	{#if !canPreview}
		<div class="flex flex-col items-center gap-3 py-20 text-sm text-gray-400">
			<p>{m.unsupported_preview()}</p>
			<a
				href={authedDlUrl}
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
				<audio controls src={blobUrl} class="w-full max-w-md"></audio>
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
			<pre class="max-h-[80vh] overflow-auto p-5 text-sm leading-relaxed text-gray-700"><code>{formatTextContent(textContent)}</code></pre>
		{/if}
	{/if}
</Dialog>
