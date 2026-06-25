<script lang="ts">
  import {
    Check,
    Star,
    Lock,
    LockOpen,
    Download,
    Trash2,
    X,
    FolderInput,
    Share2,
  } from "@lucide/svelte";
  import { fade, fly } from "svelte/transition";
  import { fmtSize, fmtTime } from "$lib/utils/format";
  import type { NormalizedFile } from "$lib/types/file";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { lockManager } from "$lib/services/lockManager.svelte";
  import { previewManager } from "$lib/services/previewManager.svelte";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { Tooltip } from "$lib/ui/tooltip";
  import FileActionsDropdown from "./FileActionsDropdown.svelte";
  import LazyThumbnail from "./LazyThumbnail.svelte";
  import {
    isImageFile,
    canPreview,
    authedThumbnailUrl,
  } from "$lib/utils/file-helpers";
  import { toast } from "svelte-sonner";

  let {
    files,
    onNavigateDir,
    onMoveFile,
    onShowDetails,
    onCopyLink,
    onCopyHash,
    onShare,
    onBatchDownload,
    onBatchShare,
    onBatchMove,
  }: {
    files: NormalizedFile[];
    onNavigateDir: (slug: string) => void;
    onMoveFile?: (file: NormalizedFile) => void;
    onShowDetails: (file: NormalizedFile) => void;
    onCopyLink: (file: NormalizedFile) => void;
    onCopyHash?: (file: NormalizedFile) => void;
    onShare?: (file: NormalizedFile) => void;
    onBatchDownload: () => void;
    onBatchShare?: (files: NormalizedFile[]) => void;
    onBatchMove?: () => void;
  } = $props();

  let failedThumbs = $state<Set<string>>(new Set());

  function showThumbnail(file: NormalizedFile): boolean {
    return isImageFile(file) && !failedThumbs.has(file.id);
  }

  function markThumbnailFailed(fileId: string) {
    failedThumbs.add(fileId);
    failedThumbs = new Set(failedThumbs);
  }

  function handleBatchDelete() {
    const selected = fileManager.selectedIds;
    const locked = Array.from(selected).filter((id) => {
      const f = files.find((f) => f.id === id);
      return f && lockManager.isEffectivelyLocked(f);
    });
    const ids = Array.from(selected).filter((id) => {
      const f = files.find((f) => f.id === id);
      return !f || !lockManager.isEffectivelyLocked(f);
    });
    if (locked.length > 0) toast.info(`已跳过 ${locked.length} 个加锁目录`);
    if (ids.length > 0) fileManager.batchRemove(ids);
  }

  function handelClick(f: NormalizedFile) {
    if (f.isDir) {
      onNavigateDir(f.slug);
    } else {
      canPreview(f) ? previewManager.open(f) : void 0;
    }
  }

  $inspect(lockManager.unlockedSlugs);
</script>

<div
  class="relative overflow-hidden rounded-xl border border-line-soft bg-white"
  in:fade={{ duration: 150 }}
>
  <table class="w-full table-fixed text-sm">
    <thead>
      <tr class="border-b border-line-soft text-left text-xs text-ink-4">
        <th class="w-[54%] px-2 py-2.5 font-medium">
          <div class="flex items-center gap-2">
            <button
              type="button"
              onclick={() => fileManager.toggleSelectAll()}
              class="flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors {fileManager.allSelected
                ? 'border-primary bg-primary text-white'
                : 'border-line hover:border-primary'}"
            >
              {#if fileManager.allSelected}
                <Check size={10} />
              {/if}
            </button>
            {m.col_filename()}
          </div>
        </th>
        <th class="w-[12%] px-4 py-2.5 text-right font-medium"
          >{m.col_size()}</th
        >
        <th class="w-[17%] px-4 py-2.5 text-right font-medium text-ink-3"
          >{m.sort_created()}</th
        >
        <th class="w-[17%] px-4 py-2.5 text-right font-medium text-ink-3"
          >{m.col_modified()}</th
        >
      </tr>
    </thead>
    <tbody>
      {#each files as f, i (f.id)}
        {@const isSelected = fileManager.selectedIds.has(f.id)}
        <tr
          class="group border-b border-line-soft transition-colors last:border-0 hover:bg-surface-muted/80 {f.isDir ||
          canPreview(f)
            ? 'cursor-pointer'
            : ''} {isSelected ? 'bg-primary-soft/50' : ''}"
          onclick={() => handelClick(f)}
        >
          <td class="px-2 py-2.5">
            <div class="flex items-center gap-2.5">
              {#if f.isSystem}
                <span class="h-4 w-4 shrink-0"></span>
              {:else}
                <button
                  type="button"
                  onclick={(e) => {
                    e.stopPropagation();
                    fileManager.toggleSelect(f.id);
                  }}
                  class="flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-opacity {isSelected
                    ? 'border-primary bg-primary text-white opacity-100'
                    : 'border-line opacity-0 group-hover:opacity-100'}"
                >
                  {#if isSelected}
                    <Check size={10} />
                  {/if}
                </button>
              {/if}
              {#if showThumbnail(f)}
                <span
                  class="flex h-6 w-6 shrink-0 items-center justify-center overflow-hidden rounded bg-surface-muted"
                >
                  <LazyThumbnail
                    src={authedThumbnailUrl(f)}
                    containerClass="flex h-full w-full"
                    imgClass="h-full w-full object-cover"
                    onError={() => markThumbnailFailed(f.id)}
                  />
                </span>
              {:else}
                <MimeIcon
                  mimeType={f.mimeType}
                  name={f.name}
                  isDir={f.isDir}
                  category={f.fileCategory}
                  size={24}
                />
              {/if}
              <span class="min-w-0 flex-1 truncate text-ink-2" title={f.name}
                >{f.name}</span
              >
              {#if f.isSystem}
                <span
                  class="shrink-0 rounded-full border border-info bg-info-soft px-1.5 py-0.5 text-[10px] font-medium leading-4 text-info"
                >
                  {m.system_badge()}
                </span>
              {/if}
              {#if f.hasPassword}
                {#if lockManager.isUnlocked(f.slug)}
                  <LockOpen size={12} class="shrink-0 text-success" />
                {:else}
                  <Lock size={12} class="shrink-0 text-ink-4" />
                {/if}
              {/if}
              {#if f.isStarred}
                <Star
                  size={12}
                  class="shrink-0 text-warning"
                  fill="currentColor"
                />
              {/if}
              <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
              <span
                class="flex h-7 w-7 shrink-0 items-center justify-center opacity-0 transition-opacity group-hover:opacity-100 {isSelected
                  ? 'invisible pointer-events-none'
                  : ''}"
                onclick={(e) => e.stopPropagation()}
              >
                <FileActionsDropdown
                  file={f}
                  onMove={onMoveFile}
                  {onShowDetails}
                  {onCopyLink}
                  {onCopyHash}
                  {onShare}
                  triggerClass="rounded-md p-1 text-ink-4 hover:bg-surface-sunken hover:text-ink-3"
                />
              </span>
            </div>
          </td>
          <td class="px-4 py-2.5 text-right text-ink-3">
            {f.isDir ? "-" : fmtSize(f.size)}
          </td>
          <td
            class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-ink-3"
          >
            {fmtTime(f.createdAt)}
          </td>
          <td
            class="whitespace-nowrap px-4 py-2.5 text-right text-xs text-ink-3"
          >
            {fmtTime(f.updatedAt)}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

{#if fileManager.hasSelection}
  <div
    class="fixed bottom-6 left-1/2 z-50 max-w-[calc(100vw-1rem)] -translate-x-1/2"
  >
    <div
      class="flex items-center gap-2 overflow-x-auto rounded-full border border-line-soft bg-white/95 px-3 py-2 shadow-[0_12px_36px_rgba(15,23,42,0.16)] backdrop-blur"
      transition:fly={{ y: 16, duration: 180, opacity: 0 }}
    >
      <span class="shrink-0 px-3 text-sm font-medium text-ink-2"
        >{m.selected_count({
          count: String(fileManager.selectedIds.size),
        })}</span
      >
      <div class="h-7 w-px shrink-0 bg-surface-sunken"></div>
      <div class="flex items-center gap-1">
        <Tooltip
          content={m.download()}
          delayDuration={200}
          triggerProps={{
            "aria-label": m.download(),
            onclick: onBatchDownload,
          }}
          triggerClass="h-8 w-8 rounded-full text-ink-3 transition-colors hover:bg-primary-soft hover:text-primary"
        >
          <Download size={16} />
        </Tooltip>
        {#if onBatchMove}
          <Tooltip
            content={m.move_to()}
            delayDuration={200}
            triggerProps={{
              "aria-label": m.move_to(),
              onclick: onBatchMove,
            }}
            triggerClass="h-8 w-8 rounded-full text-ink-3 transition-colors hover:bg-primary-soft hover:text-primary"
          >
            <FolderInput size={16} />
          </Tooltip>
        {/if}
        {#if onBatchShare}
          <Tooltip
            content="分享"
            delayDuration={200}
            triggerProps={{
              "aria-label": "分享",
              onclick: () =>
                onBatchShare?.(
                  files.filter((f) => fileManager.selectedIds.has(f.id)),
                ),
            }}
            triggerClass="h-8 w-8 rounded-full text-ink-3 transition-colors hover:bg-primary-soft hover:text-primary"
          >
            <Share2 size={16} />
          </Tooltip>
        {/if}
        <Tooltip
          content={m.delete_label()}
          delayDuration={200}
          triggerProps={{
            "aria-label": m.delete_label(),
            onclick: handleBatchDelete,
          }}
          triggerClass="h-8 w-8 rounded-full text-ink-3 transition-colors hover:bg-danger-soft hover:text-danger"
        >
          <Trash2 size={16} />
        </Tooltip>
        <div class="mx-1 h-7 w-px bg-surface-sunken"></div>
        <Tooltip
          content={m.close()}
          delayDuration={200}
          triggerProps={{
            "aria-label": m.close(),
            onclick: () => fileManager.clearSelection(),
          }}
          triggerClass="h-8 w-8 shrink-0 rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
        >
          <X size={16} />
        </Tooltip>
      </div>
    </div>
  </div>
{/if}
