<script lang="ts">
	import { Copy, Download, WrapText } from '@lucide/svelte';
	import { downloadUrl } from '$lib/api/files';
	import { getAccessToken } from '$lib/api/client';
	import { fmtSize, copyToClipboard } from '$lib/utils/format';
	import { isJsonFile, isTextPreviewFile } from '$lib/utils/code-files';
	import VirtualTextViewer from './VirtualTextViewer.svelte';
	import { Dialog } from '$lib/ui/dialog';
	import * as m from '$lib/paraglide/messages';
	import { toast } from 'svelte-sonner';

	const JSON_FORMAT_LIMIT = 512 * 1024;
	const LARGE_TEXT_LIMIT = 1024 * 1024;
	const LARGE_TEXT_PREVIEW_LINES = 50;
	const LARGE_TEXT_MAX_PREVIEW_CHARS = 512 * 1024;

	type TextPreviewResult = {
		content: string;
		truncated: boolean;
	};

	let {
		id,
		name,
		mimeType,
		size,
		open = $bindable(false),
		onOpenChangeComplete
	}: {
		id: string;
		name: string;
		mimeType: string;
		size: number;
		open?: boolean;
		onOpenChangeComplete?: (open: boolean) => void;
	} = $props();

	let textContent = $state<string | null>(null);
	let textError = $state<string | null>(null);
	let textTruncated = $state(false);
	let loadingText = $state(false);
	let textWrap = $state(false);

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
	let shouldLimitTextPreview = $derived(size > LARGE_TEXT_LIMIT);
	let formattedTextContent = $derived(textContent === null ? null : formatTextContent(textContent));

	$effect(() => {
		if (!open || !isText) return;
		textContent = null;
		textError = null;
		textTruncated = false;
		loadingText = true;
		const controller = new AbortController();
		const token = getAccessToken() ?? '';
		fetch(dlUrl, { headers: { Authorization: `Bearer ${token}` }, signal: controller.signal })
			.then((r) => {
				if (!r.ok) throw new Error('Failed to load');
				return readTextPreview(r, shouldLimitTextPreview);
			})
			.then(({ content, truncated }) => {
				textContent = content;
				textTruncated = truncated;
			})
			.catch((e) => {
				if (e.name !== 'AbortError') textError = e.message;
			})
			.finally(() => {
				if (!controller.signal.aborted) loadingText = false;
			});

		return () => controller.abort();
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
		}
	}

	function handleClose() {
		open = false;
	}

	async function copyLink() {
		const url = new URL(authedDlUrl, window.location.origin);
		if (await copyToClipboard(url.toString())) {
			toast.success(m.copied());
		} else {
			toast.error(m.copy_failed());
		}
	}

	function formatTextContent(content: string) {
		if (textTruncated || !isJson || content.length > JSON_FORMAT_LIMIT) return content;
		try {
			return JSON.stringify(JSON.parse(content), null, 2);
		} catch {
			return content;
		}
	}

	async function readTextPreview(response: Response, limitLines: boolean): Promise<TextPreviewResult> {
		if (!limitLines || !response.body) {
			return { content: await response.text(), truncated: false };
		}

		const reader = response.body.getReader();
		const decoder = new TextDecoder();
		let content = '';
		let truncated = false;

		try {
			while (true) {
				const { value, done } = await reader.read();
				if (done) break;

				content += decoder.decode(value, { stream: true });
				const limited = limitTextLines(content);

				if (limited.truncated || content.length >= LARGE_TEXT_MAX_PREVIEW_CHARS) {
					content = limited.truncated ? limited.content : content.slice(0, LARGE_TEXT_MAX_PREVIEW_CHARS);
					truncated = true;
					await reader.cancel();
					break;
				}
			}

			if (!truncated) content += decoder.decode();
		} finally {
			reader.releaseLock();
		}

		const limited = limitTextLines(content);
		return { content: limited.content, truncated: truncated || limited.truncated };
	}

	function limitTextLines(content: string): TextPreviewResult {
		const lines = content.split(/\r\n|\r|\n/);
		if (lines.length <= LARGE_TEXT_PREVIEW_LINES) return { content, truncated: false };

		return {
			content: lines.slice(0, LARGE_TEXT_PREVIEW_LINES).join('\n'),
			truncated: true
		};
	}
</script>

<Dialog
	bind:open
	onOpenChange={handleOpenChange}
	onOpenChangeComplete={onOpenChangeComplete}
	title={name}
	description={fmtSize(size)}
	footer={false}
	class="w-[70vw]! max-w-[70vw]! max-sm:w-[calc(100vw-1rem)]! max-sm:max-w-[calc(100vw-1rem)]!"
	bodyClass="p-0"
>
	{#snippet headerExtra()}
		<button
			type="button"
			onclick={copyLink}
			class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
			title={m.copy_url()}
			aria-label={m.copy_url()}
		>
			<Copy size={18} />
		</button>
		<a
			href={authedDlUrl}
			download={name}
			class="rounded-lg p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
			title={m.download()}
			aria-label={m.download()}
		>
			<Download size={18} />
		</a>
	{/snippet}

	{#if !canPreview}
		<div class="flex flex-col items-center gap-3 py-20 text-sm text-ink-4">
			<p>{m.unsupported_preview()}</p>
			<a
				href={authedDlUrl}
				download={name}
				class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
			>
				{m.download_file()}
			</a>
		</div>
	{:else if isImage}
		{#if loadingBlob}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-ink-4">{m.loading()}</p>
			</div>
		{:else if blobError}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-danger">{blobError}</p>
			</div>
		{:else if blobUrl}
			<div class="flex items-center justify-center p-4">
					<img src={blobUrl} alt={name} loading="lazy" class="max-h-[75vh] rounded-lg object-contain" />
			</div>
		{/if}
	{:else if isVideo}
		{#if loadingBlob}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-ink-4">{m.loading()}</p>
			</div>
		{:else if blobError}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-danger">{blobError}</p>
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
				<p class="text-sm text-ink-4">{m.loading()}</p>
			</div>
		{:else if blobError}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-danger">{blobError}</p>
			</div>
		{:else if blobUrl}
			<div class="flex items-center justify-center py-16">
				<audio controls src={blobUrl} class="w-full max-w-md"></audio>
			</div>
		{/if}
	{:else if isPdf}
		{#if loadingBlob}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-ink-4">{m.loading()}</p>
			</div>
		{:else if blobError}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-danger">{blobError}</p>
			</div>
		{:else if blobUrl}
			<iframe src={blobUrl} class="h-[80vh] w-full" title={name}></iframe>
		{/if}
	{:else if isText}
		{#if loadingText}
			<div class="flex items-center justify-center py-16">
				<div class="text-sm text-ink-4">{m.loading()}</div>
			</div>
		{:else if textError}
			<div class="flex items-center justify-center py-16">
				<p class="text-sm text-danger">{textError}</p>
			</div>
		{:else if formattedTextContent !== null}
			<div class="flex items-center justify-end gap-2 border-b border-line-soft bg-surface-muted/60 px-4 py-2">
				<button
					type="button"
					onclick={() => (textWrap = !textWrap)}
					class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs transition-colors {textWrap ? 'bg-primary-soft text-primary' : 'text-ink-3 hover:bg-surface-sunken hover:text-ink-2'}"
					title={textWrap ? m.text_preview_no_wrap() : m.text_preview_wrap()}
				>
					<WrapText size={14} />
					<span>{textWrap ? m.text_preview_no_wrap() : m.text_preview_wrap()}</span>
				</button>
			</div>
			{#if textTruncated}
				<div class="border-b border-warning bg-warning-soft px-5 py-2 text-xs text-warning">
					{m.text_preview_truncated({ lines: String(LARGE_TEXT_PREVIEW_LINES) })}
				</div>
			{/if}
			<VirtualTextViewer content={formattedTextContent} ariaLabel={m.preview()} wrap={textWrap} />
		{/if}
	{/if}
</Dialog>
