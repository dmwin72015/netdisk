<script lang="ts">
  import { Dialog } from "bits-ui";
  import { goto } from "$app/navigation";
  import { Search, X, LoaderCircle } from "@lucide/svelte";
  import { listFiles, type FileItem } from "$lib/api/files";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { fmtSize } from "$lib/utils/format";
  import * as m from "$lib/paraglide/messages";
  import noSearchResultsSvg from "$lib/assets/empty-states/no-search-results.svg";

  type CategoryDef = {
    key: string | null;
    label: string;
  };

  const categories: CategoryDef[] = [
    { key: null, label: m.all() },
    { key: "folder", label: m.directory() },
    { key: "image", label: m.category_image() },
    { key: "video", label: m.category_video() },
    { key: "audio", label: m.category_audio() },
    { key: "document", label: m.category_document() },
    { key: "archive", label: m.category_archive() },
    { key: "other", label: m.category_other() },
  ];

  let {
    open = $bindable(false),
  }: {
    open?: boolean;
  } = $props();

  let query = $state("");
  let results = $state<FileItem[]>([]);
  let total = $state(0);
  let loading = $state(false);
  let selectedCategory = $state<string | null>(null);
  let searchTimer: ReturnType<typeof setTimeout> | undefined;
  let searchRequestId = 0;
  let inputEl = $state<HTMLInputElement | undefined>(undefined);
  let highlightedIndex = $state(0);

  function resetSearch() {
    clearTimeout(searchTimer);
    searchRequestId += 1;
    query = "";
    results = [];
    total = 0;
    loading = false;
    selectedCategory = null;
    highlightedIndex = 0;
  }

  function closeDialog() {
    open = false;
    resetSearch();
  }

  function doSearch() {
    const term = query.trim();
    const category = selectedCategory;
    const requestId = ++searchRequestId;

    if (!term && !category) {
      results = [];
      total = 0;
      loading = false;
      return;
    }
    loading = true;
    highlightedIndex = 0;
    listFiles(
      undefined,
      1,
      20,
      undefined,
      category,
      "created_at",
      "DESC",
      false,
      false,
      term || undefined,
    )
      .then((data) => {
        if (
          requestId !== searchRequestId ||
          term !== query.trim() ||
          category !== selectedCategory
        )
          return;
        results = data.files;
        total = data.total;
      })
      .catch(() => {
        if (requestId !== searchRequestId) return;
        results = [];
        total = 0;
      })
      .finally(() => {
        if (requestId !== searchRequestId) return;
        loading = false;
      });
  }

  function onInput() {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(doSearch, 200);
  }

  function selectCategory(key: string | null) {
    selectedCategory = key;
    clearTimeout(searchTimer);
    doSearch();
  }

  function handleConfirm() {
    const item = results[highlightedIndex];
    if (item) {
      closeDialog();
      if (item.isDir) {
        void goto(`/files/all/${item.slug}`);
      } else {
        void goto(`/files/all/${item.parentSlug || ""}`);
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      closeDialog();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      highlightedIndex = Math.min(highlightedIndex + 1, results.length - 1);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      highlightedIndex = Math.max(highlightedIndex - 1, 0);
    } else if (e.key === "Enter") {
      e.preventDefault();
      handleConfirm();
    }
  }

  function clearQuery() {
    query = "";
    results = [];
    total = 0;
    loading = false;
    searchRequestId += 1;
    clearTimeout(searchTimer);
    inputEl?.focus();
  }

  $effect(() => {
    if (open) {
      setTimeout(() => inputEl?.focus(), 50);
    }
  });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<svelte:window onkeydown={handleKeydown} />

<Dialog.Root
  bind:open
  onOpenChangeComplete={(v) => {
    if (!v) closeDialog();
  }}
>
  <Dialog.Overlay
    class="fixed inset-0 z-50 bg-overlay backdrop-blur-sm
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			duration-200"
  />
  <Dialog.Content
    class="w-full max-w-2xl overflow-hidden rounded-xl border border-line bg-surface shadow-dialog fixed left-1/2 top-[15vh] z-50 -translate-x-1/2
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			duration-200"
  >
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
      class="w-full overflow-hidden rounded-xl border border-line bg-surface shadow-dialog"
      onclick={(e) => e.stopPropagation()}
      onkeydown={() => {}}
    >
      <!-- Search input -->
      <div class="flex items-center gap-3 border-b border-line-soft px-4 py-3">
        <Search size={18} class="shrink-0 text-ink-4" />
        <input
          bind:this={inputEl}
          type="text"
          bind:value={query}
          oninput={onInput}
          placeholder="{m.search_files()}..."
          class="min-w-0 flex-1 text-base text-ink-2 outline-none placeholder:text-ink-4"
        />
        {#if query}
          <button
            type="button"
            onclick={clearQuery}
            class="rounded-lg p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
          >
            <X size={16} />
          </button>
        {/if}
      </div>

      <!-- Category chips -->
      <div class="flex flex-wrap gap-1.5 border-b border-line-soft px-4 py-2.5">
        {#each categories as cat (cat.key ?? "all")}
          <button
            type="button"
            onclick={() => selectCategory(cat.key)}
            class="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors {selectedCategory ===
            cat.key
              ? 'bg-primary text-primary-on'
              : 'bg-surface-sunken text-ink-3 hover:bg-line hover:text-ink-2'}"
          >
            {cat.label}
          </button>
        {/each}
      </div>

      <!-- Results -->
      <div class="max-h-[50vh] overflow-y-auto">
        {#if loading}
          <div class="flex items-center justify-center py-12">
            <LoaderCircle size={20} class="animate-spin text-ink-4" />
          </div>
        {:else if results.length > 0}
          {#each results as item, i (item.slug)}
            {@const isHighlighted = i === highlightedIndex}
            <button
              type="button"
              onclick={() => {
                closeDialog();
                if (item.isDir) void goto(`/files/all/${item.slug}`);
                else void goto(`/files/all/${item.parentSlug || ""}`);
              }}
              onmouseenter={() => {
                highlightedIndex = i;
              }}
              class="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors {isHighlighted
                ? 'bg-primary-soft'
                : 'hover:bg-surface-sunken'}"
            >
              <MimeIcon
                mimeType={item.mimeType}
                name={item.fileName}
                isDir={item.isDir}
                category={item.fileCategory}
                size={18}
              />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm text-ink-2">{item.fileName}</p>
                <p class="truncate text-xs text-ink-4">
                  {#if item.parentName}
                    <span>{item.parentName} · </span>
                  {/if}
                  {fmtSize(item.fileSize)}
                </p>
              </div>
              <span
                class="shrink-0 rounded bg-surface-sunken px-1.5 py-0.5 text-[10px] font-medium uppercase text-ink-3"
                >{item.fileCategory || "—"}</span
              >
            </button>
          {/each}
          {#if results.length < total}
            <div
              class="border-t border-line-soft px-4 py-2 text-center text-xs text-ink-4"
            >
              {m.total_items({ total: String(total) })} · {m.search_files()}
              {m.load_more()}
            </div>
          {/if}
        {:else if query.trim() || selectedCategory}
          <div class="py-8 text-center">
            <img
              src={noSearchResultsSvg}
              class="mx-auto mb-2 w-28 h-28"
              alt=""
            />
            <p class="text-sm text-ink-4">{m.no_files()}</p>
          </div>
        {:else}
          <div class="py-12 text-center text-sm text-ink-4">
            {m.search_files()}
          </div>
        {/if}
      </div>

      <!-- Footer hint -->
      <div
        class="flex items-center gap-3 border-t border-line-soft px-4 py-2 text-[11px] text-ink-4"
      >
        <span
          ><kbd
            class="rounded border border-line px-1 py-0.5 text-[10px] text-ink-3"
            >↑↓</kbd
          >
          {m.sort_by()}</span
        >
        <span
          ><kbd
            class="rounded border border-line px-1 py-0.5 text-[10px] text-ink-3"
            >↵</kbd
          >
          {m.confirm()}</span
        >
        <span
          ><kbd
            class="rounded border border-line px-1 py-0.5 text-[10px] text-ink-3"
            >Esc</kbd
          >
          {m.close()}</span
        >
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
