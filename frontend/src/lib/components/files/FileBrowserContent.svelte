<script lang="ts">
  import FileListView from "./FileListView.svelte";
  import FilesToolbar, {
    type SortField,
    type ViewMode,
  } from "./FilesToolbar.svelte";
  import Breadcrumb from "../Breadcrumb.svelte";
  import { FileQuestionMark, LoaderCircle } from "@lucide/svelte";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";

  type FolderSummary = { fileCount: number; folderCount: number; size: number };

  let {
    files,
    total,
    loading,
    loadingMore,
    notFound,
    currentSlug,
    crumbs,
    viewMode,
    sortBy,
    sortDir,
    deleting,
    includeSystemDirs,
    downloadUrlFn,
    upload,
    onNavigateDir,
    onHome,
    onSetViewMode,
    onSetSort,
    onLoadMore,
    onPreview,
    onStar,
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
    onCreateDir,
    onUploadFromURL,
    onUploadText,
    loadFolderSummary,
  }: {
    files: NormalizedFile[];
    total: number;
    loading: boolean;
    loadingMore: boolean;
    notFound: boolean;
    currentSlug: string;
    crumbs: { id: string; name: string }[];
    viewMode: ViewMode;
    sortBy: SortField;
    sortDir: "ASC" | "DESC";
    deleting: boolean;
    includeSystemDirs: boolean;
    downloadUrlFn: (id: string) => string;
    upload: {
      onPick: (e: Event) => void;
      onPickFolder: (e: Event) => void;
    };
    onNavigateDir: (slug: string) => void;
    onHome: () => void;
    onSetViewMode: (mode: ViewMode) => void;
    onSetSort: (field: SortField) => void;
    onLoadMore: () => void;
    onPreview: (file: NormalizedFile) => void;
    onStar: (slug: string, starred: boolean) => void;
    onRename: (slug: string, name: string) => void;
    onDelete: (slug: string, name: string) => void;
    onBatchDelete: (ids: string[]) => void;
    onBatchShare: (files: NormalizedFile[]) => void;
    onMove: (ids: string[], slug: string) => Promise<void>;
    onAddToMedia: (file: NormalizedFile) => void;
    onShare: (file: NormalizedFile) => void;
    onSetDirectoryLock: (file: NormalizedFile) => void;
    onClearDirectoryLock: (file: NormalizedFile) => void;
    onForceDeleteDir: (file: NormalizedFile) => void;
    onCreateDir: () => void;
    onUploadFromURL?: () => void;
    onUploadText?: () => void;
    loadFolderSummary?: (slug: string) => Promise<FolderSummary>;
  } = $props();

  let fileInput: HTMLInputElement | undefined = $state();
  let folderInput: HTMLInputElement | undefined = $state();
  let breadcrumbRef: Breadcrumb | undefined = $state();
</script>

{#if notFound}
  <div class="flex flex-col items-center justify-center py-24 text-center">
    <FileQuestionMark size={48} class="mb-4 text-ink-4" />
    <p class="mb-4 text-lg text-ink-3">{m.file_not_found()}</p>
    <button
      type="button"
      onclick={onHome}
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
  {#if currentSlug}
    <Breadcrumb
      bind:this={breadcrumbRef}
      items={crumbs}
      onNavigate={(id) => onNavigateDir(id)}
      {onHome}
    />
  {/if}

  <FilesToolbar
    {sortBy}
    {sortDir}
    {viewMode}
    onSort={onSetSort}
    onViewModeChange={onSetViewMode}
    onUploadFiles={() => fileInput?.click()}
    onUploadFolder={() => folderInput?.click()}
    {onCreateDir}
    {onUploadFromURL}
    {onUploadText}
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

  <div class="relative">
    {#if deleting}
      <div
        class="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-white/60 backdrop-blur-[1px]"
      >
        <LoaderCircle size={24} class="animate-spin text-ink-4" />
      </div>
    {/if}
    <FileListView
      {files}
      {viewMode}
      {loading}
      directoryId={currentSlug}
      currentPath={crumbs}
      {includeSystemDirs}
      {downloadUrlFn}
      emptyMessage={currentSlug ? m.dir_empty() : m.no_files()}
      {onNavigateDir}
      {onPreview}
      {onStar}
      {onRename}
      {onDelete}
      {onBatchDelete}
      {onBatchShare}
      {onMove}
      {onAddToMedia}
      {onShare}
      {onSetDirectoryLock}
      {onClearDirectoryLock}
      {onForceDeleteDir}
      {loadFolderSummary}
    />
  </div>

  {#if files.length > 0}
    <div class="flex items-center justify-between text-xs text-ink-4">
      <span>{m.total_files({ total })}</span>
      {#if files.length < total}
        <button
          type="button"
          onclick={onLoadMore}
          disabled={loadingMore}
          class="text-ink-3 transition-colors hover:text-ink-2 disabled:opacity-50"
        >
          {loadingMore ? m.loading() : m.load_more()}
        </button>
      {/if}
    </div>
  {/if}
{/if}
