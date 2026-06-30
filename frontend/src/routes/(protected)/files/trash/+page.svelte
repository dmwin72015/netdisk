<script lang="ts">
 import { onMount } from "svelte";
 import { goto } from "$app/navigation";
 import { user, authReady } from "$lib/stores/auth";
 import {
 listTrashed,
 restoreFile,
 permanentDelete,
 emptyTrash,
 restoreAll,
 type FileItem,
 } from "$lib/api/files";
  import { Trash2, RotateCcw, LoaderCircle, CircleAlert } from "@lucide/svelte";
 import { toast } from "svelte-sonner";
 import MimeIcon from "$lib/components/MimeIcon.svelte";
 import { confirmDelete, confirmAction } from "$lib/dialog";
  import * as m from "$lib/paraglide/messages";
  import emptyTrashSvg from "$lib/assets/empty-states/empty-trash.svg";
  import { fmtSize, fmtTime } from "$lib/utils/format";

 let files = $state<FileItem[]>([]);
 let total = $state(0);
 let loading = $state(true);

 async function refresh() {
 if (!$user) return;
 loading = true;
 try {
 const data = await listTrashed();
 files = data.files;
 total = data.total;
 } catch (e) {
 toast.error(e instanceof Error ? e.message : m.trash_load_failed());
 } finally {
 loading = false;
 }
 }

 async function restore(slug: string, name: string) {
 if (
 !(await confirmAction(
 m.restore(),
 m.confirm_restore({ name }),
 m.restore(),
 ))
 )
 return;
 try {
 await restoreFile(slug);
 await refresh();
 } catch (e) {
 toast.error(e instanceof Error ? e.message : m.restore_failed());
 }
 }

 async function permanent(slug: string, name: string) {
 if (!(await confirmDelete(m.confirm_permanent_delete({ name })))) return;
 try {
 await permanentDelete(slug);
 await refresh();
 } catch (e) {
 toast.error(e instanceof Error ? e.message : m.delete_failed());
 }
 }

 async function handleEmptyTrash() {
 if (!(await confirmDelete(m.confirm_empty_trash()))) return;
 try {
 const result = await emptyTrash();
 toast.success(m.trash_emptied());
 await refresh();
 } catch (e) {
 toast.error(e instanceof Error ? e.message : m.delete_failed());
 }
 }

 async function handleRestoreAll() {
 if (
 !(await confirmAction(
 m.restore_all(),
 m.confirm_restore_all(),
 m.restore_all(),
 ))
 )
 return;
 try {
 const result = await restoreAll();
 toast.success(m.all_restored());
 await refresh();
 } catch (e) {
 toast.error(e instanceof Error ? e.message : m.restore_failed());
 }
 }

 onMount(() => {
 if (!$user) void goto("/login");
 else void refresh();
 });
</script>

{#if !$authReady}{:else if $user}
 <div class="space-y-4 px-6 pt-4 pb-6">
 <div class="flex items-center justify-between">
 <div class="flex items-center gap-2">
 <Trash2 size={20} class="text-ink-3" />
 <h1 class="text-lg font-semibold text-ink">{m.trash_title()}</h1>
 <span class="text-sm text-ink-4"
 >{m.total_items({ total: String(total) })}</span
 >
 </div>

 {#if files.length > 0}
 <div class="flex items-center gap-2">
 <button
 type="button"
 onclick={handleRestoreAll}
 class="flex h-8 items-center gap-1.5 rounded-lg border border-line bg-surface px-3 text-sm text-ink-2 transition-colors hover:border-line hover:bg-surface-sunken"
 >
 <RotateCcw size={14} />
 {m.restore_all()}
 </button>
 <button
 type="button"
 onclick={handleEmptyTrash}
 class="flex h-8 items-center gap-1.5 rounded-lg border border-danger bg-surface px-3 text-sm text-danger transition-colors hover:border-danger hover:bg-danger-soft"
 >
 <Trash2 size={14} />
 {m.empty_trash()}
 </button>
 </div>
 {/if}
 </div>

 <div class="flex items-center gap-2 rounded-lg border border-warning bg-warning-soft px-3.5 py-2.5 text-sm text-warning">
 <CircleAlert size={16} class="shrink-0" />
 <span>{m.trash_retention_hint()}</span>
 </div>

 {#if loading}
 <div class="flex items-center justify-center py-16">
 <LoaderCircle size={24} class="animate-spin text-ink-4" />
 </div>
 {:else if files.length === 0}
 <div
 class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center"
 >
  <img src={emptyTrashSvg} class="mb-2 w-32 h-32" alt="" />
  <p class="text-sm text-ink-4">{m.trash_empty()}</p>
 </div>
 {:else}
 <div
 class="overflow-hidden rounded-xl border border-line-soft bg-surface "
 >
 <table class="w-full table-fixed text-sm">
 <thead>
 <tr
 class="border-b border-line-soft text-left text-xs text-ink-4"
 >
 <th class="w-[50%] px-4 py-2.5 font-medium">{m.col_filename()}</th
 >
 <th class="w-[15%] px-4 py-2.5 text-right font-medium"
 >{m.col_size()}</th
 >
 <th class="w-[15%] px-4 py-2.5 text-right font-medium"
 >{m.col_deleted()}</th
 >
 <th class="w-[20%] px-4 py-2.5 text-right font-medium"
 >{m.col_actions()}</th
 >
 </tr>
 </thead>
 <tbody>
 {#each files as f (f.slug)}
 <tr
 class="border-b border-line-soft transition-colors last:border-0 hover:bg-surface-sunken/80"
 >
 <td class="px-4 py-2.5">
 <div class="flex items-center gap-2.5">
 <span class="shrink-0"
 ><MimeIcon
 mimeType={f.mimeType}
 name={f.fileName}
 isDir={f.isDir}
 size={18}
 /></span
 >
 <span class="truncate text-ink-2" title={f.fileName}
 >{f.fileName}</span
 >
 </div>
 </td>
 <td class="px-4 py-2.5 text-right text-ink-3"
 >{f.isDir ? "-" : fmtSize(f.fileSize)}</td
 >
 <td
 class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-ink-4"
 >
 {fmtTime(f.updatedAt)}
 </td>
 <td class="px-4 py-2.5 text-right">
 <div class="flex items-center justify-end gap-1">
 <button
 type="button"
 onclick={() => restore(f.slug, f.fileName)}
 class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-success"
 title={m.upload_resume()}
 >
 <RotateCcw size={15} />
 </button>
 <button
 type="button"
 onclick={() => permanent(f.slug, f.fileName)}
 class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-danger-soft hover:text-danger"
 title={m.permanent_delete()}
 >
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
{/if}
