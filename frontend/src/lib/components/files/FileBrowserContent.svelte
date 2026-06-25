<script lang="ts">
  import FileListView from "./FileListView.svelte";
  import FilesToolbar from "./FilesToolbar.svelte";
  import Breadcrumb from "../Breadcrumb.svelte";
  import { FileQuestionMark } from "@lucide/svelte";
  import type { NormalizedFile } from "$lib/types/file";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { getContext } from "svelte";
  import type { createUploadManager } from "$lib/upload-manager.svelte";
  import * as m from "$lib/paraglide/messages";
  import { goto } from "$app/navigation";

  type UploadManager = ReturnType<typeof createUploadManager>;

  const upload = getContext<UploadManager>("upload");

  let {
    onBatchShare,
    onUploadFiles,
    onUploadFolder,
    onUploadFromURL,
    onUploadText,
  }: {
    onBatchShare?: (files: NormalizedFile[]) => void;
    onUploadFiles: () => void;
    onUploadFolder: () => void;
    onUploadFromURL?: () => void;
    onUploadText?: () => void;
  } = $props();

  function navigateHome() {
    void goto("/files/all", { keepFocus: true });
  }

  async function navigateToDir(slug: string) {
    const result = await fileManager.navigateToDir(slug);
    if (result === "navigated") {
      void goto("/files/all/" + slug, { keepFocus: true, noScroll: true });
    }
  }
</script>

{#if fileManager.notFound}
  <div class="flex flex-col items-center justify-center py-24 text-center">
    <FileQuestionMark size={48} class="mb-4 text-ink-4" />
    <p class="mb-4 text-lg text-ink-3">{m.file_not_found()}</p>
    <button
      type="button"
      onclick={navigateHome}
      class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-hover"
    >
      {m.back_to_root()}
    </button>
  </div>
{:else}
  <Breadcrumb
    items={fileManager.crumbs}
    onNavigate={(id) => navigateToDir(id)}
    onHome={navigateHome}
  />

  <FilesToolbar
    {onUploadFiles}
    {onUploadFolder}
    {onUploadFromURL}
    {onUploadText}
  />
  <input type="file" multiple class="hidden" onchange={upload.onPick} />
  <input
    type="file"
    multiple
    webkitdirectory
    class="hidden"
    onchange={upload.onPickFolder}
  />

  <div class="relative">
    <FileListView
      files={fileManager.normalizedFiles}
      loading={fileManager.loading}
      emptyMessage={fileManager.currentSlug ? m.dir_empty() : m.no_files()}
      onNavigateDir={navigateToDir}
      {onBatchShare}
    />
  </div>

  {#if fileManager.files.length > 0}
    <div class="flex items-center justify-between text-xs text-ink-4">
      <span>{m.total_files({ total: fileManager.total })}</span>
      {#if fileManager.files.length < fileManager.total}
        <button
          type="button"
          onclick={() => fileManager.loadMore()}
          disabled={fileManager.loadingMore}
          class="text-ink-3 transition-colors hover:text-ink-2 disabled:opacity-50"
        >
          {fileManager.loadingMore ? m.loading() : m.load_more()}
        </button>
      {/if}
    </div>
  {/if}
{/if}
