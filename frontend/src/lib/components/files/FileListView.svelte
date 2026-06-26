<script lang="ts">
  import { LoaderCircle } from "@lucide/svelte";
  import { fade } from "svelte/transition";
  import type { NormalizedFile } from "$lib/types/file";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { lockManager } from "$lib/services/lockManager.svelte";
  import { settingsManager } from "$lib/services/settingsManager.svelte";
  import * as m from "$lib/paraglide/messages";
  import noFilesSvg from "$lib/assets/empty-states/no-files.svg";
  import { getAccessToken } from "$lib/api/client";
  import { toast } from "svelte-sonner";
  import { authedUrl } from "$lib/utils/format";
  import { copyToClipboard } from "$lib/utils/format";
  import FileGrid from "./FileGrid.svelte";
  import FileTable from "./FileTable.svelte";
  import FileDetailDialog from "./FileDetailDialog.svelte";
  import MoveDialog from "./MoveDialog.svelte";

  let {
    files,
    loading,
    emptyMessage,
    onNavigateDir,
    onBatchShare,
  }: {
    files: NormalizedFile[];
    loading: boolean;
    emptyMessage: string;
    onNavigateDir: (slug: string) => void;
    onBatchShare?: (files: NormalizedFile[]) => void;
  } = $props();

  // --- Local UI state (dialogs) ---
  let detailFile = $state<NormalizedFile | null>(null);
  let detailOpen = $state(false);
  let detailSummary = $state<{
    fileCount: number;
    folderCount: number;
    size: number;
  } | null>(null);
  let detailSummaryLoading = $state(false);
  let moveOpen = $state(false);
  let moveTargets = $state<NormalizedFile[]>([]);

  let moveExcludedIds = $derived(
    moveTargets.filter((file) => file.isDir).map((file) => file.id),
  );

  // --- Detail dialog ---
  function showDetails(file: NormalizedFile) {
    if (file.isDir && lockManager.isEffectivelyLocked(file)) {
      toast.error(m.dir_locked_cannot_view());
      return;
    }
    detailFile = file;
    detailOpen = true;
    detailSummary = null;
    detailSummaryLoading = file.isDir;

    if (file.isDir) {
      const fileId = file.id;
      fileManager
        .loadFolderSummary(fileId)
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

  function handleDetailOpenChangeComplete(open: boolean) {
    if (!open) detailFile = null;
  }

  // --- Move dialog ---
  function openMoveDialog(targetFiles: NormalizedFile[]) {
    const valid = targetFiles.filter(
      (file) => !file.isSystem && !lockManager.isEffectivelyLocked(file),
    );
    if (valid.length === 0) return;
    moveTargets = valid;
    moveOpen = true;
  }

  function openSingleMoveDialog(file: NormalizedFile) {
    openMoveDialog([file]);
  }

  function openSelectedMoveDialog() {
    openMoveDialog(
      files.filter((file) => fileManager.selectedIds.has(file.id)),
    );
  }

  function handleMoveClose() {
    moveOpen = false;
    moveTargets = [];
  }

  async function confirmMove(targetParentSlug: string) {
    if (moveTargets.length === 0) return;
    await fileManager.move(
      moveTargets.map((file) => file.id),
      targetParentSlug,
    );
    fileManager.clearSelection();
    moveOpen = false;
    moveTargets = [];
  }

  // --- Batch actions ---
  function handleBatchDownload() {
    const selectedFiles = files.filter(
      (f) => fileManager.selectedIds.has(f.id) && !f.isDir,
    );
    for (const f of selectedFiles) {
      const url = authedUrl(fileManager.getDownloadUrl(f.id));
      const a = document.createElement("a");
      a.href = url;
      a.download = f.name;
      a.click();
      a.remove();
    }
  }

  function handleBatchShare() {
    onBatchShare?.(files.filter((f) => fileManager.selectedIds.has(f.id)));
  }

  // --- Copy helpers ---
  async function copyFileLink(file: NormalizedFile) {
    const url = new URL(
      fileManager.getDownloadUrl(file.id),
      window.location.origin,
    );
    const token = getAccessToken();
    if (token) url.searchParams.set("access_token", token);
    if (await copyToClipboard(url.toString())) {
      toast.success(m.copied());
    } else {
      toast.error(m.copy_failed());
    }
  }

  async function copyFileHash(file: NormalizedFile) {
    if (!file.fileHash) return;
    if (await copyToClipboard(file.fileHash)) {
      toast.success(m.copied());
    } else {
      toast.error(m.copy_failed());
    }
  }
</script>

<div class="relative min-h-37.5">
  {#if loading}
    <div
      class="absolute inset-0 z-10 flex items-center justify-center"
      transition:fade={{ duration: 150 }}
    >
      <LoaderCircle size={24} class="animate-spin text-ink-4" />
    </div>
  {/if}

  {#if !loading && files.length === 0}
    <div
      class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center"
      in:fade={{ duration: 150 }}
    >
      <img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
      <p class="text-sm text-ink-4">{emptyMessage}</p>
    </div>
  {:else if files.length > 0}
    {#if fileManager.viewMode.current === "grid"}
      <FileGrid
        {files}
        {onNavigateDir}
        onMoveFile={openSingleMoveDialog}
        onShowDetails={showDetails}
        onCopyLink={copyFileLink}
        onCopyHash={copyFileHash}
        onShare={onBatchShare
          ? (f) => {
              onBatchShare([f]);
            }
          : undefined}
      />
    {:else}
      <FileTable
        {files}
        {onNavigateDir}
        onMoveFile={openSingleMoveDialog}
        onShowDetails={showDetails}
        onCopyLink={copyFileLink}
        onCopyHash={copyFileHash}
        onBatchDownload={handleBatchDownload}
        onBatchShare={onBatchShare ? handleBatchShare : undefined}
        onBatchMove={openSelectedMoveDialog}
      />
    {/if}
  {/if}
</div>

<MoveDialog
  bind:open={moveOpen}
  excludedIds={moveExcludedIds}
  includeSystemDirs={settingsManager.showSystemDirs}
  onClose={handleMoveClose}
  onConfirm={confirmMove}
/>

<FileDetailDialog
  bind:open={detailOpen}
  file={detailFile}
  currentPath={fileManager.crumbs}
  downloadUrlFn={fileManager.getDownloadUrl}
  summary={detailSummary}
  summaryLoading={detailSummaryLoading}
  onOpenChangeComplete={handleDetailOpenChangeComplete}
/>
