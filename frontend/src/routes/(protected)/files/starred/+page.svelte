<script lang="ts">
	import { onMount } from 'svelte';
	import { user, authReady } from '$lib/stores/auth';
	import { listStarred, setStarred, downloadUrl, type FileItem } from '$lib/api/files';
	import { Star, Download, Eye, LoaderCircle, FolderPlus } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize, fmtTime, authedUrl } from '$lib/utils/format';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import DrivePreview from '$lib/components/DrivePreview.svelte';

	let files = $state<FileItem[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let previewFile = $state<{ slug: string; name: string; mimeType: string; size: number } | null>(null);

	async function refresh() {
		if (!$user) return;
		loading = true;
		try {
			const data = await listStarred();
			files = data.files;
			total = data.total;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.load_failed());
		} finally {
			loading = false;
		}
	}

	async function unstar(slug: string) {
		try {
			await setStarred(slug, false);
			files = files.filter(f => f.slug !== slug);
			total--;
		} catch (e) {
			toast.error(e instanceof Error ? e.message : m.unstar_failed());
		}
	}

	onMount(() => {
		void refresh();
	});
</script>

{#if $authReady && $user}
	<div class="space-y-4">
		<div class="flex items-center gap-2">
			<Star size={20} class="text-amber-400" fill="currentColor" />
			<h1 class="text-lg font-semibold text-gray-900">{m.starred_title()}</h1>
			<span class="text-sm text-gray-400">{m.total_items({ total: String(total) })}</span>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if files.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
				<FolderPlus size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.starred_empty()}</p>
			</div>
		{:else}
			<div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
				<table class="w-full table-fixed text-sm">
					<thead>
						<tr class="border-b border-gray-100 text-left text-xs text-gray-400">
							<th class="w-[45%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
							<th class="w-[15%] px-4 py-2.5 font-medium">{m.col_type()}</th>
							<th class="w-[10%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_modified()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody>
						{#each files as f (f.slug)}
							<tr class="border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80">
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-2.5">
										<span class="shrink-0"><MimeIcon mimeType={f.mimeType} name={f.fileName} isDir={f.isDir} size={18} /></span>
										<span class="truncate text-gray-700" title={f.fileName}>{f.fileName}</span>
									</div>
								</td>
								<td class="truncate px-4 py-2.5 text-xs text-gray-400">{f.isDir ? m.directory() : f.mimeType}</td>
								<td class="px-4 py-2.5 text-right text-gray-500">{f.isDir ? '-' : fmtSize(f.fileSize)}</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400">
									{fmtTime(f.updatedAt)}
								</td>
								<td class="px-4 py-2.5 text-right">
									<div class="flex items-center justify-end">
										<button type="button" onclick={() => unstar(f.slug)} class="rounded-md p-1.5 text-amber-400 transition-colors hover:bg-amber-50" title={m.remove_star()}>
											<Star size={15} fill="currentColor" />
										</button>
										{#if !f.isDir}
											<button type="button" onclick={() => (previewFile = { slug: f.slug, name: f.fileName, mimeType: f.mimeType || '', size: f.fileSize })} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
												<Eye size={15} />
											</button>
											<button type="button" onclick={() => { const url = authedUrl(downloadUrl(f.slug)); const a = document.createElement('a'); a.href = url; a.download = f.fileName; a.click(); a.remove(); }} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
												<Download size={15} />
											</button>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
{/if}

{#if previewFile}
	<DrivePreview
		id={previewFile.slug}
		name={previewFile.name}
		mimeType={previewFile.mimeType}
		size={previewFile.size}
		open={true}
		close={() => (previewFile = null)}
	/>
{/if}
