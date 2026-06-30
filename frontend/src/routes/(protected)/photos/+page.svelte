<script lang="ts">
  import { onMount } from "svelte";
  import { user, authReady } from "$lib/stores/auth";
  import { authedUrl } from "$lib/utils/format";
  import * as m from "$lib/paraglide/messages";
  import noFilesSvg from "$lib/assets/empty-states/no-files.svg";
  import { toast } from "svelte-sonner";
  import { LoaderCircle, ChevronDown } from "@lucide/svelte";
  import { Dropdown, DropdownBase } from "$lib/ui/dropdown";

  import {
    listPhotos,
    thumbnailUrl,
    type PhotoItem,
  } from "$lib/api/photos";
  import PhotoViewer from "$lib/components/PhotoViewer.svelte";

  const PAGE_SIZE = 50;

  let photos = $state<PhotoItem[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let loadingMore = $state(false);
  let page = $state(1);

  let photoSize = $state<"small" | "medium" | "large">("medium");
  let groupByDate = $state(true);
  let sizeLabel = $derived(
    photoSize === "large"
      ? m.photos_size_large()
      : photoSize === "medium"
        ? m.photos_size_medium()
        : m.photos_size_small(),
  );

  // Lightbox
  let viewerSlug = $state<string | null>(null);
  let viewerOpen = $state(false);
  let viewerIndex = $state(0);
  let allSlugs = $derived(photos.map((p) => p.slug));

  let hasMore = $derived(photos.length < total);
  let gridClass = $derived(
    photoSize === "large"
      ? "grid-cols-2 gap-1.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 min-[1920px]:grid-cols-8 min-[2560px]:grid-cols-10 min-[3840px]:grid-cols-14"
      : photoSize === "small"
        ? "grid-cols-4 gap-1 sm:grid-cols-7 md:grid-cols-9 lg:grid-cols-10 xl:grid-cols-11 2xl:grid-cols-13 min-[1920px]:grid-cols-15 min-[2560px]:grid-cols-19 min-[3840px]:grid-cols-25"
        : "grid-cols-3 gap-1 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-9 2xl:grid-cols-10 min-[1920px]:grid-cols-11 min-[2560px]:grid-cols-13 min-[3840px]:grid-cols-17",
  );

  $effect(() => {
    const savedSize = localStorage.getItem("nd.photos.size");
    if (
      savedSize === "small" ||
      savedSize === "medium" ||
      savedSize === "large"
    )
      photoSize = savedSize;
    const savedGroup = localStorage.getItem("nd.photos.group");
    if (savedGroup !== null) groupByDate = savedGroup === "true";
  });
  $effect(() => {
    localStorage.setItem("nd.photos.size", photoSize);
  });
  $effect(() => {
    localStorage.setItem("nd.photos.group", String(groupByDate));
  });

  function isSameDay(a: Date, b: Date): boolean {
    return (
      a.getFullYear() === b.getFullYear() &&
      a.getMonth() === b.getMonth() &&
      a.getDate() === b.getDate()
    );
  }

  function getDayLabel(isoDate: string): string {
    const d = new Date(isoDate);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    if (isSameDay(d, today)) return m.photos_today();
    if (isSameDay(d, yesterday)) return m.photos_yesterday();
    return d.toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  }

  type DayGroup = { date: string; label: string; items: PhotoItem[] };

  let groupedPhotos = $derived.by(() => {
    const groups = new Map<string, PhotoItem[]>();
    for (const f of photos) {
      const day = f.createdAt.slice(0, 10);
      if (!groups.has(day)) groups.set(day, []);
      groups.get(day)!.push(f);
    }
    return Array.from(groups.entries())
      .sort(([a], [b]) => b.localeCompare(a))
      .map(
        ([date, items]): DayGroup => ({
          date,
          label: getDayLabel(date),
          items,
        }),
      );
  });

  async function fetchPhotos() {
    loading = true;
    page = 1;
    try {
      const data = await listPhotos(1, PAGE_SIZE);
      photos = data.items;
      total = data.total;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.load_failed());
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (loadingMore || !hasMore) return;
    loadingMore = true;
    const nextPage = page + 1;
    try {
      const data = await listPhotos(nextPage, PAGE_SIZE);
      photos = [...photos, ...data.items];
      page = nextPage;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.load_failed());
    } finally {
      loadingMore = false;
    }
  }

  function openViewer(slug: string) {
    viewerIndex = allSlugs.indexOf(slug);
    viewerSlug = slug;
    viewerOpen = true;
  }

  onMount(() => {
    void fetchPhotos();
  });
</script>

{#if $authReady && $user}
  <div class="space-y-4 px-6 pt-4 pb-6">
    <!-- Toolbar -->
    <div class="flex items-center justify-between">
      <span class="text-xs text-ink-4">{m.photos_total({ total })}</span>
      <div class="flex items-center gap-2">
        <Dropdown
          triggerClass="flex h-8 items-center gap-1.5 rounded-lg border border-line bg-surface px-2.5 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
          contentClass="min-w-[180px]"
        >
          {#snippet trigger()}
            <span class="hidden sm:inline">{sizeLabel}</span>
            <ChevronDown size={14} class="text-ink-4" />
          {/snippet}
          <DropdownBase.Item onSelect={() => (groupByDate = !groupByDate)}>
            {#snippet children()}
              {groupByDate ? "✓ " : ""}{m.photos_group_by_date()}
            {/snippet}
          </DropdownBase.Item>
          <DropdownBase.Separator />
          <DropdownBase.Item onSelect={() => (photoSize = "large")}>
            {#snippet children()}
              {photoSize === "large" ? "✓ " : ""}{m.photos_size_large()}
            {/snippet}
          </DropdownBase.Item>
          <DropdownBase.Item onSelect={() => (photoSize = "medium")}>
            {#snippet children()}
              {photoSize === "medium" ? "✓ " : ""}{m.photos_size_medium()}
            {/snippet}
          </DropdownBase.Item>
          <DropdownBase.Item onSelect={() => (photoSize = "small")}>
            {#snippet children()}
              {photoSize === "small" ? "✓ " : ""}{m.photos_size_small()}
            {/snippet}
          </DropdownBase.Item>
        </Dropdown>
      </div>
    </div>

    <!-- Loading -->
    {#if loading}
      <div class="flex items-center justify-center py-24">
        <LoaderCircle size={24} class="animate-spin text-ink-4" />
      </div>
    {:else if photos.length === 0}
      <div
        class="flex flex-col items-center justify-center py-24 text-center"
      >
        <img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
        <p class="text-sm text-ink-4">{m.photos_empty()}</p>
      </div>
    {:else}
      {#if groupByDate}
        {#each groupedPhotos as group (group.date)}
          <div class="mb-6">
            <div class="sticky top-0 z-30 bg-surface-muted/85 backdrop-blur px-1 py-2">
              <h2 class="text-sm font-semibold text-ink">{group.label}</h2>
            </div>
            <div class="grid {gridClass}">
              {#each group.items as photo (photo.slug)}
                <button
                  type="button"
                  onclick={() => openViewer(photo.slug)}
                  class="group relative aspect-square overflow-hidden bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                >
                  <img
                    src={authedUrl(thumbnailUrl(photo.slug))}
                    alt={photo.fileName}
                    class="h-full w-full object-cover transition group-hover:scale-105"
                    loading="lazy"
                  />
                  <div
                    class="pointer-events-none absolute inset-0 bg-black/0 transition group-hover:bg-black/15"
                  ></div>
                  {#if photo.isStarred}
                    <div
                      class="pointer-events-none absolute right-1 top-1 rounded-full bg-surface/80 p-0.5"
                    >
                      <svg
                        viewBox="0 0 24 24"
                        class="h-2.5 w-2.5 fill-star text-warning"
                        ><path
                          d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                        /></svg
                      >
                    </div>
                  {/if}
                </button>
              {/each}
            </div>
          </div>
        {/each}
      {:else}
        <div class="grid {gridClass}">
          {#each photos as photo (photo.slug)}
            <button
              type="button"
              onclick={() => openViewer(photo.slug)}
              class="group relative aspect-square overflow-hidden bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
            >
              <img
                src={authedUrl(thumbnailUrl(photo.slug))}
                alt={photo.fileName}
                class="h-full w-full object-cover transition group-hover:scale-105"
                loading="lazy"
              />
              <div
                class="pointer-events-none absolute inset-0 bg-black/0 transition group-hover:bg-black/15"
              ></div>
              {#if photo.isStarred}
                <div
                  class="pointer-events-none absolute right-1 top-1 rounded-full bg-surface/80 p-0.5"
                >
                  <svg
                    viewBox="0 0 24 24"
                    class="h-2.5 w-2.5 fill-star text-warning"
                    ><path
                      d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                    /></svg
                  >
                </div>
              {/if}
            </button>
          {/each}
        </div>
      {/if}

      {#if hasMore}
        <div class="flex justify-center py-4">
          <button
            type="button"
            onclick={loadMore}
            disabled={loadingMore}
            class="text-xs text-ink-3 transition-colors hover:text-ink-2 disabled:opacity-50"
          >
            {#if loadingMore}
              <LoaderCircle size={12} class="mr-1 inline animate-spin" />
            {/if}
            {m.photos_load_more()}
          </button>
        </div>
      {:else if photos.length > 0}
        <div class="flex justify-center py-4">
          <span class="text-xs text-ink-4">{m.no_more()}</span>
        </div>
      {/if}
    {/if}
  </div>

  <!-- Lightbox -->
  <PhotoViewer
    bind:open={viewerOpen}
    slug={viewerSlug}
    bind:fileSlugs={allSlugs}
    index={viewerIndex}
    onClose={() => (viewerSlug = null)}
    {photos}
  />
{/if}
