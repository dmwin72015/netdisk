<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { toast } from 'svelte-sonner';
	import { extractClipboardFiles, filterPasteFiles, isEditablePasteTarget } from '$lib/paste-upload';
	import { extractClipboardText, validateTextSize, getDefaultFileName, createTextFile } from '$lib/paste-text-upload';
	import PasteUploadConfirmDialog from './PasteUploadConfirmDialog.svelte';
	import PasteTextUploadConfirmDialog from './PasteTextUploadConfirmDialog.svelte';

	let {
		enabled = true,
		targetLabel,
		acceptFile,
		onUpload,
	}: {
		enabled?: boolean;
		targetLabel: string;
		acceptFile?: (file: File) => boolean;
		onUpload: (files: File[]) => void | Promise<unknown>;
	} = $props();

	let dialogOpen = $state(false);
	let acceptedFiles = $state<File[]>([]);
	let rejectedFiles = $state<File[]>([]);
	let textDialogOpen = $state(false);
	let clipboardText = $state<string | null>(null);
	let textFileName = $state('');

	function reset() {
		dialogOpen = false;
		acceptedFiles = [];
		rejectedFiles = [];
	}

	async function confirmTextUpload(file: File) {
		textDialogOpen = false;
		try {
			await onUpload([file]);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : '粘贴上传失败');
		} finally {
			clipboardText = null;
			textFileName = '';
		}
	}

	function handlePaste(event: ClipboardEvent) {
		if (!enabled) return;

		// 优先检测文件
		const pastedFiles = extractClipboardFiles(event.clipboardData);
		if (pastedFiles.length > 0) {
			const result = filterPasteFiles(pastedFiles, acceptFile);
			acceptedFiles = result.accepted;
			rejectedFiles = result.rejected;
			if (acceptedFiles.length > 0) dialogOpen = true;
			return;
		}

		// 新增：检测文本
		const text = extractClipboardText(event.clipboardData);
		if (text && text.trim().length > 0) {
			event.preventDefault();

			// 检查文本大小
			const sizeCheck = validateTextSize(text);
			if (!sizeCheck.valid) {
				toast.error(sizeCheck.error || '文本内容过大');
				return;
			}

			clipboardText = text;
			textFileName = getDefaultFileName(text);
			textDialogOpen = true;
		}
	}

	async function confirmUpload() {
		const files = acceptedFiles;
		reset();
		try {
			await onUpload(files);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : '粘贴上传失败');
		}
	}

	onMount(() => {
		if (!browser) return;
		window.addEventListener('paste', handlePaste);
	});

	onDestroy(() => {
		if (!browser) return;
		window.removeEventListener('paste', handlePaste);
	});
</script>

<PasteUploadConfirmDialog
	bind:open={dialogOpen}
	{acceptedFiles}
	{rejectedFiles}
	{targetLabel}
	onConfirm={confirmUpload}
	onCancel={reset}
/>

<PasteTextUploadConfirmDialog
	bind:open={textDialogOpen}
	text={clipboardText || ''}
	targetLabel={targetLabel}
	defaultFileName={textFileName}
	onConfirm={confirmTextUpload}
	onCancel={() => {
		textDialogOpen = false;
		clipboardText = null;
		textFileName = '';
	}}
/>
