<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { MoreHorizontal, Star, Eye, Download, Pencil, Trash2, Film } from '@lucide/svelte';
	import { downloadUrl, type FileItem } from '$lib/api/files';
	import * as m from '$lib/paraglide/messages';

	let {
		file,
		onStar,
		onPreview,
		onRename,
		onDelete,
		onAddToMedia,
		triggerClass = ''
	}: {
		file: FileItem;
		onStar: (slug: string, starred: boolean) => void;
		onPreview: (file: FileItem) => void;
		onRename: (slug: string, name: string) => void;
		onDelete: (slug: string, name: string) => void;
		onAddToMedia?: (file: FileItem) => void;
		triggerClass?: string;
	} = $props();

	let isVideo = $derived(file.mime_type?.startsWith('video/') ?? false);
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 {triggerClass}"
	>
		<MoreHorizontal size={16} />
	</DropdownMenu.Trigger>
	<DropdownMenu.Portal>
		<DropdownMenu.Content
			class="z-50 min-w-35 rounded-xl border border-gray-100 bg-white p-1.5 shadow-lg"
			sideOffset={4}
		>
			<DropdownMenu.Item
				class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
				onclick={() => onStar(file.slug, file.is_starred)}
			>
				<Star size={14} class={file.is_starred ? 'text-amber-400' : 'text-gray-400'} fill={file.is_starred ? 'currentColor' : 'none'} />
				{file.is_starred ? m.unstar_file() : m.star_file()}
			</DropdownMenu.Item>
			{#if !file.is_dir}
				<DropdownMenu.Item
					class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
					onclick={() => onPreview(file)}
				>
					<Eye size={14} class="text-gray-400" />
					{m.preview()}
				</DropdownMenu.Item>
				<DropdownMenu.Item
					class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
				>
					<a href={downloadUrl(file.slug)} download={file.file_name} class="flex items-center gap-2.5">
						<Download size={14} class="text-gray-400" />
						{m.download()}
					</a>
				</DropdownMenu.Item>
			{/if}
			{#if isVideo && onAddToMedia}
				<DropdownMenu.Item
					class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
					onclick={() => onAddToMedia(file)}
				>
					<Film size={14} class="text-gray-400" />
					{m.add_to_media_library()}
				</DropdownMenu.Item>
			{/if}
			<DropdownMenu.Separator class="my-1 h-px bg-gray-100" />
			<DropdownMenu.Item
				class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none transition-colors hover:bg-gray-50 focus:bg-gray-50"
				onclick={() => onRename(file.slug, file.file_name)}
			>
				<Pencil size={14} class="text-gray-400" />
				{m.rename()}
			</DropdownMenu.Item>
			<DropdownMenu.Item
				class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-red-600 outline-none transition-colors hover:bg-red-50 focus:bg-red-50"
				onclick={() => onDelete(file.slug, file.file_name)}
			>
				<Trash2 size={14} />
				{m.delete_label()}
			</DropdownMenu.Item>
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>
