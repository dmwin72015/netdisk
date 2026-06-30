<script lang="ts">
	import { fmtSize } from '$lib/utils/format';
	import { FileText, Check, X, AlertTriangle, ChevronDown, ChevronUp } from '@lucide/svelte';
	import { Dialog } from '$lib/ui/dialog';
	import { MAX_PASTE_TEXT_SIZE, getDefaultFileName } from '$lib/paste-text-upload';
	import * as m from '$lib/paraglide/messages';

	let {
		open = $bindable(false),
		text,
		targetLabel,
		defaultFileName,
		sizeError,
		onConfirm,
		onCancel
	}: {
		open?: boolean;
		text: string;
		targetLabel: string;
		defaultFileName?: string;
		sizeError?: string;
		onConfirm: (file: File) => void | Promise<void>;
		onCancel: () => void;
	} = $props();

	let fileName = $state('');
	let expanded = $state(false);

	const PREVIEW_MAX_LENGTH = 500;
	const isLongText = $derived(text.length > PREVIEW_MAX_LENGTH);
	const displayText = $derived(expanded || !isLongText ? text : text.slice(0, PREVIEW_MAX_LENGTH));

	$effect(() => {
		if (open) {
			fileName = defaultFileName ?? getDefaultFileName(text);
		}
	});

	function handleOpenChangeComplete(value: boolean) {
		if (!value) {
			onCancel();
		}
	}

	function handleConfirm() {
		if (!fileName.trim() || sizeError) return;
		const file = new File([text], fileName.trim(), { type: 'text/plain;charset=utf-8' });
		onConfirm(file);
		open = false;
	}

	function togglePreview() {
		expanded = !expanded;
	}
</script>

<Dialog
	bind:open
	onOpenChangeComplete={handleOpenChangeComplete}
	onCancel={onCancel}
	title={sizeError ? m.paste_text_size_exceeded() : m.paste_text_title()}
	footer={false}
	size="md"
	bodyClass="p-0"
	closable={false}
>
	{#if sizeError}
		<!-- Error state -->
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
		<!-- Normal state header -->
		<div class="border-b border-line-soft px-5 py-4">
			<div class="flex items-start gap-3">
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
					<FileText size={20} />
				</div>
				<div class="min-w-0">
					<p class="text-sm text-ink-2">
						{m.paste_saving_text({ target: targetLabel })}
					</p>
					<p class="mt-1 text-xs text-ink-4">
						{m.paste_text_size({ size: fmtSize(new Blob([text]).size), count: String(text.length) })}
					</p>
				</div>
			</div>
		</div>

		<!-- Filename input -->
		<div class="border-b border-line-soft px-5 py-3">
<label for="filename" class="text-sm font-medium text-ink-3">{m.paste_filename()}</label>
			<input
				id="filename"
				type="text"
				bind:value={fileName}
				placeholder={m.paste_filename_placeholder()}
				class="mt-1.5 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
			/>
		</div>

		<!-- Preview -->
		<div class="border-b border-line-soft px-5 py-3">
			<button
				type="button"
				onclick={togglePreview}
				class="flex w-full items-center justify-between text-left"
			>
				<span class="text-sm font-medium text-ink-3">{m.paste_text_preview()}</span>
				<div class="flex items-center gap-1.5 text-xs text-ink-4">
					<span>{m.paste_text_chars({ count: String(text.length) })}</span>
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
				<p class="mt-2 rounded-lg bg-surface-sunken p-2.5 font-mono text-xs text-ink-3">
					{displayText}...
				</p>
			{:else}
				<p class="mt-2 max-h-[30vh] overflow-y-auto rounded-lg bg-surface-sunken p-2.5 font-mono text-xs text-ink-3 whitespace-pre-wrap">
					{displayText}
				</p>
			{/if}
		</div>
	{/if}

	<!-- Footer -->
	<div class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3">
		<button
			type="button"
			onclick={() => { open = false; }}
			class="inline-flex items-center gap-1.5 rounded-lg border border-line bg-surface px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-sunken"
		>
			<X size={14} /> {m.cancel()}
		</button>
		<button
			type="button"
			onclick={handleConfirm}
			disabled={!fileName.trim() || !!sizeError}
			class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> {m.confirm()}
		</button>
	</div>
</Dialog>
