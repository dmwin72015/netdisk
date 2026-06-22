<script lang="ts">
	import { fmtSize } from '$lib/utils/format';
	import { FileText, Check, X, AlertTriangle, ChevronDown, ChevronUp } from '@lucide/svelte';
	import { Dialog } from '$lib/ui/dialog';
	import { MAX_PASTE_TEXT_SIZE, getDefaultFileName } from '$lib/paste-text-upload';

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

	function handleOpenChange(value: boolean) {
		if (!value) onCancel();
	}

	function handleConfirm() {
		if (!fileName.trim() || sizeError) return;
		const file = new File([text], fileName.trim(), { type: 'text/plain;charset=utf-8' });
		onConfirm(file);
	}

	function togglePreview() {
		expanded = !expanded;
	}
</script>

<Dialog
	bind:open
	onOpenChange={handleOpenChange}
	onCancel={onCancel}
	title={sizeError ? '文本大小超出限制' : '确认粘贴文本'}
	footer={false}
	class="max-w-lg"
>
	{#if sizeError}
		<!-- Error state -->
		<div class="border-b border-error/30 bg-error-soft px-5 py-4">
			<div class="flex items-start gap-3">
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
						将粘贴文本保存到 <span class="font-semibold text-ink">{targetLabel}</span>
					</p>
					<p class="mt-1 text-xs text-ink-4">
						文本大小 {fmtSize(new Blob([text]).size)}，共 {text.length} 个字符
					</p>
				</div>
			</div>
		</div>

		<!-- Filename input -->
		<div class="border-b border-line-soft px-5 py-3">
			<label for="filename" class="text-xs font-medium text-ink-3">文件名</label>
			<input
				id="filename"
				type="text"
				bind:value={fileName}
				placeholder="请输入文件名"
				class="mt-1.5 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-primary focus:ring-2 focus:ring-primary/20"
			/>
		</div>

		<!-- Preview -->
		<div class="border-b border-line-soft px-5 py-3">
			<button
				type="button"
				onclick={togglePreview}
				class="flex w-full items-center justify-between text-left"
			>
				<span class="text-xs font-medium text-ink-3">文本预览</span>
				<div class="flex items-center gap-1.5 text-xs text-ink-4">
					<span>{text.length} 个字符</span>
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
				<p class="mt-2 rounded-lg bg-surface-sunken p-3 font-mono text-xs text-ink-3">
					{displayText}...
				</p>
			{:else}
				<p class="mt-2 max-h-[30vh] overflow-y-auto rounded-lg bg-surface-sunken p-3 font-mono text-xs text-ink-3 whitespace-pre-wrap">
					{displayText}
				</p>
			{/if}
		</div>
	{/if}

	<!-- Footer -->
	<div class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3">
		<button
			type="button"
			onclick={onCancel}
			class="inline-flex items-center gap-1.5 rounded-lg border border-line bg-white px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-muted"
		>
			<X size={14} /> 取消
		</button>
		<button
			type="button"
			onclick={handleConfirm}
			disabled={!fileName.trim() || !!sizeError}
			class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> 确认
		</button>
	</div>
</Dialog>
