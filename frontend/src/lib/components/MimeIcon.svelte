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

	// File-type tints. These are SIGNAL colors (let users scan a list by type),
	// not decoration. Hues are spaced across the wheel and toned to coexist
	// with the cool-gray neutrals; chroma is intentionally one notch below
	// Tailwind defaults so they don't punch through the page.
	const categoryColorMap: Record<string, string> = {
		folder: 'text-primary',
		video: 'text-[oklch(60%_0.12_300)]',   // muted plum
		audio: 'text-[oklch(64%_0.14_350)]',   // muted rose
		image: 'text-[oklch(62%_0.13_165)]',   // muted teal-green
		document: 'text-[oklch(64%_0.14_55)]', // muted amber
		archive: 'text-[oklch(65%_0.12_85)]',  // muted ochre
		other: 'text-ink-4'
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
			if (isDir) return 'text-primary';
			if (isCode) return 'text-[oklch(58%_0.13_220)]';
			return categoryColorMap[category] ?? 'text-ink-4';
		}
		if (isDir) return 'text-primary';
		if (mt.startsWith('video/')) return categoryColorMap.video;
		if (mt.startsWith('audio/')) return categoryColorMap.audio;
		if (mt.startsWith('image/')) return categoryColorMap.image;
		if (isCode) return 'text-[oklch(58%_0.13_220)]';
		if (mt.startsWith('text/') || mt.includes('pdf')) return categoryColorMap.document;
		if (mt.includes('zip') || mt.includes('rar') || mt.includes('tar') || mt.includes('gzip') || mt.includes('7z')) return categoryColorMap.archive;
		return 'text-ink-4';
	});
</script>

<Icon {size} class="{color} {className}" />
