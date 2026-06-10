<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { toast } from 'svelte-sonner';
	import { extractClipboardFiles, filterPasteFiles, isEditablePasteTarget } from '$lib/paste-upload';
	import PasteUploadConfirmDialog from './PasteUploadConfirmDialog.svelte';

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

	function reset() {
		dialogOpen = false;
		acceptedFiles = [];
		rejectedFiles = [];
	}

	function handlePaste(event: ClipboardEvent) {
		if (!enabled) return;

		const pastedFiles = extractClipboardFiles(event.clipboardData);
		if (pastedFiles.length === 0) return;
		if (isEditablePasteTarget(event.target)) return;

		event.preventDefault();
		const result = filterPasteFiles(pastedFiles, acceptFile);
		acceptedFiles = result.accepted;
		rejectedFiles = result.rejected;

		if (acceptedFiles.length === 0) {
			toast.error(result.rejected.length > 0 ? '粘贴的文件类型不支持上传' : '剪贴板中没有可上传的文件');
			return;
		}

		dialogOpen = true;
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
