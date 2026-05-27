<script lang="ts">
	import {
		FileVideo,
		FileAudio,
		FileImage,
		FileText,
		FileArchive,
		File,
		Folder
	} from '@lucide/svelte';

	let {
		mimeType,
		isDir = false,
		size = 20,
		class: className = ''
	}: {
		mimeType: string | null;
		isDir?: boolean;
		size?: number;
		class?: string;
	} = $props();

	let mt = $derived(mimeType ?? '');

	let icon = $derived.by(() => {
		if (isDir) return Folder;
		if (mt.startsWith('video/')) return FileVideo;
		if (mt.startsWith('audio/')) return FileAudio;
		if (mt.startsWith('image/')) return FileImage;
		if (mt.startsWith('text/') || mt.includes('pdf')) return FileText;
		if (mt.includes('zip') || mt.includes('rar') || mt.includes('tar') || mt.includes('gzip') || mt.includes('7z')) return FileArchive;
		return File;
	});

	let color = $derived.by(() => {
		if (isDir) return 'text-blue-500';
		if (mt.startsWith('video/')) return 'text-purple-500';
		if (mt.startsWith('audio/')) return 'text-pink-500';
		if (mt.startsWith('image/')) return 'text-emerald-500';
		if (mt.startsWith('text/') || mt.includes('pdf')) return 'text-orange-500';
		if (mt.includes('zip') || mt.includes('rar') || mt.includes('tar') || mt.includes('gzip') || mt.includes('7z')) return 'text-yellow-600';
		return 'text-gray-400';
	});
</script>

<svelte:component this={icon} {size} class="{color} {className}" />
