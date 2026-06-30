<script lang="ts">
  import { Lock, LockOpen, Check } from "@lucide/svelte";
  import { fade } from "svelte/transition";
  import { fmtSize, fmtTime } from "$lib/utils/format";
  import type { NormalizedFile } from "$lib/types/file";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { lockManager } from "$lib/services/lockManager.svelte";
  import { previewManager } from "$lib/services/previewManager.svelte";
  import { confirmAction } from "$lib/dialog";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import FileActionsDropdown from "./FileActionsDropdown.svelte";
  import LazyThumbnail from "./LazyThumbnail.svelte";
  import {
    isImageFile,
    canPreview,
    authedThumbnailUrl,
  } from "$lib/utils/file-helpers";

  let {
    files,
    onNavigateDir,
    onMoveFile,
    onShowDetails,
    onCopyLink,
    onCopyHash,
    onShare,
  }: {
    files: NormalizedFile[];
    onNavigateDir: (slug: string) => void;
    onMoveFile?: (file: NormalizedFile) => void;
    onShowDetails: (file: NormalizedFile) => void;
    onCopyLink: (file: NormalizedFile) => void;
    onCopyHash?: (file: NormalizedFile) => void;
    onShare?: (file: NormalizedFile) => void;
  } = $props();

  let failedThumbs = $state<Set<string>>(new Set());

  function showThumbnail(file: NormalizedFile): boolean {
    return isImageFile(file) && !failedThumbs.has(file.id);
  }

  function markThumbnailFailed(fileId: string) {
    failedThumbs.add(fileId);
    failedThumbs = new Set(failedThumbs);
  }

  async function toggleLock(f: NormalizedFile) {
    if (lockManager.isEffectivelyLocked(f)) {
      await lockManager.unlock(f.slug, f.name);
    } else {
      const ok = await confirmAction(m.dir_password(), m.dir_relock_confirm({ name: f.name }), m.confirm());
      if (ok) lockManager.relock(f.slug);
    }
  }
</script>

<div
  class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
  in:fade={{ duration: 150 }}
>
  {#each files as f, i (f.id)}
    <div
      class="group relative flex flex-col items-center rounded-xl border border-line-soft bg-surface p-4 transition-all hover:border-line hover:{f.isDir ||
      canPreview(f)
        ? 'cursor-pointer'
        : ''}"
      role="button"
      tabindex="0"
      onclick={f.isDir
        ? () => onNavigateDir(f.id)
        : !f.isDir && canPreview(f)
          ? () => previewManager.open(f)
          : undefined}
      onkeydown={(e) => {
        if (
          (e.key === "Enter" || e.key === " ") &&
          (f.isDir || canPreview(f))
        ) {
          e.preventDefault();
          if (f.isDir) onNavigateDir(f.id);
          else if (canPreview(f)) previewManager.open(f);
        }
      }}
    >
      {#if !f.isSystem}
        {@const isSelected = !!fileManager.selectedIds[f.id]}
        <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
        <span
          role="checkbox"
          aria-checked={isSelected}
          tabindex="-1"
          class="absolute left-1.5 top-1.5 z-10 flex h-4 w-4 cursor-pointer items-center justify-center rounded-full border transition-colors {isSelected
            ? 'border-primary bg-primary text-primary-on opacity-100'
            : 'border-line bg-surface/90 opacity-0 backdrop-blur group-hover:opacity-100 hover:border-primary'}"
          onclick={(e) => {
            e.stopPropagation();
            fileManager.toggleSelect(f.id);
          }}
          onkeydown={(e) => {
            if (e.key === " " || e.key === "Enter") {
              e.stopPropagation();
              fileManager.toggleSelect(f.id);
            }
          }}
        >
          {#if isSelected}
            <Check size={10} />
          {/if}
        </span>
      {/if}
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
          onMove={onMoveFile}
          {onShowDetails}
          {onCopyLink}
          {onCopyHash}
          {onShare}
          triggerClass="rounded-lg bg-surface/90 p-1 text-ink-4 backdrop-blur transition-colors hover:bg-surface hover:text-ink-3"
        />
      </div>
      {#if showThumbnail(f)}
        <LazyThumbnail
          src={authedThumbnailUrl(f)}
          containerClass="flex h-12 w-12 shrink-0 overflow-hidden rounded-lg border border-line-soft bg-surface-sunken"
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
        {#if f.hasPassword}
          <button
            type="button"
            class="shrink-0 rounded-md p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
            onclick={(e) => { e.stopPropagation(); toggleLock(f); }}
            aria-label={lockManager.isEffectivelyLocked(f) ? m.dir_password() : m.dir_unlocked()}
          >
            {#if lockManager.isEffectivelyLocked(f)}
              <Lock size={12} />
            {:else}
              <LockOpen size={12} class="text-success" />
            {/if}
          </button>
        {/if}
      </div>
      <p class="mt-0.5 text-xs text-ink-4">
        {#if f.isDir}
          {fmtTime(f.updatedAt)}
        {:else}
          {fmtSize(f.size)} · {fmtTime(f.updatedAt)}
        {/if}
      </p>
    </div>
  {/each}
</div>
