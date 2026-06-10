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
  }: {
    sortBy: SortField;
    sortDir: "ASC" | "DESC";
    viewMode: ViewMode;
    onSort: (field: SortField) => void;
    onViewModeChange: (mode: ViewMode) => void;
    onUploadFiles: () => void;
    onUploadFolder: () => void;
    onCreateDir: () => void;
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
  class="flex flex-col gap-3 rounded-2xl lg:flex-row lg:items-center lg:justify-end"
>
  <div class="flex items-center gap-2">
    <Dropdown
      triggerClass="flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-2.5 text-sm text-gray-600 transition-colors hover:border-gray-300 hover:bg-gray-50"
      contentClass="min-w-[144px]"
    >
      {#snippet trigger()}
        <ArrowUpDown size={14} />
        <span class="hidden sm:inline"
          >{sortOptions.find((o) => o.field === sortBy)?.label()}</span
        >
        {#if sortDir === "ASC"}
          <ArrowUp size={14} class="text-blue-500" />
        {:else}
          <ArrowDown size={14} class="text-blue-500" />
        {/if}
      {/snippet}

      {#each sortOptions as opt (opt.field)}
        <DropdownBase.Item onSelect={() => onSort(opt.field)}>
          <span class={sortBy === opt.field ? "font-medium text-gray-900" : ""}
            >{opt.label()}</span
          >
          {#if sortBy === opt.field}
            {#if sortDir === "ASC"}
              <ArrowUp size={14} class="ml-auto text-blue-500" />
            {:else}
              <ArrowDown size={14} class="ml-auto text-blue-500" />
            {/if}
          {/if}
        </DropdownBase.Item>
      {/each}
    </Dropdown>

    <div class="flex overflow-hidden rounded-lg border border-gray-200">
      <button
        type="button"
        onclick={() => onViewModeChange("list")}
        class="p-1.5 transition-colors {viewMode === 'list'
          ? 'bg-blue-50 text-blue-600'
          : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}"
      >
        <LayoutList size={15} />
      </button>
      <button
        type="button"
        onclick={() => onViewModeChange("grid")}
        class="p-1.5 transition-colors {viewMode === 'grid'
          ? 'bg-blue-50 text-blue-600'
          : 'bg-white text-gray-400 hover:bg-gray-50 hover:text-gray-600'}"
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
        class="flex h-8 items-center overflow-hidden rounded-lg bg-blue-600 text-sm font-medium text-white shadow-sm transition-colors"
      >
        <button
          type="button"
          onclick={onUploadFiles}
          class="flex h-full items-center gap-1.5 bg-blue-600 px-3.5 hover:bg-blue-700 active:bg-blue-800"
        >
          <Upload size={15} />
          {m.upload_files()}
        </button>
        <Popover
          bind:open={showUploadMenu}
          triggerClass="flex h-full items-center px-1.5 bg-blue-600 hover:bg-blue-700 active:bg-blue-800"
          contentClass="w-auto min-w-40 p-1.5"
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
              class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-gray-700 outline-none hover:bg-gray-50 focus-visible:outline-none"
              onclick={() => {
                showUploadMenu = false;
                onUploadFiles();
              }}
            >
              <Upload size={15} class="text-gray-500" />
              {m.upload_files()}
            </button>
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-gray-700 outline-none hover:bg-gray-50 focus-visible:outline-none"
              onclick={() => {
                showUploadMenu = false;
                onUploadFolder();
              }}
            >
              <FolderOpen size={15} class="text-gray-500" />
              {m.upload_folder()}
            </button>
            <div class="my-1 border-t border-gray-100"></div>
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-gray-700 outline-none hover:bg-gray-50 focus-visible:outline-none"
              onclick={() => {
                showUploadMenu = false;
                onCreateDir();
              }}
            >
              <FolderPlus size={15} class="text-gray-500" />
              {m.new_folder()}
            </button>
          </div>
        </Popover>
      </div>
    </div>
  </div>
</div>
