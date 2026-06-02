<script lang="ts">
	import {
		FileVideo,
		FileAudio,
		FileImage,
		FileText,
		FileCodeCorner,
		FileArchive,
		File,
		Folder
	} from '@lucide/svelte';
	import { isCodeLikeFile } from '$lib/utils/code-files';

	let {
		mimeType,
		name = '',
		isDir = false,
		category,
		size = 20,
		class: className = ''
	}: {
		mimeType: string | null;
		name?: string;
		isDir?: boolean;
		category?: string;
		size?: number;
		class?: string;
	} = $props();

	let mt = $derived(mimeType ?? '');
	let isCode = $derived(isCodeLikeFile(name, mt));

	const categoryIconMap: Record<string, typeof File> = {
		folder: Folder,
		video: FileVideo,
		audio: FileAudio,
		image: FileImage,
		document: FileText,
		archive: FileArchive,
		other: File
	};

	const categoryColorMap: Record<string, string> = {
		folder: 'text-blue-500',
		video: 'text-purple-500',
		audio: 'text-pink-500',
		image: 'text-emerald-500',
		document: 'text-orange-500',
		archive: 'text-yellow-600',
		other: 'text-gray-400'
	};

	let Icon = $derived.by(() => {
		if (category) {
			if (isDir) return Folder;
			if (isCode) return FileCodeCorner;
			return categoryIconMap[category] ?? File;
		}
		if (isDir) return Folder;
		if (mt.startsWith('video/')) return FileVideo;
		if (mt.startsWith('audio/')) return FileAudio;
		if (mt.startsWith('image/')) return FileImage;
		if (isCode) return FileCodeCorner;
		if (mt.startsWith('text/') || mt.includes('pdf')) return FileText;
		if (mt.includes('zip') || mt.includes('rar') || mt.includes('tar') || mt.includes('gzip') || mt.includes('7z')) return FileArchive;
		return File;
	});

	let color = $derived.by(() => {
		if (category) {
			if (isDir) return 'text-blue-500';
			if (isCode) return 'text-sky-600';
			return categoryColorMap[category] ?? 'text-gray-400';
		}
		if (isDir) return 'text-blue-500';
		if (mt.startsWith('video/')) return 'text-purple-500';
		if (mt.startsWith('audio/')) return 'text-pink-500';
		if (mt.startsWith('image/')) return 'text-emerald-500';
		if (isCode) return 'text-sky-600';
		if (mt.startsWith('text/') || mt.includes('pdf')) return 'text-orange-500';
		if (mt.includes('zip') || mt.includes('rar') || mt.includes('tar') || mt.includes('gzip') || mt.includes('7z')) return 'text-yellow-600';
		return 'text-gray-400';
	});
</script>

<Icon {size} class="{color} {className}" />
