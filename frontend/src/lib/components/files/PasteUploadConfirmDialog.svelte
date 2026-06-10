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
	bodyClass="!p-0"
>
	<div class="border-b border-gray-100 px-5 py-4">
		<div class="flex items-start gap-3">
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-blue-50 text-blue-600">
				<Clipboard size={20} />
			</div>
			<div class="min-w-0">
				<p class="text-sm text-gray-700">
					将上传 <span class="font-semibold text-gray-950">{acceptedFiles.length}</span> 个文件到
					<span class="font-semibold text-gray-950">{targetLabel}</span>
				</p>
				<p class="mt-1 text-xs text-gray-400">总大小 {fmtSize(totalSize)}</p>
			</div>
		</div>
	</div>

	{#if rejectedFiles.length > 0}
		<div class="mx-5 mt-4 rounded-xl border border-amber-100 bg-amber-50 px-3 py-2 text-sm text-amber-700">
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
					<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-500">
						<Check size={15} />
					</div>
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm text-gray-700" title={file.name}>{file.name}</p>
					</div>
					<span class="shrink-0 text-xs text-gray-400">{fmtSize(file.size)}</span>
				</div>
			{/each}
			{#if hiddenAcceptedCount > 0}
				<p class="px-3 py-2 text-xs text-gray-400">还有 {hiddenAcceptedCount} 个文件未显示</p>
			{/if}
		</div>
	{:else}
		<div class="px-5 py-8 text-center text-sm text-gray-500">
			没有可上传的文件。
		</div>
	{/if}

	<div class="flex items-center justify-end gap-2 border-t border-gray-100 px-5 py-3">
		<button
			type="button"
			onclick={onCancel}
			class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50"
		>
			<X size={14} /> 取消
		</button>
		<button
			type="button"
			onclick={onConfirm}
			disabled={acceptedFiles.length === 0}
			class="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
		>
			<Check size={14} /> 确认上传
		</button>
	</div>
</Dialog>
