<script lang="ts">
  import { LoaderCircle, FolderPlus } from "@lucide/svelte";
  import { fade } from "svelte/transition";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";
  import { getAccessToken } from "$lib/api/client";
  import { toast } from "svelte-sonner";
  import { authedUrl } from "$lib/utils/format";
  import { copyToClipboard } from "$lib/utils/format";
  import FileGrid from "./FileGrid.svelte";
  import FileTable from "./FileTable.svelte";
  import FileDetailDialog from "./FileDetailDialog.svelte";
  import MoveDialog from "./MoveDialog.svelte";

  type FolderSummary = {
    fileCount: number;
    folderCount: number;
    size: number;
  };

  let {
    files,
    viewMode,
    loading,
    directoryId = "",
    currentPath = [],
    includeSystemDirs = true,
    emptyMessage,
    downloadUrlFn,
    onNavigateDir,
    onStar,
    onPreview,
    onRename,
    onDelete,
    onBatchDelete,
    onBatchShare,
    onMove,
    onAddToMedia,
    onShare,
    onSetDirectoryLock,
    onClearDirectoryLock,
    onForceDeleteDir,
    loadFolderSummary,
  }: {
    files: NormalizedFile[];
    viewMode: "list" | "grid";
    loading: boolean;
    directoryId?: string;
    currentPath?: { id: string; name: string }[];
    includeSystemDirs?: boolean;
    emptyMessage: string;
    downloadUrlFn: (id: string) => string;
    onNavigateDir: (id: string) => void;
    onStar?: (id: string, starred: boolean) => void;
    onPreview: (file: NormalizedFile) => void;
    onRename: (id: string, name: string) => void;
    onDelete: (id: string, name: string) => void;
    onBatchDelete?: (ids: string[]) => void;
    onBatchShare?: (files: NormalizedFile[]) => void;
    onMove?: (ids: string[], targetSlug: string) => Promise<void>;
    onAddToMedia?: (file: NormalizedFile) => void;
    onShare?: (file: NormalizedFile) => void;
    onSetDirectoryLock?: (file: NormalizedFile) => void;
    onClearDirectoryLock?: (file: NormalizedFile) => void;
    onForceDeleteDir?: (file: NormalizedFile) => void;
    loadFolderSummary?: (id: string) => Promise<FolderSummary>;
  } = $props();

  let selected = $state<Set<string>>(new Set());
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

  let selectableFiles = $derived(files.filter((file) => !file.isSystem));
  let allSelected = $derived(
    selectableFiles.length > 0 &&
      selectableFiles.every((f) => selected.has(f.id)),
  );
  let hasSelection = $derived(selected.size > 0);
  let moveExcludedIds = $derived(
    moveTargets.filter((file) => file.isDir).map((file) => file.id),
  );

  function toggleSelect(id: string) {
    const file = files.find((item) => item.id === id);
    if (file?.isSystem) return;
    if (selected.has(id)) selected.delete(id);
    else selected.add(id);
    selected = new Set(selected);
  }

  function toggleSelectAll() {
    if (allSelected) selected = new Set();
    else selected = new Set(selectableFiles.map((f) => f.id));
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

  function handleBatchDownload() {
    const selectedFiles = files.filter((f) => selected.has(f.id) && !f.isDir);
    for (const f of selectedFiles) {
      const url = authedUrl(downloadUrlFn(f.id));
      const a = document.createElement("a");
      a.href = url;
      a.download = f.name;
      a.click();
      a.remove();
    }
  }

  function handleDetailOpenChangeComplete(open: boolean) {
    if (!open) {
      detailFile = null;
    }
  }

  function openMoveDialog(targetFiles: NormalizedFile[]) {
    if (!onMove || targetFiles.length === 0) return;
    moveTargets = targetFiles.filter((file) => !file.isSystem);
    if (moveTargets.length === 0) return;
    moveOpen = true;
  }

  function openSingleMoveDialog(file: NormalizedFile) {
    openMoveDialog([file]);
  }

  function openSelectedMoveDialog() {
    openMoveDialog(files.filter((file) => selected.has(file.id)));
  }

  function handleMoveClose() {
    moveOpen = false;
    moveTargets = [];
  }

  async function confirmMove(targetParentSlug: string) {
    if (!onMove || moveTargets.length === 0) return;
    await onMove(
      moveTargets.map((file) => file.id),
      targetParentSlug,
    );
    selected = new Set();
    moveOpen = false;
    moveTargets = [];
  }

  async function copyFileLink(file: NormalizedFile) {
    const url = new URL(downloadUrlFn(file.id), window.location.origin);
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

  function handleBatchShare() {
    onBatchShare?.(files.filter((f) => selected.has(f.id)));
  }
</script>

{#if loading}
  <div
    class="flex items-center justify-center py-16"
    transition:fade={{ duration: 150 }}
  >
    <LoaderCircle size={24} class="animate-spin text-ink-4" />
  </div>
{:else if files.length === 0}
  <div
    class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-line py-16 text-center"
    in:fade={{ duration: 150 }}
  >
    <FolderPlus size={40} class="mb-3 text-ink-4" />
    <p class="text-sm text-ink-4">{emptyMessage}</p>
  </div>
{:else if viewMode === "grid"}
  <FileGrid
    {files}
    {downloadUrlFn}
    {onNavigateDir}
    {onPreview}
    {onStar}
    {onRename}
    {onDelete}
    onMoveFile={onMove ? openSingleMoveDialog : undefined}
    {onAddToMedia}
    {onShare}
    {onSetDirectoryLock}
    {onClearDirectoryLock}
    {onForceDeleteDir}
    onShowDetails={showDetails}
    onCopyLink={copyFileLink}
    onCopyHash={copyFileHash}
  />
{:else}
  <FileTable
    {files}
    {selected}
    {onNavigateDir}
    {onPreview}
    {onStar}
    {onRename}
    {onDelete}
    onToggleSelect={toggleSelect}
    onToggleSelectAll={toggleSelectAll}
    {hasSelection}
    {allSelected}
    {downloadUrlFn}
    onMoveFile={onMove ? openSingleMoveDialog : undefined}
    {onAddToMedia}
    {onShare}
    {onSetDirectoryLock}
    {onClearDirectoryLock}
    {onForceDeleteDir}
    onShowDetails={showDetails}
    onCopyLink={copyFileLink}
    onCopyHash={copyFileHash}
    onBatchDownload={handleBatchDownload}
    {onBatchDelete}
    onBatchShare={onBatchShare ? handleBatchShare : undefined}
    onBatchMove={onMove ? openSelectedMoveDialog : undefined}
    onCloseSelection={() => {
      selected = new Set();
    }}
  />
{/if}

<MoveDialog
  bind:open={moveOpen}
  excludedIds={moveExcludedIds}
  {includeSystemDirs}
  onClose={handleMoveClose}
  onConfirm={confirmMove}
/>

<FileDetailDialog
  bind:open={detailOpen}
  file={detailFile}
  {currentPath}
  {downloadUrlFn}
  summary={detailSummary}
  summaryLoading={detailSummaryLoading}
  onOpenChangeComplete={handleDetailOpenChangeComplete}
/>
