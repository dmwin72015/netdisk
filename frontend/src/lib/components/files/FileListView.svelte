<script lang="ts">
  import { LoaderCircle, FolderPlus, Star } from "@lucide/svelte";
  import { fade } from "svelte/transition";
  import type { NormalizedFile } from "$lib/types/file";
  import { fmtSize, fmtTime } from "$lib/utils/format";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import FileActionsDropdown from "./FileActionsDropdown.svelte";

  let {
    files,
    viewMode,
    loading,
    directoryId = '',
    emptyMessage,
    downloadUrlFn,
    onNavigateDir,
    onStar,
    onPreview,
    onRename,
    onDelete,
    onAddToMedia,
  }: {
    files: NormalizedFile[];
    viewMode: "list" | "grid";
    loading: boolean;
    directoryId?: string;
    emptyMessage: string;
    downloadUrlFn: (id: string) => string;
    onNavigateDir: (id: string) => void;
    onStar?: (id: string, starred: boolean) => void;
    onPreview: (file: NormalizedFile) => void;
    onRename: (id: string, name: string) => void;
    onDelete: (id: string, name: string) => void;
    onAddToMedia?: (file: NormalizedFile) => void;
  } = $props();

  // directoryId 变化时（切换目录）才触发容器淡入动画
  let listKey = $state(0);
  let prevDirId = $state('');

  $effect(() => {
    if (directoryId !== prevDirId) {
      prevDirId = directoryId;
      listKey++;
    }
  });

  function canPreview(mimeType: string | null): boolean {
    if (!mimeType) return false;
    return (
      mimeType.startsWith("image/") ||
      mimeType.startsWith("video/") ||
      mimeType.startsWith("audio/") ||
      mimeType === "application/pdf" ||
      mimeType.startsWith("text/")
    );
  }

  function categoryLabel(cat: string): string {
    const map: Record<string, string> = {
      video: m.category_video(),
      audio: m.category_audio(),
      image: m.category_image(),
      document: m.category_document(),
      archive: m.category_archive(),
      other: m.category_other(),
    };
    return map[cat] ?? cat;
  }
</script>

<!-- Loading overlay - shown on top when loading -->
{#if loading}
  <div
    class="flex items-center justify-center py-16 absolute left-0 right-0 mx-auto"
    transition:fade={{ duration: 150 }}
  >
    <LoaderCircle size={24} class="animate-spin text-gray-300" />
  </div>
{:else if files.length === 0}
  <div
    class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16 text-center"
    in:fade={{ duration: 150 }}
  >
    <FolderPlus size={40} class="mb-3 text-gray-300" />
    <p class="text-sm text-gray-400">{emptyMessage}</p>
  </div>
{:else if viewMode === "grid"}
  <!-- key 变化时整个网格淡入，覆盖切换目录场景 -->
  {#key listKey}
    <div
      class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
      in:fade={{ duration: 180 }}
    >
      {#each files as f, i (f.id)}
        <div
          class="group relative flex flex-col items-center rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:border-gray-200 hover:shadow-md {f.isDir ||
          canPreview(f.mimeType)
            ? 'cursor-pointer'
            : ''}"
          onclick={f.isDir
            ? () => onNavigateDir(f.id)
            : !f.isDir && canPreview(f.mimeType)
              ? () => onPreview(f)
              : undefined}
        >
          <div
            class="absolute right-1.5 top-1.5"
            onclick={(e) => e.stopPropagation()}
          >
            <FileActionsDropdown
              file={f}
              {downloadUrlFn}
              {onStar}
              {onPreview}
              {onRename}
              {onDelete}
              {onAddToMedia}
              triggerClass="rounded-lg bg-white/90 p-1 text-gray-400 shadow-sm backdrop-blur transition-colors hover:bg-white hover:text-gray-600"
            />
          </div>
          <MimeIcon
            mimeType={f.mimeType}
            isDir={f.isDir}
            category={f.fileCategory}
            size={36}
          />
          <p
            class="mt-3 w-full truncate text-center text-sm font-medium text-gray-700"
            title={f.name}
          >
            {f.name}
          </p>
          <p class="mt-0.5 text-xs text-gray-400">
            {f.isDir ? "" : fmtSize(f.size)}
          </p>
        </div>
      {/each}
    </div>
  {/key}
{:else}
  <!-- key 变化时整个表格淡入，覆盖切换目录场景 -->
  {#key listKey}
    <div
      class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm"
      in:fade={{ duration: 180 }}
    >
      <table class="w-full table-fixed text-sm">
        <thead>
          <tr class="border-b border-gray-100 text-left text-xs text-gray-400">
            <th class="w-[50%] px-4 py-2.5 font-medium">{m.col_filename()}</th>
            <th class="w-[15%] px-4 py-2.5 font-medium">{m.col_type()}</th>
            <th class="w-[10%] px-4 py-2.5 text-right font-medium"
              >{m.col_size()}</th
            >
            <th class="w-[15%] px-4 py-2.5 text-right font-medium"
              >{m.col_modified()}</th
            >
            <th class="w-[10%] px-4 py-2.5 text-right font-medium"
              >{m.col_actions()}</th
            >
          </tr>
        </thead>
        <tbody>
          {#each files as f, i (f.id)}
            <tr
              class="group border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80 {f.isDir ||
              canPreview(f.mimeType)
                ? 'cursor-pointer'
                : ''}"
              onclick={f.isDir
                ? () => onNavigateDir(f.id)
                : canPreview(f.mimeType)
                  ? () => onPreview(f)
                  : undefined}
            >
              <td class="px-4 py-2.5">
                <div class="flex items-center gap-2.5">
                  <span class="shrink-0">
                    <MimeIcon
                      mimeType={f.mimeType}
                      isDir={f.isDir}
                      category={f.fileCategory}
                      size={18}
                    />
                  </span>
                  <span class="truncate text-gray-700" title={f.name}
                    >{f.name}</span
                  >
                  {#if f.isStarred}
                    <Star
                      size={12}
                      class="shrink-0 text-amber-400"
                      fill="currentColor"
                    />
                  {/if}
                </div>
              </td>
              <td class="truncate px-4 py-2.5 text-xs text-gray-400">
                {f.isDir ? m.directory() : categoryLabel(f.fileCategory)}
              </td>
              <td class="px-4 py-2.5 text-right text-gray-500">
                {f.isDir ? "-" : fmtSize(f.size)}
              </td>
              <td
                class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-400"
              >
                {fmtTime(f.updatedAt)}
              </td>
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <td
                class="px-4 py-2.5 text-right"
                onclick={(e) => e.stopPropagation()}
              >
                <div class="flex items-center justify-end">
                  <FileActionsDropdown
                    file={f}
                    {downloadUrlFn}
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
  {/key}
{/if}
