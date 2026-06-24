<script lang="ts">
	import { getFileExtension, isCodeLikeFile } from '$lib/utils/code-files';

	import apkIcon from '$lib/assets/file-types/apk.png';
	import archiveIcon from '$lib/assets/file-types/archive.png';
	import audioIcon from '$lib/assets/file-types/audio.png';
	import bmpIcon from '$lib/assets/file-types/bmp.png';
	import codeIcon from '$lib/assets/file-types/code.png';
	import excelIcon from '$lib/assets/file-types/excel.png';
	import folderIcon from '$lib/assets/file-types/folder.png';
	import gifIcon from '$lib/assets/file-types/gif.png';
	import heicIcon from '$lib/assets/file-types/heic.png';
	import imageIcon from '$lib/assets/file-types/image.png';
	import jpgIcon from '$lib/assets/file-types/jpg.png';
	import pdfIcon from '$lib/assets/file-types/pdf.png';
	import pngIcon from '$lib/assets/file-types/png.png';
	import pptIcon from '$lib/assets/file-types/ppt.png';
	import psdIcon from '$lib/assets/file-types/psd.png';
	import svgIcon from '$lib/assets/file-types/svg.png';
	import textIcon from '$lib/assets/file-types/text.png';
	import tifIcon from '$lib/assets/file-types/tif.png';
	import unknownIcon from '$lib/assets/file-types/unknown.png';
	import videoIcon from '$lib/assets/file-types/video.png';
	import wordIcon from '$lib/assets/file-types/word.png';

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

	const mt = $derived(mimeType ?? '');
	const ext = $derived(getFileExtension(name));
	const isCode = $derived(isCodeLikeFile(name, mt));

	// Extension → icon (highest priority after isDir). Covers types where we
	// have a dedicated visual; falls through to category/mime if no match.
	const extIconMap: Record<string, string> = {
		// documents
		pdf: pdfIcon,
		doc: wordIcon,
		docx: wordIcon,
		xls: excelIcon,
		xlsx: excelIcon,
		csv: excelIcon,
		ppt: pptIcon,
		pptx: pptIcon,
		psd: psdIcon,
		txt: textIcon,
		// images
		png: pngIcon,
		jpg: jpgIcon,
		jpeg: jpgIcon,
		gif: gifIcon,
		bmp: bmpIcon,
		svg: svgIcon,
		webp: imageIcon, // no dedicated webp icon — fall to generic image
		heic: heicIcon,
		heif: heicIcon,
		tif: tifIcon,
		tiff: tifIcon,
		// archives
		zip: archiveIcon,
		rar: archiveIcon,
		tar: archiveIcon,
		gz: archiveIcon,
		'7z': archiveIcon,
		bz2: archiveIcon,
		xz: archiveIcon,
		// applications
		apk: apkIcon,
	};

	// Note: webp doesn't have a dedicated asset in this kit; falls back to
	// generic image.png. If a webp icon is added later, swap the line above.

	const categoryIconMap: Record<string, string> = {
		folder: folderIcon,
		video: videoIcon,
		audio: audioIcon,
		image: imageIcon,
		document: textIcon,
		archive: archiveIcon,
		other: unknownIcon,
	};

	const iconSrc = $derived.by(() => {
		if (isDir) return folderIcon;

		// 1. Exact extension match wins (most specific)
		const extMatch = extIconMap[ext];
		if (extMatch) return extMatch;

		// 2. Code-like files (json/ts/svelte/etc.)
		if (isCode) return codeIcon;

		// 3. Category from backend (preferred over raw mime sniffing)
		if (category && categoryIconMap[category]) {
			return categoryIconMap[category];
		}

		// 4. Raw mime fallback
		if (mt.startsWith('video/')) return videoIcon;
		if (mt.startsWith('audio/')) return audioIcon;
		if (mt.startsWith('image/')) return imageIcon;
		if (mt === 'application/pdf') return pdfIcon;
		if (mt.startsWith('text/')) return textIcon;
		if (
			mt.includes('zip') ||
			mt.includes('rar') ||
			mt.includes('tar') ||
			mt.includes('gzip') ||
			mt.includes('7z')
		) {
			return archiveIcon;
		}

		return unknownIcon;
	});
</script>

<img
	src={iconSrc}
	alt=""
	width={size}
	height={size}
	draggable="false"
	class="inline-block shrink-0 select-none object-contain {className}"
	style="width: {size}px; height: {size}px;"
/>
