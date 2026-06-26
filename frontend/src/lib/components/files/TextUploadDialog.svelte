<script lang="ts">
	import { fmtSize } from '$lib/utils/format';
	import { FileText, Check, X, ChevronDown, ChevronUp, AlertTriangle } from '@lucide/svelte';
	import { Dialog } from '$lib/ui/dialog';
	import { MAX_PASTE_TEXT_SIZE } from '$lib/paste-text-upload';
	import * as m from '$lib/paraglide/messages';

	let {
		open = $bindable(false),
		targetLabel,
		onConfirm,
		onCancel
	}: {
		open?: boolean;
		targetLabel: string;
		onConfirm: (file: File) => void | Promise<void>;
		onCancel: () => void;
	} = $props();

	let text = $state('');
	let fileName = $state('');
	let submitting = $state(false);
	let expanded = $state(false);
	let filenameInput: HTMLInputElement | undefined = $state();

	const PREVIEW_MAX_LENGTH = 500;
	const isLongText = $derived(text.length > PREVIEW_MAX_LENGTH);
	const displayText = $derived(expanded || !isLongText ? text : text.slice(0, PREVIEW_MAX_LENGTH));
	const charCount = $derived(text.length);
	const textSize = $derived(new Blob([text]).size);
	const sizeError = $derived(textSize > MAX_PASTE_TEXT_SIZE ? m.text_size_exceeded({ current: fmtSize(textSize), max: fmtSize(MAX_PASTE_TEXT_SIZE) }) : undefined);
	const canConfirm = $derived(text.trim().length > 0 && fileName.trim().length > 0 && !sizeError && !submitting);

	$effect(() => {
		if (open) {
			text = '';
			fileName = '';
			submitting = false;
			expanded = false;
		}
	});

	function onOpenAutoFocus(e: Event) {
		e.preventDefault();
		requestAnimationFrame(() => filenameInput?.focus());
	}

	function handleClose(value: boolean) {
		if (!value) onCancel();
	}

	async function handleConfirm() {
		if (!canConfirm) return;
		submitting = true;
		try {
			const file = new File([text], fileName.trim(), { type: 'text/plain;charset=utf-8' });
			await onConfirm(file);
			open = false;
		} finally {
			submitting = false;
		}
	}
</script>

<Dialog
	bind:open
	onOpenChangeComplete={handleClose}
	onOpenAutoFocus={onOpenAutoFocus}
	title={sizeError ? m.paste_text_size_exceeded() : m.text_upload_title()}
	footer={false}
	size="md"
	bodyClass="p-0"
>
	{#if sizeError}
		<!-- Error header -->
		<div class="border-b border-error/30 bg-error-soft px-5 py-4">
			<div class="flex items-start gap-2.5">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-error/10 text-error">
					<AlertTriangle size={20} />
				</div>
				<div class="min-w-0">
					<p class="text-sm text-ink">{sizeError}</p>
				</div>
			</div>
		</div>
	{:else}
		<!-- Header -->
		<div class="border-b border-line-soft px-5 py-4">
			<div class="flex items-start gap-3">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
					<FileText size={20} />
				</div>
				<div class="min-w-0">
					<p class="text-sm text-ink-2">
						{m.text_uploading_to({ target: targetLabel })}
					</p>
					<p class="mt-1 text-xs text-ink-4">
						{m.text_upload_size({ size: fmtSize(textSize), count: String(charCount) })}
					</p>
				</div>
			</div>
		</div>

		<!-- Filename input -->
		<div class="border-b border-line-soft px-5 py-3">
			<label for="filename-input" class="text-sm font-medium text-ink-2">
				{m.text_filename_label()}
			</label>
			<input
				id="filename-input"
				type="text"
				bind:value={fileName}
				bind:this={filenameInput}
				placeholder={m.text_filename_placeholder()}
				class="mt-1.5 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
			/>
		</div>

		<!-- Text content input -->
		<div class="border-b border-line-soft px-5 py-3">
			<label for="text-content" class="text-sm font-medium text-ink-2">
				{m.text_content_label()}
			</label>
			<textarea
				id="text-content"
				bind:value={text}
				rows={6}
				placeholder={m.text_content_placeholder()}
				class="mt-1.5 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20 resize-none"
			></textarea>
		</div>

		<!-- Text preview -->
		{#if charCount > 0}
			<div class="border-b border-line-soft px-5 py-3">
				<button
					type="button"
					onclick={() => (expanded = !expanded)}
					class="flex w-full items-center justify-between text-left"
				>
					<span class="text-sm font-medium text-ink-2">{m.paste_text_preview()}</span>
					<div class="flex items-center gap-1.5 text-xs text-ink-4">
						<span>{m.paste_text_chars({ count: String(charCount) })}</span>
						{#if isLongText}
							{#if expanded}
								<ChevronUp size={14} />
							{:else}
								<ChevronDown size={14} />
							{/if}
						{/if}
					</div>
				</button>
				{#if isLongText && !expanded}
					<p class="mt-2 rounded-lg bg-surface-sunken p-2.5 font-mono text-xs text-ink-3 whitespace-pre-wrap">
						{displayText}...
					</p>
				{:else}
					<p
						class="mt-2 max-h-[30vh] overflow-y-auto rounded-lg bg-surface-sunken p-2.5 font-mono text-xs text-ink-3 whitespace-pre-wrap"
					>
						{displayText}
					</p>
				{/if}
			</div>
		{/if}
	{/if}

	<!-- Footer -->
	<div class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3">
		<button
			type="button"
			onclick={onCancel}
			class="inline-flex items-center gap-1.5 rounded-lg border border-line bg-white px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-muted"
		>
			<X size={14} /> {m.cancel()}
		</button>
		<button
			type="button"
			onclick={handleConfirm}
			disabled={!canConfirm}
			class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> {m.confirm()}
		</button>
	</div>
</Dialog>
