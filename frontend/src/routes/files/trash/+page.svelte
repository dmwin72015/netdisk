<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, authReady } from '$lib/stores/auth';
	import { listTrashed, restoreFile, permanentDelete, type FileItem } from '$lib/api/files';
	import { Trash2, RotateCcw, Loader2, FolderPlus } from '@lucide/svelte';
	import MimeIcon from '$lib/components/MimeIcon.svelte';
	import { confirmDelete, confirmAction } from '$lib/dialog';
	import * as m from '$lib/paraglide/messages';
	import { fmtSize, fmtTime } from '$lib/utils/format';

	let files = $state<FileItem[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let error = $state<string | null>(null);

	async function refresh() {
		if (!$user) return;
		loading = true;
		error = null;
		try {
			const data = await listTrashed();
			files = data.files;
			total = data.total;
		} catch (e) {
			error = e instanceof Error ? e.message : m.trash_load_failed();
		} finally {
			loading = false;
		}
	}

	async function restore(slug: string, name: string) {
		if (!(await confirmAction(m.restore(), m.confirm_restore({ name }), m.restore()))) return;
		try {
			await restoreFile(slug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.restore_failed();
		}
	}

	async function permanent(slug: string, name: string) {
		if (!(await confirmDelete(m.confirm_permanent_delete({ name })))) return;
		try {
			await permanentDelete(slug);
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : m.delete_failed();
		}
	}

	onMount(() => {
		if (!$user) void goto('/login');
		else void refresh();
	});
</script>

{#if !$authReady}
{:else if $user}
	<div class="space-y-4">
		<div class="flex items-center gap-2">
			<Trash2 size={20} class="text-gray-500" />
			<h1 class="text-lg font-semibold text-gray-900">{m.trash_title()}</h1>
			<span class="text-sm text-gray-400">{m.total_items({ total: String(total) })}</span>
		</div>

		{#if error}
			<div class="rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700">{error}</div>
		{/if}

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<Loader2 size={24} class="animate-spin text-gray-300" />
			</div>
		{:else if files.length === 0}
			<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center">
				<FolderPlus size={40} class="mb-3 text-gray-300" />
				<p class="text-sm text-gray-400">{m.trash_empty()}</p>
			</div>
		{:else}
			<div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
				<table class="w-full table-fixed text-sm">
					<thead>
						<tr class="border-b border-gray-100 text-left text-xs text-gray-400">
							<th class="w-[50%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
							<th class="w-[15%] px-4 py-2.5 text-right font-medium">{m.col_deleted()}</th>
							<th class="w-[20%] px-4 py-2.5 text-right font-medium">{m.col_actions()}</th>
						</tr>
					</thead>
					<tbody>
						{#each files as f (f.slug)}
							<tr class="border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80">
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-2.5">
										<span class="shrink-0"><MimeIcon mimeType={f.mime_type} isDir={f.is_dir} size={18} /></span>
										<span class="truncate text-gray-700" title={f.file_name}>{f.file_name}</span>
									</div>
								</td>
								<td class="px-4 py-2.5 text-right text-gray-500">{f.is_dir ? '-' : fmtSize(f.file_size)}</td>
								<td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400">
									{fmtTime(f.updated_at)}
								</td>
								<td class="px-4 py-2.5 text-right">
									<div class="flex items-center justify-end gap-1">
										<button type="button" onclick={() => restore(f.slug, f.file_name)} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-green-600" title={m.upload_resume()}>
											<RotateCcw size={15} />
										</button>
										<button type="button" onclick={() => permanent(f.slug, f.file_name)} class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500" title={m.permanent_delete()}>
											<Trash2 size={15} />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
{:else}
	<p class="text-gray-600">{@html m.please_login({ link: `<a href="/login" class="text-blue-600 underline hover:text-blue-700">${m.login_link_text()}</a>` })}</p>
{/if}
