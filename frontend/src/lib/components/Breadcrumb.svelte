<script lang="ts">
  import { ChevronRight, ChevronsLeft } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";

  export type Crumb = {
    id: string;
    name: string;
  };

  let {
    items,
    showHome = true,
    onNavigate,
    onHome,
  }: {
    items: Crumb[];
    showHome?: boolean;
    onNavigate: (id: string) => void;
    onHome?: () => void;
  } = $props();

  let expanded = $state(false);
  let containerWidth = $state(0);

  const HOME_WIDTH = 100;
  const ITEM_WIDTH = 100;
  const ELLIPSIS_WIDTH = 40;
  const CHEVRON_WIDTH = 20;

  // How many items can fit in collapsed mode
  let collapsedCount = $derived.by(() => {
    if (items.length <= 2) return items.length;
    const available = containerWidth - HOME_WIDTH - CHEVRON_WIDTH;
    if (items.length <= 3) return items.length;
    const maxMiddle = Math.max(
      0,
      Math.floor(
        (available - 2 * ITEM_WIDTH - ELLIPSIS_WIDTH) /
          (ITEM_WIDTH + CHEVRON_WIDTH),
      ),
    );
    return Math.max(3, 2 + Math.min(maxMiddle, items.length - 2));
  });

  let canCollapse = $derived(items.length > collapsedCount);
</script>

{#if items.length > 0 || showHome}
  <div
    class="flex items-center gap-1.5 overflow-hidden text-base"
    bind:clientWidth={containerWidth}
  >
    {#if showHome}
      <button
        type="button"
        onclick={() => {
          expanded = false;
          onHome?.();
        }}
        class="shrink-0 rounded px-1.5 py-1 text-base font-medium text-ink-3 transition-colors hover:text-ink"
      >
        {m.all_files()}
      </button>
    {/if}

    {#if expanded}
      <!-- Show all items -->
      {#each items as crumb, i}
        <ChevronRight size={14} class="shrink-0 text-ink-4" />
        {#if i === items.length - 1}
          <span
            class="max-w-48 truncate font-medium text-ink sm:max-w-64 md:max-w-80"
            title={crumb.name}>{crumb.name}</span
          >
        {:else}
          <button
            type="button"
            onclick={() => {
              expanded = false;
              onNavigate(crumb.id);
            }}
            class="max-w-32 shrink truncate rounded px-1 text-ink-3 transition-colors hover:text-ink sm:max-w-40"
            title={crumb.name}>{crumb.name}</button
          >
        {/if}
      {/each}
      {#if canCollapse}
        <button
          type="button"
          onclick={() => (expanded = false)}
          class="shrink-0 rounded p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
          title="Collapse"
        >
          <ChevronsLeft size={14} />
        </button>
      {/if}
    {:else}
      <!-- Collapsed mode -->
      {#each items as crumb, i}
        {#if items.length <= collapsedCount}
          <!-- All items fit -->
          {#if i > 0 || showHome}
            <ChevronRight size={14} class="shrink-0 text-ink-4" />
          {/if}
          {#if i === items.length - 1}
            <span
              class="max-w-48 truncate font-medium text-ink sm:max-w-64 md:max-w-80"
              title={crumb.name}>{crumb.name}</span
            >
          {:else}
            <button
              type="button"
              onclick={() => onNavigate(crumb.id)}
              class="max-w-32 shrink truncate rounded px-1 text-ink-3 transition-colors hover:text-ink sm:max-w-40"
              title={crumb.name}>{crumb.name}</button
            >
          {/if}
        {:else if i === 0}
          <!-- First item -->
          <ChevronRight size={14} class="shrink-0 text-ink-4" />
          <button
            type="button"
            onclick={() => onNavigate(crumb.id)}
            class="max-w-32 shrink truncate rounded px-1 text-ink-3 transition-colors hover:text-ink sm:max-w-40"
            title={crumb.name}>{crumb.name}</button
          >
        {:else if i === 1}
          <!-- Ellipsis -->
          <ChevronRight size={14} class="shrink-0 text-ink-4" />
          <button
            type="button"
            onclick={() => (expanded = true)}
            class="shrink-0 rounded px-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3"
            title={m.show_full_path()}>...</button
          >
          <!-- Last item -->
          <ChevronRight size={14} class="shrink-0 text-ink-4" />
          <span
            class="max-w-48 truncate font-medium text-ink sm:max-w-64 md:max-w-80"
            title={items[items.length - 1].name}
            >{items[items.length - 1].name}</span
          >
        {/if}
      {/each}
    {/if}
  </div>
{/if}
