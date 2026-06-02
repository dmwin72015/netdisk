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
  import { Trash2, RotateCcw, LoaderCircle, FolderPlus, CircleAlert } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { confirmDelete, confirmAction } from "$lib/dialog";
  import * as m from "$lib/paraglide/messages";
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
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Trash2 size={20} class="text-gray-500" />
        <h1 class="text-lg font-semibold text-gray-900">{m.trash_title()}</h1>
        <span class="text-sm text-gray-400"
          >{m.total_items({ total: String(total) })}</span
        >
      </div>

      {#if files.length > 0}
        <div class="flex items-center gap-2">
          <button
            type="button"
            onclick={handleRestoreAll}
            class="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50"
          >
            <RotateCcw size={14} />
            {m.restore_all()}
          </button>
          <button
            type="button"
            onclick={handleEmptyTrash}
            class="flex h-8 items-center gap-1.5 rounded-lg border border-red-200 bg-white px-3 text-sm text-red-600 transition-colors hover:border-red-300 hover:bg-red-50"
          >
            <Trash2 size={14} />
            {m.empty_trash()}
          </button>
        </div>
      {/if}
    </div>

    <div class="flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-sm text-amber-700">
      <CircleAlert size={16} class="shrink-0" />
      <span>{m.trash_retention_hint()}</span>
    </div>

    {#if loading}
      <div class="flex items-center justify-center py-16">
        <LoaderCircle size={24} class="animate-spin text-gray-300" />
      </div>
    {:else if files.length === 0}
      <div
        class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center"
      >
        <FolderPlus size={40} class="mb-3 text-gray-300" />
        <p class="text-sm text-gray-400">{m.trash_empty()}</p>
      </div>
    {:else}
      <div
        class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm"
      >
        <table class="w-full table-fixed text-sm">
          <thead>
            <tr
              class="border-b border-gray-100 text-left text-xs text-gray-400"
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
                class="border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80"
              >
                <td class="px-4 py-2.5">
                  <div class="flex items-center gap-2.5">
                    <span class="shrink-0"
                      ><MimeIcon
                        mimeType={f.mimeType}
                        isDir={f.isDir}
                        size={18}
                      /></span
                    >
                    <span class="truncate text-gray-700" title={f.fileName}
                      >{f.fileName}</span
                    >
                  </div>
                </td>
                <td class="px-4 py-2.5 text-right text-gray-500"
                  >{f.isDir ? "-" : fmtSize(f.fileSize)}</td
                >
                <td
                  class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400"
                >
                  {fmtTime(f.updatedAt)}
                </td>
                <td class="px-4 py-2.5 text-right">
                  <div class="flex items-center justify-end gap-1">
                    <button
                      type="button"
                      onclick={() => restore(f.slug, f.fileName)}
                      class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-green-600"
                      title={m.upload_resume()}
                    >
                      <RotateCcw size={15} />
                    </button>
                    <button
                      type="button"
                      onclick={() => permanent(f.slug, f.fileName)}
                      class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500"
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
