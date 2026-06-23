<script lang="ts">
	import { fmtSize } from '$lib/utils/format';
	import { AlertTriangle, Check, Clipboard, X } from '@lucide/svelte';
	import { Dialog } from '$lib/ui/dialog';

	let {
		acceptedFiles,
		rejectedFiles = [],
		targetLabel,
		open = $bindable(false),
		onConfirm,
		onCancel,
	}: {
		acceptedFiles: File[];
		rejectedFiles?: File[];
		targetLabel: string;
		open?: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	const MAX_VISIBLE_FILES = 8;
	let visibleAccepted = $derived(acceptedFiles.slice(0, MAX_VISIBLE_FILES));
	let hiddenAcceptedCount = $derived(Math.max(0, acceptedFiles.length - visibleAccepted.length));
	let totalSize = $derived(acceptedFiles.reduce((sum, file) => sum + file.size, 0));

	function handleOpenChange(value: boolean) {
		if (!value) onCancel();
	}
</script>

<Dialog
	bind:open
	onOpenChange={handleOpenChange}
	onCancel={onCancel}
	title="确认上传粘贴的文件？"
	footer={false}
	class="max-w-lg"
	bodyClass="p-0"
>
	<div class="border-b border-line-soft px-5 py-4">
		<div class="flex items-start gap-3">
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-soft text-primary">
				<Clipboard size={20} />
			</div>
			<div class="min-w-0">
				<p class="text-sm text-ink-2">
					将上传 <span class="font-semibold text-ink">{acceptedFiles.length}</span> 个文件到
					<span class="font-semibold text-ink">{targetLabel}</span>
				</p>
				<p class="mt-1 text-xs text-ink-4">总大小 {fmtSize(totalSize)}</p>
			</div>
		</div>
	</div>

	{#if rejectedFiles.length > 0}
		<div class="mx-5 mt-4 rounded-xl border border-warning bg-warning-soft px-3 py-2 text-sm text-warning">
			<div class="flex items-start gap-2">
				<AlertTriangle size={16} class="mt-0.5 shrink-0" />
				<p>已跳过 {rejectedFiles.length} 个不支持的文件。</p>
			</div>
		</div>
	{/if}

	{#if acceptedFiles.length > 0}
		<div class="max-h-[45vh] overflow-y-auto px-2 py-3">
			{#each visibleAccepted as file (file.name + file.size + file.lastModified)}
				<div class="flex items-center gap-3 rounded-lg px-3 py-2">
					<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-surface-sunken text-ink-3">
						<Check size={15} />
					</div>
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm text-ink-2" title={file.name}>{file.name}</p>
					</div>
					<span class="shrink-0 text-xs text-ink-4">{fmtSize(file.size)}</span>
				</div>
			{/each}
			{#if hiddenAcceptedCount > 0}
				<p class="px-3 py-2 text-xs text-ink-4">还有 {hiddenAcceptedCount} 个文件未显示</p>
			{/if}
		</div>
	{:else}
		<div class="px-5 py-8 text-center text-sm text-ink-3">
			没有可上传的文件。
		</div>
	{/if}

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
			onclick={onConfirm}
			disabled={acceptedFiles.length === 0}
			class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> 确认上传
		</button>
	</div>
</Dialog>
