<script lang="ts">
  import { goto } from '$app/navigation';
  import { Search, X, LoaderCircle } from '@lucide/svelte';
  import { listFiles, type FileItem } from '$lib/api/files';
  import MimeIcon from '$lib/components/MimeIcon.svelte';
  import { fmtSize } from '$lib/utils/format';
  import * as m from '$lib/paraglide/messages';

  type CategoryDef = {
    key: string | null;
    label: string;
  };

  const categories: CategoryDef[] = [
    { key: null, label: m.all() },
    { key: 'folder', label: m.directory() },
    { key: 'image', label: m.category_image() },
    { key: 'video', label: m.category_video() },
    { key: 'audio', label: m.category_audio() },
    { key: 'document', label: m.category_document() },
    { key: 'archive', label: m.category_archive() },
    { key: 'other', label: m.category_other() },
  ];

  let {
    open = $bindable(false),
  }: {
    open?: boolean;
  } = $props();

  let query = $state('');
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
    query = '';
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
      'created_at',
      'DESC',
      false,
      false,
      term || undefined
    )
      .then((data) => {
        if (requestId !== searchRequestId || term !== query.trim() || category !== selectedCategory) return;
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
        void goto(`/files/all/${item.parentSlug || ''}`);
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      closeDialog();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      highlightedIndex = Math.min(highlightedIndex + 1, results.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      highlightedIndex = Math.max(highlightedIndex - 1, 0);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      handleConfirm();
    }
  }

  function clearQuery() {
    query = '';
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

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-start justify-center bg-black/30 pt-[15vh] backdrop-blur-sm"
    onclick={closeDialog}
    onkeydown={(e) => { if (e.key === 'Escape') closeDialog(); }}
  >
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
      class="w-full max-w-2xl overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-2xl"
      onclick={(e) => e.stopPropagation()}
      onkeydown={() => {}}
    >
      <!-- Search input -->
      <div class="flex items-center gap-3 border-b border-gray-100 px-4 py-3">
        <Search size={18} class="shrink-0 text-gray-400" />
        <input
          bind:this={inputEl}
          type="text"
          bind:value={query}
          oninput={onInput}
          placeholder="{m.search_files()}..."
          class="min-w-0 flex-1 text-base text-gray-800 outline-none placeholder:text-gray-400"
        />
        {#if query}
          <button type="button" onclick={clearQuery} class="rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600">
            <X size={16} />
          </button>
        {/if}
      </div>

      <!-- Category chips -->
      <div class="flex flex-wrap gap-1.5 border-b border-gray-50 px-4 py-2.5">
        {#each categories as cat (cat.key ?? 'all')}
          <button
            type="button"
            onclick={() => selectCategory(cat.key)}
            class="rounded-lg px-2.5 py-1 text-xs font-medium transition-colors {selectedCategory === cat.key ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 hover:text-gray-800'}"
          >
            {cat.label}
          </button>
        {/each}
      </div>

      <!-- Results -->
      <div class="max-h-[50vh] overflow-y-auto">
        {#if loading}
          <div class="flex items-center justify-center py-12">
            <LoaderCircle size={20} class="animate-spin text-gray-300" />
          </div>
        {:else if results.length > 0}
          {#each results as item, i (item.slug)}
            {@const isHighlighted = i === highlightedIndex}
            <button
              type="button"
              onclick={() => {
                closeDialog();
                if (item.isDir) void goto(`/files/all/${item.slug}`);
                else void goto(`/files/all/${item.parentSlug || ''}`);
              }}
              onmouseenter={() => { highlightedIndex = i; }}
              class="flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors {isHighlighted ? 'bg-blue-50' : 'hover:bg-gray-50'}"
            >
              <MimeIcon mimeType={item.mimeType} name={item.fileName} isDir={item.isDir} category={item.fileCategory} size={18} />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm text-gray-800">{item.fileName}</p>
                <p class="truncate text-xs text-gray-400">
                  {#if item.parentName}
                    <span>{item.parentName} · </span>
                  {/if}
                  {fmtSize(item.fileSize)}
                </p>
              </div>
              <span class="shrink-0 rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium uppercase text-gray-500">{item.fileCategory || '—'}</span>
            </button>
          {/each}
          {#if results.length < total}
            <div class="border-t border-gray-50 px-4 py-2 text-center text-xs text-gray-400">
              {m.total_items({ total: String(total) })} · {m.search_files()} {m.load_more()}
            </div>
          {/if}
        {:else if query.trim() || selectedCategory}
          <div class="py-12 text-center text-sm text-gray-400">{m.no_files()}</div>
        {:else}
          <div class="py-12 text-center text-sm text-gray-400">{m.search_files()}</div>
        {/if}
      </div>

      <!-- Footer hint -->
      <div class="flex items-center gap-3 border-t border-gray-50 px-4 py-2 text-[11px] text-gray-400">
        <span><kbd class="rounded border border-gray-200 px-1 py-0.5 text-[10px] text-gray-500">↑↓</kbd> {m.sort_by()}</span>
        <span><kbd class="rounded border border-gray-200 px-1 py-0.5 text-[10px] text-gray-500">↵</kbd> {m.confirm()}</span>
        <span><kbd class="rounded border border-gray-200 px-1 py-0.5 text-[10px] text-gray-500">Esc</kbd> {m.close()}</span>
      </div>
    </div>
  </div>
{/if}
