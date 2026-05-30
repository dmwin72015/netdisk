<script lang="ts">
  import { LoaderCircle, FolderPlus, Star, Check, Download, Trash2, X } from "@lucide/svelte";
  import { fade, fly } from "svelte/transition";
  import type { NormalizedFile } from "$lib/types/file";
  import { fmtSize, fmtTime } from "$lib/utils/format";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { Tooltip } from "$lib/ui/tooltip";
  import { Dialog } from "$lib/ui/dialog";
  import { getAccessToken } from "$lib/api/client";
  import { toast } from "svelte-sonner";
  import FileActionsDropdown from "./FileActionsDropdown.svelte";

  type FolderSummary = {
    fileCount: number;
    folderCount: number;
    size: number;
  };

  let {
    files,
    viewMode,
    loading,
    directoryId = '',
    currentPath = [],
    emptyMessage,
    downloadUrlFn,
    onNavigateDir,
    onStar,
    onPreview,
    onRename,
    onDelete,
    onAddToMedia,
    loadFolderSummary,
  }: {
    files: NormalizedFile[];
    viewMode: "list" | "grid";
    loading: boolean;
    directoryId?: string;
    currentPath?: { id: string; name: string }[];
    emptyMessage: string;
    downloadUrlFn: (id: string) => string;
    onNavigateDir: (id: string) => void;
    onStar?: (id: string, starred: boolean) => void;
    onPreview: (file: NormalizedFile) => void;
    onRename: (id: string, name: string) => void;
    onDelete: (id: string, name: string) => void;
    onAddToMedia?: (file: NormalizedFile) => void;
    loadFolderSummary?: (id: string) => Promise<FolderSummary>;
  } = $props();

  // 切换目录或数据加载完成时重建容器，触发 transition:fade
  let fadeKey = $state(0);
  let selected = $state<Set<string>>(new Set());
  let detailFile = $state<NormalizedFile | null>(null);
  let detailOpen = $state(false);
  let detailSummary = $state<FolderSummary | null>(null);
  let detailSummaryLoading = $state(false);
  let allSelected = $derived(files.length > 0 && files.every(f => selected.has(f.id)));
  let hasSelection = $derived(selected.size > 0);

  function toggleSelect(id: string) {
    if (selected.has(id)) selected.delete(id);
    else selected.add(id);
    selected = new Set(selected);
  }

  function toggleSelectAll() {
    if (allSelected) selected = new Set();
    else selected = new Set(files.map(f => f.id));
  }

  function showDetails(file: NormalizedFile) {
    detailFile = file;
    detailOpen = true;
    detailSummary = null;
    detailSummaryLoading = Boolean(file.isDir && loadFolderSummary);

    if (file.isDir && loadFolderSummary) {
      const fileId = file.id;
      loadFolderSummary(fileId)
        .then((summary) => {
          if (detailFile?.id === fileId) detailSummary = summary;
        })
        .catch(() => {
          if (detailFile?.id === fileId) toast.error(m.load_failed());
        })
        .finally(() => {
          if (detailFile?.id === fileId) detailSummaryLoading = false;
        });
    }
  }

  function handleDetailOpenChange(open: boolean) {
    detailOpen = open;
    if (!open) {
      detailFile = null;
      detailSummary = null;
      detailSummaryLoading = false;
    }
  }

  async function copyFileLink(file: NormalizedFile) {
    try {
      const url = new URL(downloadUrlFn(file.id), window.location.origin);
      const token = getAccessToken();
      if (token) url.searchParams.set("access_token", token);
      await navigator.clipboard.writeText(url.toString());
      toast.success(m.copied());
    } catch {
      toast.error(m.copy_failed());
    }
  }

  function describeFolder(file: NormalizedFile): string {
    const size = fmtSize(detailSummary?.size ?? file.size);
    if (detailSummary) {
      return m.folder_detail_size({
        size,
        files: String(detailSummary.fileCount),
        folders: String(detailSummary.folderCount),
      });
    }
    return detailSummaryLoading ? m.folder_detail_size_loading({ size }) : size;
  }

  function describeDetailFile(file: NormalizedFile): string {
    return file.isDir ? describeFolder(file) : fmtSize(file.size);
  }

  function detailPathParts(): string[] {
    return [m.all_files(), ...currentPath.map((item) => item.name)];
  }

  function authedFileUrl(file: NormalizedFile): string {
    const token = getAccessToken();
    const url = downloadUrlFn(file.id);
    if (!token) return url;
    const sep = url.includes("?") ? "&" : "?";
    return `${url}${sep}access_token=${encodeURIComponent(token)}`;
  }

  function isImageFile(file: NormalizedFile): boolean {
    return file.mimeType?.startsWith("image/") ?? false;
  }

  function isVideoFile(file: NormalizedFile): boolean {
    return file.mimeType?.startsWith("video/") ?? false;
  }
  let prevDirId = $state('');
  let prevFileCount = $state(0);

  $effect(() => {
    const dirChanged = directoryId !== prevDirId;
    const loaded = prevFileCount === 0 && files.length > 0;
    prevDirId = directoryId;
    prevFileCount = files.length;
    if (dirChanged || loaded) {
      fadeKey++;
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
  {#key fadeKey}
    <div
      class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
      in:fade={{ duration: 150 }}
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
            onShowDetails={showDetails}
            onCopyLink={copyFileLink}
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
  {#key fadeKey}
    <div
      class="relative overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm"
      in:fade={{ duration: 150 }}
    >
    <table class="w-full table-fixed text-sm">
      <thead>
        <tr class="border-b border-gray-100 text-left text-xs text-gray-400">
          <th class="w-[54%] px-2 py-2.5 font-medium">
            <div class="flex items-center gap-2">
              <button type="button" onclick={toggleSelectAll} class="flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors {allSelected ? 'border-blue-500 bg-blue-500 text-white' : 'border-gray-300 hover:border-blue-400'}">
                {#if allSelected}
                  <Check size={10} />
                {/if}
              </button>
              {m.col_filename()}
            </div>
          </th>
          <th class="w-[12%] px-4 py-2.5 text-right font-medium">{m.col_size()}</th>
          <th class="w-[17%] px-4 py-2.5 text-right font-medium text-gray-500">{m.sort_created()}</th>
          <th class="w-[17%] px-4 py-2.5 text-right font-medium text-gray-500">{m.col_modified()}</th>
        </tr>
      </thead>
      <tbody>
        {#each files as f, i (f.id)}
          {@const isSelected = selected.has(f.id)}
          <tr
            class="group border-b border-gray-50 transition-colors last:border-0 hover:bg-gray-50/80 {f.isDir ||
            canPreview(f.mimeType)
              ? 'cursor-pointer'
              : ''} {isSelected ? 'bg-blue-50/50' : ''}"
            onclick={f.isDir
              ? () => onNavigateDir(f.id)
              : canPreview(f.mimeType)
                ? () => onPreview(f)
                : undefined}
          >
            <td class="px-2 py-2.5">
              <div class="flex items-center gap-2.5">
                <!-- Checkbox: hidden by default, show on hover or when selected -->
                <button type="button" onclick={(e) => { e.stopPropagation(); toggleSelect(f.id); }} class="flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-opacity {isSelected ? 'border-blue-500 bg-blue-500 text-white opacity-100' : 'border-gray-300 opacity-0 group-hover:opacity-100'}">
                  {#if isSelected}
                    <Check size={10} />
                  {/if}
                </button>
                <span class="shrink-0">
                  <MimeIcon
                    mimeType={f.mimeType}
                    isDir={f.isDir}
                    category={f.fileCategory}
                    size={18}
                  />
                </span>
                <span class="min-w-0 flex-1 truncate text-gray-700" title={f.name}>{f.name}</span>
                {#if f.isStarred}
                  <Star size={12} class="shrink-0 text-amber-400" fill="currentColor" />
                {/if}
                <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
                <span
                  class="flex h-7 w-7 shrink-0 items-center justify-center opacity-0 transition-opacity group-hover:opacity-100 {isSelected ? 'invisible pointer-events-none' : ''}"
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
                    onShowDetails={showDetails}
                    onCopyLink={copyFileLink}
                    triggerClass="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                  />
                </span>
              </div>
            </td>
            <td class="px-4 py-2.5 text-right text-gray-500">
              {f.isDir ? "-" : fmtSize(f.size)}
            </td>
            <td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-500">
              {fmtTime(f.createdAt)}
            </td>
            <td class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-gray-500">
              {fmtTime(f.updatedAt)}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>

    </div>

    {#if hasSelection}
      <div
        class="fixed bottom-6 left-1/2 z-50 max-w-[calc(100vw-1rem)] -translate-x-1/2"
      >
        <div
          class="flex items-center gap-2 overflow-x-auto rounded-full border border-gray-100 bg-white/95 px-3 py-2 shadow-[0_12px_36px_rgba(15,23,42,0.16)] backdrop-blur"
          transition:fly={{ y: 16, duration: 180, opacity: 0 }}
        >
          <span class="shrink-0 px-3 text-sm font-medium text-gray-700">{m.selected_count({ count: String(selected.size) })}</span>
          <div class="h-7 w-px shrink-0 bg-gray-100"></div>
          <div class="flex items-center gap-1">
            <Tooltip
              content={m.download()}
              delayDuration={200}
              triggerProps={{ 'aria-label': m.download() }}
              triggerClass="h-8 w-8 rounded-full text-gray-600 transition-colors hover:bg-blue-50 hover:text-blue-600"
            >
              <Download size={16} />
            </Tooltip>
            {#if onStar}
              <Tooltip
                content={m.star_file()}
                delayDuration={200}
                triggerProps={{ 'aria-label': m.star_file() }}
                triggerClass="h-8 w-8 rounded-full text-gray-600 transition-colors hover:bg-amber-50 hover:text-amber-500"
              >
                <Star size={16} />
              </Tooltip>
            {/if}
            <Tooltip
              content={m.delete_label()}
              delayDuration={200}
              triggerProps={{ 'aria-label': m.delete_label() }}
              triggerClass="h-8 w-8 rounded-full text-gray-600 transition-colors hover:bg-red-50 hover:text-red-500"
            >
              <Trash2 size={16} />
            </Tooltip>
            <div class="mx-1 h-7 w-px bg-gray-100"></div>
            <Tooltip
              content={m.close()}
              delayDuration={200}
              triggerProps={{
                'aria-label': m.close(),
                onclick: () => (selected = new Set()),
              }}
              triggerClass="h-8 w-8 shrink-0 rounded-full text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
            >
              <X size={16} />
            </Tooltip>
          </div>
        </div>
      </div>
    {/if}
  {/key}
{/if}

<Dialog
  bind:open={detailOpen}
  onOpenChange={handleDetailOpenChange}
  title={detailFile?.name}
  footer={false}
  class="max-w-[520px]"
  bodyClass="px-5 py-5"
>
  {#if detailFile}
    <div>
      <div class="mb-6 flex h-[120px] items-center justify-center">
        {#if detailFile.isDir}
          <div class="relative h-[84px] w-[160px] rounded-[12px] bg-gradient-to-br from-indigo-400 via-blue-300 to-violet-300 shadow-[0_8px_18px_rgba(99,102,241,0.18)] before:absolute before:-top-[12px] before:left-[4px] before:h-[28px] before:w-[72px] before:rounded-t-[10px] before:bg-indigo-300 before:content-[''] after:absolute after:-top-[1px] after:left-0 after:h-[20px] after:w-full after:rounded-t-[10px] after:bg-white/20 after:content-['']"></div>
        {:else if isImageFile(detailFile)}
          <img
            src={authedFileUrl(detailFile)}
            alt={detailFile.name}
            class="max-h-[112px] max-w-[220px] rounded-lg border border-gray-200 object-contain"
          />
        {:else if isVideoFile(detailFile)}
          <video
            src={authedFileUrl(detailFile)}
            class="max-h-[112px] max-w-[220px] rounded-lg border border-gray-200 object-cover"
            muted
            preload="metadata"
          ></video>
        {:else}
          <MimeIcon
            mimeType={detailFile.mimeType}
            isDir={detailFile.isDir}
            category={detailFile.fileCategory}
            size={88}
          />
        {/if}
      </div>

      <section>
        <h3 class="text-sm font-medium leading-5 text-gray-800">{m.detail_info()}</h3>
        <div class="mt-4 space-y-5">
          <div>
            <p class="text-xs leading-5 text-gray-500">{detailFile.name}</p>
            <p class="mt-1.5 text-sm leading-5 text-gray-800">{describeDetailFile(detailFile)}</p>
          </div>

          <div>
            <p class="text-xs leading-5 text-gray-500">{m.file_location()}</p>
            <p class="mt-1.5 text-sm leading-5 text-gray-800">
              {#each detailPathParts() as part, index}
                {#if index > 0}
                  <span class="px-1 text-gray-800">›</span>
                {/if}
                <span class="underline underline-offset-4">{part}</span>
              {/each}
            </p>
          </div>

          <div>
            <p class="text-xs leading-5 text-gray-500">{m.cloud_created_time()}</p>
            <p class="mt-1.5 text-sm leading-5 text-gray-800">{fmtTime(detailFile.createdAt)}</p>
          </div>

          <div>
            <p class="text-xs leading-5 text-gray-500">{m.last_modified_time()}</p>
            <p class="mt-1.5 text-sm leading-5 text-gray-800">{fmtTime(detailFile.updatedAt)}</p>
          </div>
        </div>
      </section>
    </div>
  {/if}
</Dialog>
