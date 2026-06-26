import type { FileItem } from "$lib/api/files";
import {
  listFiles,
  mkdir,
  trashFile,
  batchTrashFiles,
  renameFile,
  setStarred,
  moveFile,
  forceDeleteDir,
  getBreadcrumb,
  getFolderSummary,
  downloadUrl,
} from "$lib/api/files";
import { addToLibrary } from "$lib/api/media";
import { ApiError } from "$lib/api/client";
import type { NormalizedFile } from "$lib/types/file";
import { normalizeFileItem } from "$lib/types/adapters";
import { lockManager } from "./lockManager.svelte";
import { settingsManager } from "./settingsManager.svelte";
import { persistedState } from "$lib/stores/state.svelte";
import { debounce } from "$lib/utils/debounce";
import { confirmDelete, promptInput } from "$lib/dialog";
import { toast } from "svelte-sonner";
import * as m from "$lib/paraglide/messages";

export type SortField = "file_name" | "created_at" | "updated_at" | "file_size";
export type ViewMode = "list" | "grid";

const PAGE_SIZE = 50;

class FileManager {
  // --- File data ---
  files = $state<FileItem[]>([]);
  total = $state(0);
  loading = $state(true);
  loadingMore = $state(false);
  notFound = $state(false);

  // --- Navigation ---
  currentSlug = $state("");
  crumbs = $state<{ id: string; name: string }[]>([]);

  // --- Selection ---
  selectedIds = $state<Set<string>>(new Set());

  // --- View preferences (persisted) ---
  viewMode = persistedState<ViewMode>("nd.files.view", "list");
  sortBy = persistedState<SortField>("nd.files.sortBy", "created_at");
  sortDir = persistedState<"ASC" | "DESC">("nd.files.sortDir", "DESC");

  // --- Internal ---
  private refreshId = 0;
  private loadingRequestId = 0;

  /** Normalized files. Use this in components. */
  get normalizedFiles(): NormalizedFile[] {
    return this.files.map((f) => normalizeFileItem(f));
  }

  /** Whether all selectable (non-system) files are selected. */
  get allSelected(): boolean {
    const selectable = this.normalizedFiles.filter((f) => !f.isSystem);
    return (
      selectable.length > 0 &&
      selectable.every((f) => this.selectedIds.has(f.id))
    );
  }

  /** Whether any files are selected. */
  get hasSelection(): boolean {
    return this.selectedIds.size > 0;
  }

  /** Breadcrumb label for the current directory. */
  get currentDirLabel(): string {
    return this.crumbs.at(-1)?.name ?? m.nav_files();
  }

  // --- Navigation ---

  setSlug(slug: string, collapseBreadcrumb?: () => void) {
    if (slug !== this.currentSlug) {
      this.currentSlug = slug;
      collapseBreadcrumb?.();
    }
    void this.fetchBreadcrumb(this.currentSlug);
    void this.refresh(true);
  }

  async navigateToDir(
    slug: string,
  ): Promise<"navigated" | "unlocked" | "blocked"> {
    const file = this.normalizedFiles.find((item) => item.id === slug);
    if (file && lockManager.isEffectivelyLocked(file)) {
      let unlocked: boolean;
      try {
        unlocked = await lockManager.unlock(slug, file.name);
      } catch (e) {
        toast.error(e instanceof Error ? e.message : "目录密码错误");
        return "blocked";
      }
      if (!unlocked) return "blocked";
      toast.success("目录已解锁");
      return "unlocked";
    }
    this.loading = true;
    this.files = [];
    return "navigated";
  }

  async fetchBreadcrumb(dirSlug: string) {
    if (!dirSlug) {
      this.crumbs = [];
      return;
    }
    try {
      this.crumbs = (await getBreadcrumb(dirSlug)).map((b) => ({
        id: b.slug,
        name: b.fileName,
      }));
    } catch {
      this.crumbs = [{ id: dirSlug, name: dirSlug }];
    }
  }

  // --- File listing ---

  async refresh(showLoading = false, force = false) {
    const id = ++this.refreshId;
    if (showLoading) {
      this.loadingRequestId = id;
      this.loading = true;
    }
    this.loadingMore = false;
    this.notFound = false;
    try {
      const data = await this.doFetch(1);
      if (id !== this.refreshId && !force) return;
      this.files = data.files;
      this.total = data.total;
      lockManager.syncFromFiles(data.files);
    } catch (e) {
      if (id !== this.refreshId && !force) return;
      if (e instanceof ApiError && e.status === 404) {
        this.notFound = true;
      } else if (
        e instanceof ApiError &&
        e.status === 423 &&
        this.currentSlug
      ) {
        let unlocked: boolean;
        try {
          unlocked = await lockManager.unlock(
            this.currentSlug,
            this.crumbs.at(-1)?.name,
          );
        } catch {
          toast.error("目录密码错误");
          return;
        }
        if (!unlocked) return;
        toast.success("目录已解锁");
        void this.refresh(showLoading, force);
      } else {
        toast.error(e instanceof Error ? e.message : m.load_failed());
      }
    } finally {
      if (id === this.loadingRequestId) this.loading = false;
    }
  }

  async loadMore() {
    if (this.loadingMore) return;
    this.loadingMore = true;
    const id = ++this.refreshId;
    try {
      const pageNum = Math.floor(this.files.length / PAGE_SIZE) + 1;
      const data = await this.doFetch(pageNum);
      if (id !== this.refreshId) return;
      this.files = [...this.files, ...data.files];
    } catch (e) {
      if (id !== this.refreshId) return;
      toast.error(e instanceof Error ? e.message : m.load_more_failed());
    } finally {
      if (id === this.refreshId) this.loadingMore = false;
    }
  }

  // --- View preferences ---

  setViewMode(mode: ViewMode) {
    this.viewMode.current = mode;
  }

  private debouncedRefresh = debounce(() => this.refresh(true), 200);

  setSort(field: SortField) {
    if (this.sortBy.current === field) {
      this.sortDir.current = this.sortDir.current === "ASC" ? "DESC" : "ASC";
    } else {
      this.sortBy.current = field;
      this.sortDir.current = field === "file_name" ? "ASC" : "DESC";
    }
    this.debouncedRefresh();
  }

  // --- CRUD operations (optimistic) ---

  async createDir() {
    const name = (
      await promptInput(m.new_folder(), m.enter_folder_name(), undefined, 100)
    )?.trim();
    if (!name) return;
    if (this.files.some((f) => f.isDir && f.fileName === name)) {
      toast.error(m.dir_already_exists());
      return;
    }
    try {
      await mkdir(name, this.currentSlug || undefined);
      await this.refresh();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.create_dir_failed());
    }
  }

  async remove(slug: string, name: string) {
    const file = this.normalizedFiles.find((f) => f.slug === slug);
    if (file && lockManager.isEffectivelyLocked(file)) {
      toast.error("目录已加锁，无法删除");
      return;
    }
    if (!(await confirmDelete(m.confirm_delete_file({ name })))) return;
    this.files = this.files.filter((f) => f.slug !== slug);
    this.total = Math.max(0, this.total - 1);
    try {
      await trashFile(slug);
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : m.delete_failed());
    }
  }

  async batchRemove(ids: string[]) {
    const lockedNames = ids
      .map((id) => this.normalizedFiles.find((f) => f.id === id))
      .filter(
        (f): f is NormalizedFile => !!f && lockManager.isEffectivelyLocked(f),
      )
      .map((f) => f.name);
    const unlockedIds = ids.filter((id) => {
      const f = this.normalizedFiles.find((f) => f.id === id);
      return !f || !lockManager.isEffectivelyLocked(f);
    });
    if (unlockedIds.length === 0) {
      toast.error("所有选中的目录均已加锁");
      return;
    }
    const names = unlockedIds
      .map((id) => this.normalizedFiles.find((f) => f.id === id))
      .filter(Boolean) as NormalizedFile[];
    if (names.length === 0) return;
    if (lockedNames.length > 0) {
      toast.info(`已跳过 ${lockedNames.length} 个加锁目录`);
    }
    if (
      !(await confirmDelete(
        m.confirm_delete_multiple({
          count: String(names.length),
          names: names.map((f) => f.name).join("\n"),
        }),
      ))
    )
      return;
    const deleteSet = new Set(unlockedIds);
    this.files = this.files.filter((f) => !deleteSet.has(f.slug));
    this.total = Math.max(0, this.total - unlockedIds.length);
    try {
      await batchTrashFiles(unlockedIds);
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : m.delete_failed());
    }
  }

  async rename(slug: string, currentName: string) {
    const file = this.normalizedFiles.find((f) => f.slug === slug);
    if (file && lockManager.isEffectivelyLocked(file)) {
      toast.error("目录已加锁，无法重命名");
      return;
    }
    const newName = await promptInput(
      m.rename(),
      m.enter_new_name(),
      currentName,
      100,
    );
    if (!newName || newName === currentName) return;
    this.files = this.files.map((f) =>
      f.slug === slug ? { ...f, fileName: newName } : f,
    );
    try {
      await renameFile(slug, newName);
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : m.rename_failed());
    }
  }

  async move(ids: string[], targetParentSlug: string) {
    const lockedNames = ids
      .map((id) => this.normalizedFiles.find((f) => f.id === id))
      .filter(
        (f): f is NormalizedFile => !!f && lockManager.isEffectivelyLocked(f),
      )
      .map((f) => f.name);
    const unlockedIds = ids.filter((id) => {
      const f = this.normalizedFiles.find((f) => f.id === id);
      return !f || !lockManager.isEffectivelyLocked(f);
    });
    if (unlockedIds.length === 0) {
      toast.error("所有选中的目录均已加锁，无法移动");
      return;
    }
    if (lockedNames.length > 0) {
      toast.info(`已跳过 ${lockedNames.length} 个加锁目录`);
    }
    if (unlockedIds.length === 0) return;
    const isMovingOut = targetParentSlug !== this.currentSlug;
    if (isMovingOut) {
      const moveSet = new Set(unlockedIds);
      this.files = this.files.filter((f) => !moveSet.has(f.slug));
      this.total = Math.max(0, this.total - unlockedIds.length);
    }
    try {
      await Promise.all(
        unlockedIds.map((id) => moveFile(id, targetParentSlug)),
      );
      toast.success(m.move_success({ count: unlockedIds.length }));
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : m.move_failed());
      throw e;
    }
  }

  async toggleStar(slug: string, currentlyStarred: boolean) {
    this.files = this.files.map((f) =>
      f.slug === slug ? { ...f, isStarred: !currentlyStarred } : f,
    );
    try {
      await setStarred(slug, !currentlyStarred);
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : m.unstar_failed());
    }
  }

  async forceRemoveDir(file: NormalizedFile) {
    if (lockManager.isEffectivelyLocked(file)) {
      toast.error("目录已加锁，无法强制删除");
      return;
    }
    if (
      !(await confirmDelete(
        `确认强制删除目录「${file.name}」及其所有内容？此操作不可恢复。`,
      ))
    )
      return;
    this.files = this.files.filter((f) => f.slug !== file.slug);
    this.total = Math.max(0, this.total - 1);
    try {
      await forceDeleteDir(file.id);
      toast.success(`目录「${file.name}」已强制删除`);
      void this.refresh(false);
    } catch (e) {
      await this.refresh(true);
      toast.error(e instanceof Error ? e.message : "强制删除失败");
    }
  }

  // --- Media ---

  async addToMedia(file: NormalizedFile) {
    try {
      const resp = await addToLibrary(file.id);
      toast.success(
        resp.alreadyInLibrary
          ? m.media_already_in_library()
          : m.media_add_success(),
      );
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.media_add_failed());
    }
  }

  // --- Selection ---

  toggleSelect(id: string) {
    const file = this.normalizedFiles.find((f) => f.id === id);
    if (file?.isSystem) return;
    if (this.selectedIds.has(id)) this.selectedIds.delete(id);
    else this.selectedIds.add(id);
  }

  toggleSelectAll() {
    if (this.allSelected) {
      this.selectedIds = new Set();
    } else {
      const selectable = this.normalizedFiles
        .filter((f) => !f.isSystem)
        .map((f) => f.id);
      this.selectedIds = new Set(selectable);
    }
  }

  clearSelection() {
    this.selectedIds = new Set();
  }

  // --- Folder summary ---

  async loadFolderSummary(slug: string) {
    const data = await getFolderSummary(slug);
    return {
      fileCount: data.fileCount,
      folderCount: data.folderCount,
      size: data.totalSize,
    };
  }

  // --- Download URL helper ---

  getDownloadUrl(slug: string): string {
    return downloadUrl(slug);
  }

  // --- Internal ---

  private async doFetch(pageNum: number) {
    return listFiles(
      this.currentSlug || undefined,
      pageNum,
      PAGE_SIZE,
      undefined,
      undefined,
      this.sortBy.current,
      this.sortDir.current,
      false,
      settingsManager.showSystemDirs,
    );
  }
}

export const fileManager = new FileManager();
