<script lang="ts">
  import {
    Check,
    LayoutGrid,
    LayoutList,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
  } from "@lucide/svelte";
  import { Dropdown, DropdownBase } from "$lib/ui/dropdown";
  import {
    fileManager,
    type SortField,
  } from "$lib/services/fileManager.svelte";
  import * as m from "$lib/paraglide/messages";

  const sortOptions: { field: SortField; label: () => string }[] = [
    { field: "file_name", label: () => m.sort_name() },
    { field: "file_size", label: () => m.sort_size() },
    { field: "created_at", label: () => m.sort_created() },
    { field: "updated_at", label: () => m.sort_updated() },
  ];

  let {
    total,
  }: {
    total: number;
  } = $props();
</script>

<div class="flex items-end justify-between my-3">
  <!-- Left: select all + count -->
  <div class="flex items-center gap-2 pl-1">
    <button
      type="button"
      onclick={() => fileManager.toggleSelectAll()}
      class="flex h-4 w-4 shrink-0 items-center justify-center rounded-full border transition-colors {fileManager.allSelected
        ? 'border-primary bg-primary text-white'
        : 'border-line hover:border-primary'}"
      aria-label={m.select_all()}
    >
      {#if fileManager.allSelected}
        <Check size={10} />
      {/if}
    </button>
    <span class="text-sm text-ink-2">
      {m.total_items({ total: String(total) })}
    </span>
  </div>

  <!-- Right: sort + view mode -->
  <div class="flex items-center gap-2">
    <Dropdown
      triggerClass="flex h-8 min-w-[120px] items-center justify-between gap-1.5 rounded-lg border border-line bg-white px-2.5 text-sm text-ink-3 transition-colors hover:border-line hover:bg-surface-muted"
      contentClass="min-w-40"
    >
      {#snippet trigger()}
        <span class="flex items-center gap-1.5">
          <ArrowUpDown size={14} />
          <span class="hidden sm:inline"
            >{sortOptions
              .find((o) => o.field === fileManager.sortBy.current)
              ?.label()}</span
          >
        </span>
        {#if fileManager.sortDir.current === "ASC"}
          <ArrowUp size={14} class="text-primary" />
        {:else}
          <ArrowDown size={14} class="text-primary" />
        {/if}
      {/snippet}

      {#each sortOptions as opt (opt.field)}
        <DropdownBase.Item onSelect={() => fileManager.setSort(opt.field)}>
          <span
            class={fileManager.sortBy.current === opt.field
              ? "font-medium text-ink"
              : ""}>{opt.label()}</span
          >
          {#if fileManager.sortBy.current === opt.field}
            {#if fileManager.sortDir.current === "ASC"}
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
        onclick={() => fileManager.setViewMode("list")}
        class="p-1.5 transition-colors {fileManager.viewMode.current === 'list'
          ? 'bg-primary-soft text-primary'
          : 'bg-white text-ink-4 hover:bg-surface-muted hover:text-ink-3'}"
      >
        <LayoutList size={15} />
      </button>
      <button
        type="button"
        onclick={() => fileManager.setViewMode("grid")}
        class="p-1.5 transition-colors {fileManager.viewMode.current === 'grid'
          ? 'bg-primary-soft text-primary'
          : 'bg-white text-ink-4 hover:bg-surface-muted hover:text-ink-3'}"
      >
        <LayoutGrid size={15} />
      </button>
    </div>
  </div>
</div>
