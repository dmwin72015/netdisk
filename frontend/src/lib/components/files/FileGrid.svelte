<script lang="ts">
  import { Lock } from "@lucide/svelte";
  import { fade } from "svelte/transition";
  import { fmtSize } from "$lib/utils/format";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import FileActionsDropdown from "./FileActionsDropdown.svelte";
  import LazyThumbnail from "./LazyThumbnail.svelte";
  import { isImageFile, canPreview, authedThumbnailUrl } from "$lib/utils/file-helpers";

  let {
    files,
    downloadUrlFn,
    onNavigateDir,
    onPreview,
    onStar,
    onRename,
    onDelete,
    onMoveFile,
    onAddToMedia,
    onShare,
    onSetDirectoryLock,
    onClearDirectoryLock,
    onForceDeleteDir,
    onShowDetails,
    onCopyLink,
  }: {
    files: NormalizedFile[];
    downloadUrlFn: (id: string) => string;
    onNavigateDir: (id: string) => void;
    onPreview: (file: NormalizedFile) => void;
    onStar?: (id: string, starred: boolean) => void;
    onRename: (id: string, name: string) => void;
    onDelete: (id: string, name: string) => void;
    onMoveFile?: (file: NormalizedFile) => void;
    onAddToMedia?: (file: NormalizedFile) => void;
    onShare?: (file: NormalizedFile) => void;
    onSetDirectoryLock?: (file: NormalizedFile) => void;
    onClearDirectoryLock?: (file: NormalizedFile) => void;
    onForceDeleteDir?: (file: NormalizedFile) => void;
    onShowDetails: (file: NormalizedFile) => void;
    onCopyLink: (file: NormalizedFile) => void;
  } = $props();

  let failedThumbs = $state<Set<string>>(new Set());

  function showThumbnail(file: NormalizedFile): boolean {
    return isImageFile(file) && !failedThumbs.has(file.id);
  }

  function markThumbnailFailed(fileId: string) {
    failedThumbs.add(fileId);
    failedThumbs = new Set(failedThumbs);
  }
</script>

<div
  class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
  in:fade={{ duration: 150 }}
>
  {#each files as f, i (f.id)}
    <div
      class="group relative flex flex-col items-center rounded-xl border border-line-soft bg-white p-4 transition-all hover:border-line hover:{f.isDir ||
      canPreview(f)
        ? 'cursor-pointer'
        : ''}"
      role="button"
      tabindex="0"
      onclick={f.isDir
        ? () => onNavigateDir(f.id)
        : !f.isDir && canPreview(f)
          ? () => onPreview(f)
          : undefined}
      onkeydown={(e) => {
        if (
          (e.key === "Enter" || e.key === " ") &&
          (f.isDir || canPreview(f))
        ) {
          e.preventDefault();
          if (f.isDir) onNavigateDir(f.id);
          else if (canPreview(f)) onPreview(f);
        }
      }}
    >
      <div
        class="absolute right-1.5 top-1.5"
        role="button"
        tabindex="0"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => {
          if (e.key === "Enter" || e.key === " ") e.stopPropagation();
        }}
      >
        <FileActionsDropdown
          file={f}
          {downloadUrlFn}
          {onStar}
          {onPreview}
          {onRename}
          {onDelete}
          onMove={onMoveFile}
          {onAddToMedia}
          {onShare}
          {onSetDirectoryLock}
          {onClearDirectoryLock}
          {onForceDeleteDir}
          {onShowDetails}
          {onCopyLink}
          triggerClass="rounded-lg bg-white/90 p-1 text-ink-4 backdrop-blur transition-colors hover:bg-white hover:text-ink-3"
        />
      </div>
      {#if showThumbnail(f)}
        <LazyThumbnail
          src={authedThumbnailUrl(f)}
          containerClass="flex h-12 w-12 shrink-0 overflow-hidden rounded-lg border border-line-soft bg-surface-muted"
          imgClass="h-full w-full object-cover"
          onError={() => markThumbnailFailed(f.id)}
        />
      {:else}
        <MimeIcon
          mimeType={f.mimeType}
          name={f.name}
          isDir={f.isDir}
          category={f.fileCategory}
          size={36}
        />
      {/if}
      <div class="mt-3 flex w-full min-w-0 items-center justify-center gap-1.5">
        <p
          class="min-w-0 truncate text-center text-sm font-medium text-ink-2"
          title={f.name}
        >
          {f.name}
        </p>
        {#if f.isSystem}
          <span
            class="shrink-0 rounded-full bg-info-soft px-1.5 py-0.5 text-[10px] font-medium text-info"
          >
            {m.system_badge()}
          </span>
        {/if}
        {#if f.isLocked}
          <Lock size={12} class="shrink-0 text-ink-4" />
        {/if}
      </div>
      <p class="mt-0.5 text-xs text-ink-4">
        {f.isDir ? "" : fmtSize(f.size)}
      </p>
    </div>
  {/each}
</div>
