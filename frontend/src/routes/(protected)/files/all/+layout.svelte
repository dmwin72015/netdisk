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
    getFolderSummary,
    type FileItem,
  } from "$lib/api/files";
  import { ApiError } from "$lib/api/client";
  import type { NormalizedFile } from "$lib/types/file";
  import { normalizeFileItem } from "$lib/types/adapters";
  import { addToLibrary } from "$lib/api/media";
  import { toast } from "svelte-sonner";
  import DrivePreview from "$lib/components/DrivePreview.svelte";
  import FileBrowserContent from "$lib/components/files/FileBrowserContent.svelte";
  import ShareDialog from "$lib/components/files/ShareDialog.svelte";
  import FolderUploadDialog from "$lib/components/files/FolderUploadDialog.svelte";
  import RemoteUploadDialog from "$lib/components/files/RemoteUploadDialog.svelte";
  import TextUploadDialog from "$lib/components/files/TextUploadDialog.svelte";
  import ConflictDialog from "$lib/components/files/ConflictDialog.svelte";
  import PasteUploadProvider from "$lib/components/files/PasteUploadProvider.svelte";
  import { getShowSystemDirs } from "$lib/stores/file-preferences.svelte";
  import { confirmDelete, promptInput } from "$lib/dialog";
  import type { createUploadManager as UploadMgrFn } from "$lib/upload-manager.svelte";
  type UploadManager = ReturnType<typeof UploadMgrFn>;
  import * as m from "$lib/paraglide/messages";
  import { persistedState } from "$lib/stores/state.svelte";
  import { ConflictManager } from "$lib/stores/conflict-manager.svelte";
  import type { SortField, ViewMode } from "$lib/components/files/FilesToolbar.svelte";
  import type Breadcrumb from "$lib/components/Breadcrumb.svelte";

  let { children } = $props();

  const PAGE_SIZE = 50;
  const upload = getContext<UploadManager>("upload");

  // --- Persisted preferences ---
  let viewMode = persistedState<ViewMode>("nd.files.view", "list");
  let sortBy = persistedState<SortField>("nd.files.sortBy", "created_at");
  let sortDir = persistedState<"ASC" | "DESC">("nd.files.sortDir", "DESC");

  // --- File listing ---
  let files = $state<FileItem[]>([]);
  let normalizedFiles = $derived(files.map(normalizeFileItem));
  let total = $state(0);
  let loading = $state(true);
  let loadingMore = $state(false);
  let deleting = $state(false);
  let notFound = $state(false);
  let refreshId = 0;
  let loadingRequestId = 0;

  // --- Breadcrumb / Navigation ---
  let currentSlug = $state("");
  let crumbs = $state<{ id: string; name: string }[]>([]);
  let breadcrumbEl: Breadcrumb | undefined = $state();
  let pasteTargetLabel = $derived(crumbs.at(-1)?.name ?? m.nav_files());

  // --- Dialogs ---
  let remoteUploadOpen = $state(false);
  let textUploadOpen = $state(false);
  let previewFile = $state<{
    slug: string; name: string; mimeType: string; size: number;
  } | null>(null);
  let shareOpen = $state(false);
  let shareFiles = $state<NormalizedFile[]>([]);

  // --- Conflict resolution ---
  const conflicts = new ConflictManager();

  // --- Upload handlers ---
  $effect(() => {
    upload.updateMaxConcurrent(1);
    upload.setGetCurrentSlug(() => currentSlug);
    upload.setOnCompleted(() => refresh(false, true));
    upload.setOnNameConflicts((c) => conflicts.onUploadConflicts(c));
  });

  if (browser) {
    $effect(() => {
      const handler = () => fileUploadTrigger();
      window.addEventListener("nd:open-file-upload", handler);
      return () => window.removeEventListener("nd:open-file-upload", handler);
    });
  }

  function fileUploadTrigger() {
    const input = document.querySelector<HTMLInputElement>('input[type="file"][multiple]');
    input?.click();
  }

  // --- API helpers ---
  async function doFetch(pageNum: number) {
    return listFiles(
      currentSlug || undefined,
      pageNum,
      PAGE_SIZE,
      undefined,
      undefined,
      sortBy.current,
      sortDir.current,
      false,
      getShowSystemDirs(),
    );
  }

  // --- Navigation ---
  async function unlockDir(slug: string, name?: string) {
    const password = await promptInput(
      "目录密码",
      `请输入${name ? `「${name}」` : "目录"}的密码`,
      undefined,
      128,
    );
    if (!password) return false;
    try {
      await unlockDirectory(slug, password, 1);
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

  function onHome() {
    void goto("/files/all", { keepFocus: true });
  }

  async function fetchBreadcrumb(dirSlug: string) {
    if (!dirSlug) { crumbs = []; return; }
    try {
      crumbs = (await getBreadcrumb(dirSlug)).map((b) => ({ id: b.slug, name: b.fileName }));
    } catch {
      crumbs = [{ id: dirSlug, name: dirSlug }];
    }
  }

  afterNavigate(() => {
    const slug = page.params.slug ?? "";
    if (slug !== currentSlug) {
      currentSlug = slug;
      breadcrumbEl?.collapse();
    }
    void fetchBreadcrumb(currentSlug);
    void refresh(true);
  });

  // --- File listing ---
  async function refresh(showLoading = false, force = false) {
    if (!$user) return;
    const id = ++refreshId;
    if (showLoading) { loadingRequestId = id; loading = true; }
    loadingMore = false;
    notFound = false;
    try {
      const data = await doFetch(1);
      if (id !== refreshId && !force) return;
      files = data.files;
      total = data.total;
    } catch (e) {
      if (id !== refreshId && !force) return;
      if (e instanceof ApiError && e.status === 404) {
        notFound = true;
      } else if (e instanceof ApiError && e.status === 423 && currentSlug) {
        if (await unlockDir(currentSlug, crumbs.at(-1)?.name)) {
          void refresh(showLoading, force);
        }
      } else {
        toast.error(e instanceof Error ? e.message : m.load_failed());
      }
    } finally {
      if (id === loadingRequestId) loading = false;
    }
  }

  async function loadMore() {
    if (loadingMore) return;
    loadingMore = true;
    const id = ++refreshId;
    try {
      const pageNum = Math.floor(files.length / PAGE_SIZE) + 1;
      const data = await doFetch(pageNum);
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
    const data = await getFolderSummary(slug);
    return { fileCount: data.fileCount, folderCount: data.folderCount, size: data.totalSize };
  }

  // --- File operations ---
  function setViewMode(mode: ViewMode) { viewMode.current = mode; }

  function setSort(field: SortField) {
    if (sortBy.current === field) {
      sortDir.current = sortDir.current === "ASC" ? "DESC" : "ASC";
    } else {
      sortBy.current = field;
      sortDir.current = field === "file_name" ? "ASC" : "DESC";
    }
    void refresh(true);
  }

  async function createDir() {
    const name = (await promptInput(m.new_folder(), m.enter_folder_name(), undefined, 100))?.trim();
    if (!name) return;
    if (files.some((f) => f.isDir && f.fileName === name)) {
      toast.error(m.dir_already_exists()); return;
    }
    try {
      await mkdir(name, currentSlug || undefined);
      await refresh();
    } catch (e) { toast.error(e instanceof Error ? e.message : m.create_dir_failed()); }
  }

  async function remove(slug: string, name: string) {
    if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
    deleting = true;
    try { await trashFile(slug); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : m.delete_failed()); }
    finally { deleting = false; }
  }

  async function forceRemoveDir(file: NormalizedFile) {
    if (!(await confirmDelete(`确认强制删除目录「${file.name}」及其所有内容？此操作不可恢复。`))) return;
    deleting = true;
    try { await forceDeleteDir(file.id); toast.success(`目录「${file.name}」已强制删除`); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : "强制删除失败"); }
    finally { deleting = false; }
  }

  async function batchRemove(ids: string[]) {
    const names = ids.map((id) => normalizedFiles.find((f) => f.id === id)).filter(Boolean) as NormalizedFile[];
    if (names.length === 0) return;
    if (!(await confirmDelete(m.confirm_delete_multiple({ count: String(names.length), names: names.map((f) => f.name).join("\n") })))) return;
    deleting = true;
    try { await batchTrashFiles(ids); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : m.delete_failed()); }
    finally { deleting = false; }
  }

  async function rename(slug: string, currentName: string) {
    const newName = await promptInput(m.rename(), m.enter_new_name(), currentName, 100);
    if (!newName || newName === currentName) return;
    try { await renameFile(slug, newName); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : m.rename_failed()); }
  }

  async function move(ids: string[], targetParentSlug: string) {
    if (ids.length === 0) return;
    try {
      await Promise.all(ids.map((id) => moveFile(id, targetParentSlug)));
      toast.success(m.move_success({ count: ids.length }));
      await refresh();
    } catch (e) { toast.error(e instanceof Error ? e.message : m.move_failed()); throw e; }
  }

  async function toggleStar(slug: string, currentlyStarred: boolean) {
    try { await setStarred(slug, !currentlyStarred); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : m.unstar_failed()); }
  }

  async function setDirLock(file: NormalizedFile) {
    const password = await promptInput("设置目录密码", `请输入「${file.name}」的目录密码（至少 4 位）`, undefined, 128);
    if (!password) return;
    try { await setDirectoryLock(file.id, password); toast.success("目录密码已设置"); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : "设置目录密码失败"); }
  }

  async function clearDirLock(file: NormalizedFile) {
    const password = await promptInput("取消目录密码", `请输入「${file.name}」的目录密码`, undefined, 128);
    if (!password) return;
    try { await clearDirectoryLock(file.id, password); toast.success("目录密码已取消"); await refresh(); }
    catch (e) { toast.error(e instanceof Error ? e.message : "取消目录密码失败"); }
  }

  // --- Preview / Share ---
  function onPreview(file: NormalizedFile) {
    previewFile = { slug: file.id, name: file.name, mimeType: file.mimeType || "", size: file.size };
  }

  function onPreviewComplete(open: boolean) { if (!open) previewFile = null; }

  function onShare(file: NormalizedFile) { shareFiles = [file]; shareOpen = true; }

  function onBatchShare(files: NormalizedFile[]) { shareFiles = files; shareOpen = true; }

  async function onAddToMedia(file: NormalizedFile) {
    try {
      const resp = await addToLibrary(file.id);
      toast.success(resp.alreadyInLibrary ? m.media_already_in_library() : m.media_add_success());
    } catch (e) { toast.error(e instanceof Error ? e.message : m.media_add_failed()); }
  }
</script>

{#if $authReady && $user}
  <div class="space-y-4 rounded-xl border border-line bg-white p-4">
    <FileBrowserContent
      files={normalizedFiles}
      {total}
      {loading}
      {loadingMore}
      {notFound}
      {currentSlug}
      {crumbs}
      viewMode={viewMode.current}
      sortBy={sortBy.current}
      sortDir={sortDir.current}
      {deleting}
      includeSystemDirs={getShowSystemDirs()}
      downloadUrlFn={downloadUrl}
      {upload}
      {onHome}
      onNavigateDir={navigateToDir}
      onSetViewMode={setViewMode}
      onSetSort={setSort}
      onLoadMore={loadMore}
      {onPreview}
      onStar={toggleStar}
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
      onCreateDir={createDir}
      onUploadFromURL={() => { remoteUploadOpen = true; }}
      onUploadText={() => { textUploadOpen = true; }}
      {loadFolderSummary}
    />
  </div>
{/if}

<ShareDialog bind:open={shareOpen} files={shareFiles} />

<DrivePreview
  id={previewFile!.slug}
  name={previewFile!.name}
  mimeType={previewFile!.mimeType}
  size={previewFile!.size}
  open={previewFile !== null}
  onOpenChangeComplete={onPreviewComplete}
/>

<FolderUploadDialog
  files={upload.folderDialogFiles}
  open={upload.folderDialogOpen}
  loading={upload.folderDialogLoading}
  onConfirm={upload.onFolderConfirm}
  onCancel={() => { upload.folderDialogOpen = false; }}
/>

<RemoteUploadDialog bind:open={remoteUploadOpen} parentSlug={currentSlug} />

<TextUploadDialog
  bind:open={textUploadOpen}
  targetLabel={pasteTargetLabel}
  onConfirm={async (file) => { await upload.enqueueFiles([file]); }}
  onCancel={() => { textUploadOpen = false; }}
/>

<ConflictDialog
  bind:open={conflicts.open}
  conflicts={conflicts.conflicts}
  onResolve={(results) => conflicts.finish(results)}
  onCancel={() => conflicts.cancel()}
/>

<PasteUploadProvider
  targetLabel={pasteTargetLabel}
  onUpload={(files) => upload.enqueueFiles(files)}
/>

{@render children()}
