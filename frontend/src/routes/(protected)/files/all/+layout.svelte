<script lang="ts">
  import { getContext } from "svelte";
  import { goto, afterNavigate } from "$app/navigation";
  import { page } from "$app/state";
  import { browser } from "$app/environment";
  import { user, authReady } from "$lib/stores/auth";
  import {
    listFiles,
    mkdir,
    trashFile,
    batchTrashFiles,
    renameFile,
    setStarred,
    downloadUrl,
    getBreadcrumb,
    moveFile,
    setDirectoryLock,
    clearDirectoryLock,
    unlockDirectory,
    forceDeleteDir,
    type FileItem,
  } from "$lib/api/files";
  import { ApiError } from "$lib/api/client";
  import type { NormalizedFile } from "$lib/types/file";
  import { normalizeFileItem } from "$lib/types/adapters";
  import { addToLibrary } from "$lib/api/media";
  import { toast } from "svelte-sonner";
  import DrivePreview from "$lib/components/DrivePreview.svelte";
  import FileListView from "$lib/components/files/FileListView.svelte";
  import ShareDialog from "$lib/components/files/ShareDialog.svelte";
  import FolderUploadDialog from "$lib/components/files/FolderUploadDialog.svelte";
  import RemoteUploadDialog from "$lib/components/files/RemoteUploadDialog.svelte";
import TextUploadDialog from "$lib/components/files/TextUploadDialog.svelte";
  import ConflictDialog from "$lib/components/files/ConflictDialog.svelte";
  import type {
    NameConflictInfo,
    NameConflictResult,
  } from "$lib/upload-manager.svelte";
  import PasteUploadProvider from "$lib/components/files/PasteUploadProvider.svelte";
  import Breadcrumb from "$lib/components/Breadcrumb.svelte";
  import {
    getShowSystemDirs,
    setShowSystemDirs,
    getUploadConcurrency,
    getDirectoryUnlockTtlHours,
  } from "$lib/stores/file-preferences.svelte";
  import FilesToolbar, {
    type SortField,
    type ViewMode,
  } from "$lib/components/files/FilesToolbar.svelte";
  import { confirmDelete, promptInput } from "$lib/dialog";
  import { FileQuestionMark, LoaderCircle } from "@lucide/svelte";
  import type { createUploadManager as UploadMgrFn } from "$lib/upload-manager.svelte";
  type UploadManager = ReturnType<typeof UploadMgrFn>;
  import * as m from "$lib/paraglide/messages";

  let { children } = $props();

  const PAGE_SIZE = 50;

  // --- File listing ---
  let files = $state<FileItem[]>([]);
  let normalizedFiles = $derived(files.map(normalizeFileItem));
  let total = $state(0);
  let loading = $state(true);
  let loadingMore = $state(false);
  let deleting = $state(false);
  let notFound = $state(false);
  let refreshId = 0;

  // --- Preferences ---
  let viewMode = $state<ViewMode>(
    browser
      ? (localStorage.getItem("nd.files.view") as ViewMode) || "list"
      : "list",
  );
  function setViewMode(mode: ViewMode) {
    viewMode = mode;
    if (browser) localStorage.setItem("nd.files.view", mode);
  }

  let sortBy = $state<SortField>(
    browser
      ? (localStorage.getItem("nd.files.sortBy") as SortField) || "created_at"
      : "created_at",
  );
  let sortDir = $state<"ASC" | "DESC">(
    browser
      ? (localStorage.getItem("nd.files.sortDir") as "ASC" | "DESC") || "DESC"
      : "DESC",
  );
  function handleShowSystemDirsChange(value: boolean) {
    setShowSystemDirs(value);
    void refresh(true);
  }

  // --- Upload manager (shared via context) ---
  const upload = getContext<UploadManager>("upload");

  let remoteUploadOpen = $state(false);
  let textUploadOpen = $state(false);

  // --- Conflict resolution dialog ---
  type ConflictRequest = {
    conflicts: NameConflictInfo[];
    resolve: (results: Map<string, NameConflictResult>) => void;
  };
  let conflictDialogOpen = $state(false);
  let conflictDialogConflicts = $state<NameConflictInfo[]>([]);
  let conflictQueue: ConflictRequest[] = [];
  let activeConflict: ConflictRequest | null = null;

  function showNextConflict() {
    if (activeConflict || conflictQueue.length === 0) return;
    activeConflict = conflictQueue.shift()!;
    conflictDialogConflicts = activeConflict.conflicts;
    conflictDialogOpen = true;
  }

  $effect(() => {
    upload.updateMaxConcurrent(getUploadConcurrency());
    upload.setGetCurrentSlug(() => currentSlug);
    upload.setOnCompleted(() => refresh());
    upload.setOnNameConflicts((conflicts) => {
      return new Promise<Map<string, NameConflictResult>>((resolve) => {
        conflictQueue.push({ conflicts, resolve });
        showNextConflict();
      });
    });
  });

  function finishActiveConflict(results: Map<string, NameConflictResult>) {
    const current = activeConflict;
    activeConflict = null;
    conflictDialogConflicts = [];
    current?.resolve(results);
    showNextConflict();
  }

  function onConflictResolve(results: Map<string, NameConflictResult>) {
    finishActiveConflict(results);
  }

  function onConflictCancel() {
    const allSkipped = new Map<string, NameConflictResult>(
      (activeConflict?.conflicts ?? []).map((c) => [
        c.uid,
        { strategy: "skip" as const, applyToAll: false },
      ]),
    );
    finishActiveConflict(allSkipped);
  }

  let fileInput: HTMLInputElement | undefined = $state();
  let folderInput: HTMLInputElement | undefined = $state();

  function openFileUploadFromShell() {
    fileInput?.click();
  }

  if (browser) {
    $effect(() => {
      window.addEventListener("nd:open-file-upload", openFileUploadFromShell);
      return () =>
        window.removeEventListener(
          "nd:open-file-upload",
          openFileUploadFromShell,
        );
    });
  }

  function setSort(field: SortField) {
    if (sortBy === field) {
      sortDir = sortDir === "ASC" ? "DESC" : "ASC";
    } else {
      sortBy = field;
      sortDir = field === "file_name" ? "ASC" : "DESC";
    }
    if (browser) {
      localStorage.setItem("nd.files.sortBy", sortBy);
      localStorage.setItem("nd.files.sortDir", sortDir);
    }
    void refresh(true);
  }

  // --- Breadcrumb / Navigation ---
  let currentSlug = $state("");
  let crumbs = $state<{ id: string; name: string }[]>([]);
  let breadcrumbRef: Breadcrumb | undefined = $state();
  let pasteTargetLabel = $derived(crumbs.at(-1)?.name ?? m.nav_files());

  async function unlockDir(slug: string, name?: string) {
    const password = await promptInput(
      "目录密码",
      `请输入${name ? `「${name}」` : "目录"}的密码`,
      undefined,
      128,
    );
    if (!password) return false;
    try {
      await unlockDirectory(slug, password, getDirectoryUnlockTtlHours());
      toast.success("目录已解锁");
      return true;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "目录密码错误");
      return false;
    }
  }

  async function navigateToDirInternal(slug: string) {
    const file = normalizedFiles.find((item) => item.id === slug);
    if (file?.isLocked && !(await unlockDir(slug, file.name))) return;
    loading = true;
    files = [];
    void goto("/files/all/" + slug, { keepFocus: true, noScroll: true });
  }

  function navigateToDir(slug: string) {
    void navigateToDirInternal(slug);
  }

  async function fetchBreadcrumb(dirSlug: string) {
    if (!dirSlug) {
      crumbs = [];
      return;
    }
    try {
      const items = await getBreadcrumb(dirSlug);
      crumbs = items.map((b) => ({ id: b.slug, name: b.fileName }));
    } catch {
      crumbs = [{ id: dirSlug, name: dirSlug }];
    }
  }

  afterNavigate(() => {
    const slug = page.params.slug ?? "";
    if (slug !== currentSlug) {
      currentSlug = slug;
      breadcrumbRef?.collapse();
    }
    void fetchBreadcrumb(currentSlug);
    void refresh(true);
  });

  // --- File listing ---
  async function refresh(showLoading = false) {
    if (!$user) return;
    const id = ++refreshId;
    if (showLoading) loading = true;
    loadingMore = false;
    notFound = false;
    try {
      const data = await listFiles(
        currentSlug || undefined,
        1,
        PAGE_SIZE,
        undefined,
        undefined,
        sortBy,
        sortDir,
        false,
        getShowSystemDirs(),
      );
      if (id !== refreshId) return;
      files = data.files;
      total = data.total;
    } catch (e) {
      if (id !== refreshId) return;
      if (e instanceof ApiError && e.status === 404) {
        notFound = true;
      } else if (e instanceof ApiError && e.status === 423 && currentSlug) {
        if (await unlockDir(currentSlug, crumbs.at(-1)?.name)) {
          void refresh(showLoading);
        }
      } else {
        toast.error(e instanceof Error ? e.message : m.load_failed());
      }
    } finally {
      if (id === refreshId) loading = false;
    }
  }

  async function loadMore() {
    if (loadingMore) return;
    loadingMore = true;
    const id = ++refreshId;
    try {
      const page_num = Math.floor(files.length / PAGE_SIZE) + 1;
      const data = await listFiles(
        currentSlug || undefined,
        page_num,
        PAGE_SIZE,
        undefined,
        undefined,
        sortBy,
        sortDir,
        false,
        getShowSystemDirs(),
      );
      if (id !== refreshId) return;
      files = [...files, ...data.files];
    } catch (e) {
      if (id !== refreshId) return;
      toast.error(e instanceof Error ? e.message : m.load_more_failed());
    } finally {
      if (id === refreshId) loadingMore = false;
    }
  }

  async function loadFolderSummary(slug: string) {
    const pageSize = 100;
    let pageNum = 1;
    let loaded = 0;
    let total = 0;
    const summary = { fileCount: 0, folderCount: 0, size: 0 };

    do {
      const data = await listFiles(slug, pageNum, pageSize);
      total = data.total;
      loaded += data.files.length;
      for (const file of data.files) {
        if (file.isDir) summary.folderCount += 1;
        else {
          summary.fileCount += 1;
          summary.size += file.fileSize;
        }
      }
      pageNum += 1;
    } while (loaded < total);

    return summary;
  }

  // --- Create directory ---
  async function createDir() {
    const name = await promptInput(
      m.new_folder(),
      m.enter_folder_name(),
      undefined,
      100,
    );
    if (!name) return;
    const trimmed = name.trim();
    if (files.some((f) => f.isDir && f.fileName === trimmed)) {
      toast.error(m.dir_already_exists());
      return;
    }
    try {
      await mkdir(trimmed, currentSlug || undefined);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.create_dir_failed());
    }
  }

  // --- File operations ---
  async function remove(slug: string, name: string) {
    if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
    deleting = true;
    try {
      await trashFile(slug);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.delete_failed());
    } finally {
      deleting = false;
    }
  }

  async function forceRemoveDir(file: NormalizedFile) {
    if (
      !(await confirmDelete(
        `确认强制删除目录「${file.name}」及其所有内容？此操作不可恢复。`,
      ))
    )
      return;
    deleting = true;
    try {
      await forceDeleteDir(file.id);
      toast.success(`目录「${file.name}」已强制删除`);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "强制删除失败");
    } finally {
      deleting = false;
    }
  }

  async function batchRemove(ids: string[]) {
    const files = ids
      .map((id) => normalizedFiles.find((f) => f.id === id))
      .filter(Boolean) as NormalizedFile[];
    if (files.length === 0) return;
    const names = files.map((f) => f.name);
    if (
      !(await confirmDelete(
        m.confirm_delete_multiple({
          count: String(files.length),
          names: names.join("\n"),
        }),
      ))
    )
      return;
    deleting = true;
    try {
      await batchTrashFiles(ids);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.delete_failed());
    } finally {
      deleting = false;
    }
  }

  async function rename(slug: string, currentName: string) {
    const newName = await promptInput(
      m.rename(),
      m.enter_new_name(),
      currentName,
      100,
    );
    if (!newName || newName === currentName) return;
    try {
      await renameFile(slug, newName);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.rename_failed());
    }
  }

  async function move(ids: string[], targetParentSlug: string) {
    if (ids.length === 0) return;
    try {
      await Promise.all(ids.map((id) => moveFile(id, targetParentSlug)));
      toast.success(m.move_success({ count: ids.length }));
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.move_failed());
      throw e;
    }
  }

  async function toggleStar(slug: string, currentlyStarred: boolean) {
    try {
      await setStarred(slug, !currentlyStarred);
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.unstar_failed());
    }
  }

  async function setDirLock(file: NormalizedFile) {
    const password = await promptInput(
      "设置目录密码",
      `请输入「${file.name}」的目录密码（至少 4 位）`,
      undefined,
      128,
    );
    if (!password) return;
    try {
      await setDirectoryLock(file.id, password);
      toast.success("目录密码已设置");
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "设置目录密码失败");
    }
  }

  async function clearDirLock(file: NormalizedFile) {
    const password = await promptInput(
      "取消目录密码",
      `请输入「${file.name}」的目录密码`,
      undefined,
      128,
    );
    if (!password) return;
    try {
      await clearDirectoryLock(file.id, password);
      toast.success("目录密码已取消");
      await refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "取消目录密码失败");
    }
  }

  // --- Preview ---
  let previewFile = $state<{
    slug: string;
    name: string;
    mimeType: string;
    size: number;
  } | null>(null);
  let shareOpen = $state(false);

  let shareFiles = $state<NormalizedFile[]>([]);
  function onPreview(file: NormalizedFile) {
    previewFile = {
      slug: file.id,
      name: file.name,
      mimeType: file.mimeType || "",
      size: file.size,
    };
  }

  function onPreviewComplete(open: boolean) {
    if (!open) previewFile = null;
  }

  function onShare(file: NormalizedFile) {
    shareFiles = [file];
    shareOpen = true;
  }

  function onBatchShare(selectedFiles: NormalizedFile[]) {
    shareFiles = selectedFiles;
    shareOpen = true;
  }

  async function onAddToMedia(file: NormalizedFile) {
    try {
      const resp = await addToLibrary(file.id);
      if (resp.alreadyInLibrary) {
        toast.info(m.media_already_in_library());
      } else {
        toast.success(m.media_add_success());
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.media_add_failed());
    }
  }

  let dialogOpen = $derived(!!previewFile);
</script>

{#if $authReady && $user}
  <div class="space-y-4 rounded-xl border border-line bg-white p-4">
    {#if notFound}
      <div class="flex flex-col items-center justify-center py-24 text-center">
        <FileQuestionMark size={48} class="mb-4 text-ink-4" />
        <p class="mb-4 text-lg text-ink-3">{m.file_not_found()}</p>
        <button
          type="button"
          onclick={() => {
            notFound = false;
            void goto("/files/all", { keepFocus: true });
          }}
          class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
        >
          {m.back_to_root()}
        </button>
      </div>
    {:else if loading && files.length === 0}
      <div class="flex items-center justify-center py-24">
        <LoaderCircle size={24} class="animate-spin text-ink-4" />
      </div>
    {:else}
      <!-- Breadcrumb -->
      {#if currentSlug}
        <Breadcrumb
          bind:this={breadcrumbRef}
          items={crumbs}
          onNavigate={(id) => {
            void goto("/files/all/" + id, { keepFocus: true, noScroll: true });
          }}
          onHome={() => {
            void goto("/files/all", { keepFocus: true, noScroll: true });
          }}
        />
      {/if}

      <!-- Toolbar -->
      <FilesToolbar
        {sortBy}
        {sortDir}
        {viewMode}
        onSort={setSort}
        onViewModeChange={setViewMode}
        onUploadFiles={() => fileInput?.click()}
        onUploadFolder={() => folderInput?.click()}
        onCreateDir={createDir}
        onUploadFromURL={() => {
          remoteUploadOpen = true;
        }}
        onUploadText={() => {
          textUploadOpen = true;
        }}
      />
      <input
        bind:this={fileInput}
        type="file"
        multiple
        class="hidden"
        onchange={upload.onPick}
      />
      <input
        bind:this={folderInput}
        type="file"
        webkitdirectory
        class="hidden"
        onchange={upload.onPickFolder}
      />

      <!-- File list -->
      <div class="relative">
        {#if deleting}
          <div
            class="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-white/60 backdrop-blur-[1px]"
          >
            <LoaderCircle size={24} class="animate-spin text-ink-4" />
          </div>
        {/if}
        <FileListView
          files={normalizedFiles}
          {viewMode}
          {loading}
          directoryId={currentSlug}
          currentPath={crumbs}
          includeSystemDirs={getShowSystemDirs()}
          downloadUrlFn={downloadUrl}
          emptyMessage={currentSlug ? m.dir_empty() : m.no_files()}
          onNavigateDir={navigateToDir}
          onStar={toggleStar}
          {onPreview}
          onRename={rename}
          onDelete={remove}
          onBatchDelete={batchRemove}
          {onBatchShare}
          onMove={move}
          {onAddToMedia}
          {onShare}
          onSetDirectoryLock={setDirLock}
          onClearDirectoryLock={clearDirLock}
          onForceDeleteDir={forceRemoveDir}
          {loadFolderSummary}
        />
      </div>

      {#if files.length > 0}
        <div class="flex items-center justify-between text-xs text-ink-4">
          <span>{m.total_files({ total })}</span>
          {#if files.length < total}
            <button
              type="button"
              onclick={loadMore}
              disabled={loadingMore}
              class="text-ink-3 transition-colors hover:text-ink-2 disabled:opacity-50"
            >
              {loadingMore ? m.loading() : m.load_more()}
            </button>
          {/if}
        </div>
      {/if}
    {/if}
  </div>
{/if}

<ShareDialog bind:open={shareOpen} files={shareFiles} />

<DrivePreview
  id={previewFile!.slug}
  name={previewFile!.name}
  mimeType={previewFile!.mimeType}
  size={previewFile!.size}
  bind:open={dialogOpen}
  onOpenChangeComplete={onPreviewComplete}
/>

<FolderUploadDialog
  files={upload.folderDialogFiles}
  open={upload.folderDialogOpen}
  loading={upload.folderDialogLoading}
  onConfirm={upload.onFolderConfirm}
  onCancel={() => {
    upload.folderDialogOpen = false;
  }}
/>

<RemoteUploadDialog bind:open={remoteUploadOpen} parentSlug={currentSlug} />

<TextUploadDialog
  bind:open={textUploadOpen}
  targetLabel={pasteTargetLabel}
  onConfirm={async (file) => {
    await upload.enqueueFiles([file]);
  }}
  onCancel={() => {
    textUploadOpen = false;
  }}
/>

<ConflictDialog
  bind:open={conflictDialogOpen}
  conflicts={conflictDialogConflicts}
  onResolve={onConflictResolve}
  onCancel={onConflictCancel}
/>

<PasteUploadProvider
  targetLabel={pasteTargetLabel}
  onUpload={(files) => upload.enqueueFiles(files)}
/>

{@render children()}
