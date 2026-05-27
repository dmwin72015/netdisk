<script lang="ts">
	import { Loader2, FolderPlus, Star } from '@lucide/svelte';
	import { type FileItem } from '$lib/api/files';
	import { fmtSize, fmtTime } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import FileActionsDropdown from './FileActionsDropdown.svelte';

	let {
		files,
		viewMode,
		loading,
		emptyMessage,
		onNavigateDir,
		onStar,
		onPreview,
		onRename,
		onDelete,
		onAddToMedia
	}: {
		files: FileItem[];
		viewMode: 'list' | 'grid';
		loading: boolean;
		emptyMessage: string;
		onNavigateDir: (slug: string) => void;
		onStar: (slug: string, starred: boolean) => void;
		onPreview: (file: FileItem) => void;
		onRename: (slug: string, name: string) => void;
		onDelete: (slug: string, name: string) => void;
		onAddToMedia?: (file: FileItem) => void;
	} = $props();

	function canPreview(mimeType: string | null): boolean {
		if (!mimeType) return false;
		return (
			mimeType.startsWith('image/') ||
			mimeType.startsWith('video/') ||
			mimeType.startsWith('audio/') ||
			mimeType === 'application/pdf' ||
			mimeType.startsWith('text/')
		);
	}
</script>

{#if loading && files.length === 0}
	<div class="flex items-center justify-center py-16">
		<Loader2 size={24} class="animate-spin text-gray-300" />
	</div>
{:else if files.length === 0}
	<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
		<FolderPlus size={40} class="mb-3 text-gray-300" />
		<p class="text-sm text-gray-400">{emptyMessage}</p>
	</div>
{:else if viewMode === 'grid'}
	<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
		{#each files as f (f.slug)}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class="group relative flex flex-col items-center rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-200 hover:shadow-md {f.is_dir || canPreview(f.mime_type) ? 'cursor-pointer' : ''}"
				onclick={f.is_dir ? () => onNavigateDir(f.slug) : (!f.is_dir && canPreview(f.mime_type)) ? () => onPreview(f) : undefined}
			>
				<div class="absolute right-1.5 top-1.5" onclick={(e) => e.stopPropagation()}>
					<FileActionsDropdown
						file={f}
						{onStar}
						{onPreview}
						{onRename}
						{onDelete}
						{onAddToMedia}
						triggerClass="rounded-lg bg-white/90 p-1 text-gray-400 shadow-sm backdrop-blur transition-colors hover:bg-white hover:text-gray-600"
					/>
				</div>
				<MimeIcon mimeType={f.mime_type} isDir={f.is_dir} size={36} />
				<p class="mt-3 w-full truncate text-center text-sm font-medium text-gray-700" title={f.file_name}>{f.file_name}</p>
				<p class="mt-0.5 text-xs text-gray-400">{f.is_dir ? '' : fmtSize(f.file_size)}</p>
			</div>
		{/each}
	</div>
{:else}
	<div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
		<table class="w-full table-fixed text-sm">
			<thead>
				<tr class="border-b border-gray-100 text-left text-xs text-gray-400">
					<th class="w-[50%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
					<th class="w-[15%] px-4 py-2.5 font-medium">{m.col_type()}</th>
					<th class="w-[10%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
					<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_modified()}</th>
					<th class="w-[10%] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
				</tr>
			</thead>
			<tbody>
				{#each files as f (f.slug)}
					<tr class="group border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80 {f.is_dir || canPreview(f.mime_type) ? 'cursor-pointer' : ''}" onclick={f.is_dir ? () => onNavigateDir(f.slug) : canPreview(f.mime_type) ? () => onPreview(f) : undefined}>
						<td class="px-4 py-2.5">
							<div class="flex items-center gap-2.5">
								<span class="shrink-0"><MimeIcon mimeType={f.mime_type} isDir={f.is_dir} size={18} /></span>
								<span class="truncate text-gray-700" title={f.file_name}>{f.file_name}</span>
								{#if f.is_starred}
									<Star size={12} class="shrink-0 text-amber-400" fill="currentColor" />
								{/if}
							</div>
						</td>
						<td class="truncate px-4 py-2.5 text-xs text-gray-400">{f.is_dir ? m.directory() : f.mime_type}</td>
						<td class="px-4 py-2.5 text-right text-gray-500">{f.is_dir ? '-' : fmtSize(f.file_size)}</td>
						<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400">
							{fmtTime(f.updated_at)}
						</td>
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<td class="px-4 py-2.5 text-right" onclick={(e) => e.stopPropagation()}>
							<div class="flex items-center justify-end">
								<FileActionsDropdown
									file={f}
									{onStar}
									{onPreview}
									{onRename}
									{onDelete}
									{onAddToMedia}
								/>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
