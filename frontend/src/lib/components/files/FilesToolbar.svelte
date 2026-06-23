<script lang="ts">
  import {
    Upload,
    FolderPlus,
    FolderOpen,
    ChevronDown,
    LayoutGrid,
    LayoutList,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    Globe,
    FileText,
  } from "@lucide/svelte";
  import { Dropdown, DropdownBase } from "$lib/ui/dropdown";
  import { Popover } from "$lib/ui/popover";
  import * as m from "$lib/paraglide/messages";

  export type SortField =
    | "file_name"
    | "file_size"
    | "created_at"
    | "updated_at";
  export type ViewMode = "list" | "grid";

  const sortOptions: { field: SortField; label: () => string }[] = [
    { field: "file_name", label: () => m.sort_name() },
    { field: "file_size", label: () => m.sort_size() },
    { field: "created_at", label: () => m.sort_created() },
    { field: "updated_at", label: () => m.sort_updated() },
  ];

  let {
    sortBy,
    sortDir,
    viewMode,
    onSort,
    onViewModeChange,
    onUploadFiles,
    onUploadFolder,
    onCreateDir,
    onUploadFromURL,
    onUploadText,
  }: {
    sortBy: SortField;
    sortDir: "ASC" | "DESC";
    viewMode: ViewMode;
    onSort: (field: SortField) => void;
    onViewModeChange: (mode: ViewMode) => void;
    onUploadFiles: () => void;
    onUploadFolder: () => void;
    onCreateDir: () => void;
    onUploadFromURL?: () => void;
    onUploadText?: () => void;
  } = $props();

  let showUploadMenu = $state(false);
  let menuTimeout: ReturnType<typeof setTimeout>;

  function onMenuEnter() {
    clearTimeout(menuTimeout);
    showUploadMenu = true;
  }

  function onMenuLeave() {
    menuTimeout = setTimeout(() => {
      showUploadMenu = false;
    }, 150);
  }
</script>

<div
  class="flex flex-col gap-3 rounded-xl lg:flex-row lg:items-center lg:justify-end"
>
  <div class="flex items-center gap-2">
    <Dropdown
      triggerClass="flex h-8 min-w-[120px] items-center justify-between gap-1.5 rounded-lg border border-line bg-white px-2.5 text-sm text-ink-3 transition-colors hover:border-line hover:bg-surface-muted"
      contentClass="min-w-40"
    >
      {#snippet trigger()}
        <span class="flex items-center gap-1.5">
          <ArrowUpDown size={14} />
          <span class="hidden sm:inline"
            >{sortOptions.find((o) => o.field === sortBy)?.label()}</span
          >
        </span>
        {#if sortDir === "ASC"}
          <ArrowUp size={14} class="text-primary" />
        {:else}
          <ArrowDown size={14} class="text-primary" />
        {/if}
      {/snippet}

      {#each sortOptions as opt (opt.field)}
        <DropdownBase.Item onSelect={() => onSort(opt.field)}>
          <span class={sortBy === opt.field ? "font-medium text-ink" : ""}
            >{opt.label()}</span
          >
          {#if sortBy === opt.field}
            {#if sortDir === "ASC"}
              <ArrowUp size={14} class="ml-auto text-primary" />
            {:else}
              <ArrowDown size={14} class="ml-auto text-primary" />
            {/if}
          {/if}
        </DropdownBase.Item>
      {/each}
    </Dropdown>

    <div class="flex overflow-hidden rounded-lg border border-line">
      <button
        type="button"
        onclick={() => onViewModeChange("list")}
        class="p-1.5 transition-colors {viewMode === 'list'
          ? 'bg-primary-soft text-primary'
          : 'bg-white text-ink-4 hover:bg-surface-muted hover:text-ink-3'}"
      >
        <LayoutList size={15} />
      </button>
      <button
        type="button"
        onclick={() => onViewModeChange("grid")}
        class="p-1.5 transition-colors {viewMode === 'grid'
          ? 'bg-primary-soft text-primary'
          : 'bg-white text-ink-4 hover:bg-surface-muted hover:text-ink-3'}"
      >
        <LayoutGrid size={15} />
      </button>
    </div>

    <!-- Upload split button -->
    <div
      class="relative"
      role="region"
      onmouseenter={onMenuEnter}
      onmouseleave={onMenuLeave}
    >
      <div
        class="flex h-8 items-center overflow-hidden rounded-lg bg-primary text-sm font-medium text-white transition-colors"
      >
        <button
          type="button"
          onclick={onUploadFiles}
          class="flex h-full items-center gap-1.5 bg-primary px-3.5 hover:bg-primary-hover active:bg-primary-active"
        >
          <Upload size={15} />
          {m.upload_files()}
        </button>
        <Popover
          bind:open={showUploadMenu}
          triggerClass="flex h-full items-center px-1.5 bg-primary hover:bg-primary-hover active:bg-primary-active"
          contentClass="min-w-40 p-1.5"
          sideOffset={4}
          align="end"
        >
          {#snippet trigger()}
            <ChevronDown size={14} />
          {/snippet}

          <div
            role="region"
            onmouseenter={onMenuEnter}
            onmouseleave={onMenuLeave}
          >
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-primary-soft hover:text-primary"
              onclick={() => {
                showUploadMenu = false;
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
                showUploadMenu = false;
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
                  showUploadMenu = false;
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
                  showUploadMenu = false;
                  onUploadText();
                }}
              >
                <FileText size={15} class="text-amber-500" />
                粘贴文本
              </button>
            {/if}
            <div class="bg-line-soft mx-1 my-1 h-px"></div>
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-green-50 hover:text-green-600"
              onclick={() => {
                showUploadMenu = false;
                onCreateDir();
              }}
            >
              <FolderPlus size={15} class="text-green-500" />
              {m.new_folder()}
            </button>
          </div>
        </Popover>
      </div>
    </div>
  </div>
</div>
