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

	function onPreviewComplete(open: boolean) {
		if (!open) previewFile = null;
	}

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

	let dialogOpen = $derived(!!previewFile);

</script>

{#if $authReady && $user}
	<div class="space-y-4">
		<div class="flex items-center gap-2">
			<Star size={20} class="text-warning" fill="currentColor" />
			<h1 class="text-lg font-semibold text-ink">{m.starred_title()}</h1>
			<span class="text-sm text-ink-4">{m.total_items({ total: String(total) })}</span>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<LoaderCircle size={24} class="animate-spin text-ink-4" />
			</div>
		{:else if files.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center">
				<FolderPlus size={40} class="mb-3 text-ink-4" />
				<p class="text-sm text-ink-4">{m.starred_empty()}</p>
			</div>
		{:else}
			<div class="overflow-hidden rounded-xl border border-line-soft bg-white ">
				<table class="w-full table-fixed text-sm">
					<thead>
						<tr class="border-b border-line-soft text-left text-xs text-ink-4">
							<th class="w-[45%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
							<th class="w-[15%] px-4 py-2.5 font-medium">{m.col_type()}</th>
							<th class="w-[10%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_modified()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody>
						{#each files as f (f.slug)}
							<tr class="border-b border-line-soft transition-colors last:border-0 hover:bg-surface-muted/80">
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-2.5">
										<span class="shrink-0"><MimeIcon mimeType={f.mimeType} name={f.fileName} isDir={f.isDir} size={18} /></span>
										<span class="truncate text-ink-2" title={f.fileName}>{f.fileName}</span>
									</div>
								</td>
								<td class="truncate px-4 py-2.5 text-xs text-ink-4">{f.isDir ? m.directory() : f.mimeType}</td>
								<td class="px-4 py-2.5 text-right text-ink-3">{f.isDir ? '-' : fmtSize(f.fileSize)}</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-ink-4">
									{fmtTime(f.updatedAt)}
								</td>
								<td class="px-4 py-2.5 text-right">
									<div class="flex items-center justify-end">
										<button type="button" onclick={() => unstar(f.slug)} class="rounded-md p-1.5 text-warning transition-colors hover:bg-warning-soft" title={m.remove_star()}>
											<Star size={15} fill="currentColor" />
										</button>
										{#if !f.isDir}
											<button type="button" onclick={() => (previewFile = { slug: f.slug, name: f.fileName, mimeType: f.mimeType || '', size: f.fileSize })} class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3">
												<Eye size={15} />
											</button>
											<button type="button" onclick={() => { const url = authedUrl(downloadUrl(f.slug)); const a = document.createElement('a'); a.href = url; a.download = f.fileName; a.click(); a.remove(); }} class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3">
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

	<DrivePreview
		id={previewFile!.slug}
		name={previewFile!.name}
		mimeType={previewFile!.mimeType}
		size={previewFile!.size}
		bind:open={dialogOpen}
		onOpenChangeComplete={onPreviewComplete}
	/>
