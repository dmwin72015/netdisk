<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { FileText, ChevronLeft, ChevronRight, LoaderCircle, Trash2, Star } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { adminListFiles, type AdminFile } from '$lib/api/admin';
	import { fmtSize, timeAgo } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	const PAGE_SIZE = 20;

	let files = $state<AdminFile[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let loading = $state(true);

	let currentPage = $derived(Math.floor(offset / PAGE_SIZE) + 1);
	let totalPages = $derived(Math.ceil(total / PAGE_SIZE));

	onMount(() => {
		if (!browser) return;
		loadFiles();
	});

	async function loadFiles() {
		loading = true;
		try {
			const res = await adminListFiles(PAGE_SIZE, offset);
			files = res.items;
			total = res.total;
		} catch {
			toast.error(m.load_failed());
		} finally {
			loading = false;
		}
	}

	function goPage(page: number) {
		offset = (page - 1) * PAGE_SIZE;
		loadFiles();
	}

	function fmtDate(ts: number): string {
		return new Date(ts * 1000).toLocaleString();
	}

	function dirIcon(file: AdminFile): string {
		if (file.isDir) return '📁';
		switch (file.fileCategory) {
			case 'video': return '🎬';
			case 'audio': return '🎵';
			case 'image': return '🖼️';
			case 'document': return '📄';
			case 'archive': return '📦';
			default: return '📄';
		}
	}
</script>

<div class="space-y-4">
	<div class="flex items-center gap-2">
		<FileText size={20} class="text-slate-600" />
		<h1 class="text-xl font-semibold">All Files</h1>
		<span class="ml-auto text-sm text-slate-400">{m.total_items({ total: String(total) })}</span>
	</div>

	<div class="overflow-hidden rounded-lg border bg-white">
		<table class="w-full text-left text-sm">
			<thead class="border-b bg-slate-50 text-xs text-slate-500">
				<tr>
					<th class="px-4 py-3 font-medium">{m.col_filename()}</th>
					<th class="px-4 py-3 font-medium">{m.username()}</th>
					<th class="px-4 py-3 font-medium">{m.col_type()}</th>
					<th class="px-4 py-3 font-medium">{m.col_size()}</th>
					<th class="px-4 py-3 font-medium">{m.col_upload_time()}</th>
					<th class="px-4 py-3 font-medium">Status</th>
				</tr>
			</thead>
			<tbody class="divide-y">
				{#if loading}
					<tr>
						<td colspan="6" class="px-4 py-10 text-center text-slate-400">
							<LoaderCircle size={20} class="mx-auto animate-spin" />
						</td>
					</tr>
				{:else if files.length === 0}
					<tr>
						<td colspan="6" class="px-4 py-10 text-center text-slate-400">{m.no_files()}</td>
					</tr>
				{:else}
					{#each files as f (f.id)}
						<tr class="hover:bg-slate-50 transition-colors">
							<td class="px-4 py-3">
								<div class="flex items-center gap-2">
									<span class="text-base">{dirIcon(f)}</span>
									<span class="font-medium text-slate-900">{f.fileName}</span>
								</div>
							</td>
							<td class="px-4 py-3">
								<button
									class="text-blue-600 hover:underline"
									onclick={() => goto(`/admin/users/${f.userId}`)}
								>
									{f.username}
								</button>
							</td>
							<td class="px-4 py-3 text-slate-500">
								{f.isDir ? m.directory() : (f.mimeType || f.fileCategory || '-')}
							</td>
							<td class="px-4 py-3 font-mono text-xs text-slate-600">
								{f.isDir ? '-' : fmtSize(f.fileSize)}
							</td>
							<td class="px-4 py-3 text-slate-500 text-xs">{fmtDate(f.createdAt)}</td>
							<td class="px-4 py-3">
								<div class="flex items-center gap-2">
									{#if f.isTrashed}
										<span class="rounded bg-red-50 px-2 py-0.5 text-xs text-red-600">Trashed</span>
									{/if}
									{#if f.isStarred}
										<Star size={13} class="text-amber-400" />
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2">
			<button
				class="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-slate-50 disabled:opacity-40"
				disabled={currentPage <= 1}
				onclick={() => goPage(currentPage - 1)}
			>
				<ChevronLeft size={14} />
			</button>
			<span class="text-sm text-slate-500">{currentPage} / {totalPages}</span>
			<button
				class="rounded-lg border px-3 py-1.5 text-sm transition-colors hover:bg-slate-50 disabled:opacity-40"
				disabled={currentPage >= totalPages}
				onclick={() => goPage(currentPage + 1)}
			>
				<ChevronRight size={14} />
			</button>
		</div>
	{/if}
</div>
