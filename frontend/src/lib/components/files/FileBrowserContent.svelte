<script lang="ts">
  import FileListView from "./FileListView.svelte";
  import FilesToolbar from "./FilesToolbar.svelte";
  import Breadcrumb from "../Breadcrumb.svelte";
  import {
    Plus,
    Upload,
    FolderPlus,
    FolderOpen,
    ChevronDown,
    Globe,
    FileText,
  } from "@lucide/svelte";
  import { Popover } from "$lib/ui/popover";
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

  let showFabMenu = $state(false);
  let fabTimeout: ReturnType<typeof setTimeout>;

  function onFabEnter() {
    clearTimeout(fabTimeout);
    showFabMenu = true;
  }

  function onFabLeave() {
    fabTimeout = setTimeout(() => {
      showFabMenu = false;
    }, 150);
  }

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

  <FilesToolbar total={fileManager.total} />
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
    <div class="flex justify-center py-8">
      {#if fileManager.files.length < fileManager.total}
        <button
          type="button"
          onclick={() => fileManager.loadMore()}
          disabled={fileManager.loadingMore}
          class="text-xs text-ink-3 transition-colors hover:text-ink-2 disabled:opacity-50"
        >
          {fileManager.loadingMore ? m.loading() : m.load_more()}
        </button>
      {:else}
        <span class="text-xs text-ink-4">{m.no_more()}</span>
      {/if}
    </div>
  {/if}

  <!-- Floating Action Button -->
  <div
    class="fixed bottom-6 right-6 z-40"
    role="region"
    onmouseenter={onFabEnter}
    onmouseleave={onFabLeave}
  >
    <Popover
      bind:open={showFabMenu}
      triggerClass="flex h-12 w-12 items-center justify-center rounded-full bg-primary text-white shadow-pop transition-colors hover:bg-primary-hover active:bg-primary-active"
      contentClass="min-w-40 p-1.5"
      sideOffset={8}
      align="end"
    >
      {#snippet trigger()}
        <Plus size={22} />
      {/snippet}

      <div role="region" onmouseenter={onFabEnter} onmouseleave={onFabLeave}>
        <button
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-primary-soft hover:text-primary"
          onclick={() => {
            showFabMenu = false;
            onUploadFiles();
          }}
        >
          <Upload size={15} class="text-primary" />
          {m.upload_files()}
        </button>
        <button
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-primary-soft hover:text-primary"
          onclick={() => {
            showFabMenu = false;
            onUploadFolder();
          }}
        >
          <FolderOpen size={15} class="text-primary" />
          {m.upload_folder()}
        </button>
        {#if onUploadFromURL}
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-purple-50 hover:text-purple-600"
            onclick={() => {
              showFabMenu = false;
              onUploadFromURL();
            }}
          >
            <Globe size={15} class="text-purple-500" />
            {m.remote_upload()}
          </button>
        {/if}
        {#if onUploadText}
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-amber-50 hover:text-amber-600"
            onclick={() => {
              showFabMenu = false;
              onUploadText();
            }}
          >
            <FileText size={15} class="text-amber-500" />
            {m.paste_text()}
          </button>
        {/if}
        <div class="bg-line-soft mx-1 my-1 h-px"></div>
        <button
          type="button"
          class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-green-50 hover:text-green-600"
          onclick={() => {
            showFabMenu = false;
            fileManager.createDir();
          }}
        >
          <FolderPlus size={15} class="text-green-500" />
          {m.new_folder()}
        </button>
      </div>
    </Popover>
  </div>
{/if}
