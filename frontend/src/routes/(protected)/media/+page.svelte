<script lang="ts">
  import { onMount, onDestroy, getContext } from "svelte";
  import { user, authReady } from "$lib/stores/auth";
  import { ApiError, getAccessToken } from "$lib/api/client";
  import {
    addToLibrary,
    ensureMediaUploadDir,
    listMedia,
    removeFromLibrary,
    batchRemoveFromLibrary,
    renameMediaItem,
    getMediaItem,
    readdExistingUploadToLibrary,
    type AddToLibraryResponse,
    type MediaItem,
  } from "$lib/api/media";
  import {
    Trash2,
    LoaderCircle,
    Play,
    CircleAlert,
    Clock,
    Plus,
    Upload,
    ChevronDown,
    Check,
    X,
    Pencil,
  } from "@lucide/svelte";
  import { fly } from "svelte/transition";
  import { toast } from "svelte-sonner";
  import { confirmAction, confirmDelete, promptInput } from "$lib/dialog";
  import { fmtDurationText, authedUrl } from "$lib/utils/format";
  import AddMediaDialog from "$lib/components/media/AddMediaDialog.svelte";
  import PasteUploadProvider from "$lib/components/files/PasteUploadProvider.svelte";
  import { Popover } from "$lib/ui/popover";
  import type { createUploadManager as UploadMgrFn } from "$lib/upload-manager.svelte";
  type UploadManager = ReturnType<typeof UploadMgrFn>;
  import * as m from "$lib/paraglide/messages";
  import noFilesSvg from "$lib/assets/empty-states/no-files.svg";

  let items = $state<MediaItem[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let showAddDialog = $state(false);
  let showFabMenu = $state(false);
  let fabTimer: ReturnType<typeof setTimeout> | undefined;
  let videoInput: HTMLInputElement | undefined = $state();
  let es: EventSource | undefined;
  let refreshTimer: ReturnType<typeof setTimeout> | undefined;
  let selected = $state<Record<string, boolean>>({});
  let allSelected = $derived(
    items.length > 0 && items.every((item) => !!selected[item.mediaSlug]),
  );
  let hasSelection = $derived(Object.keys(selected).length > 0);
  let selectedItems = $derived(
    items.filter((item) => !!selected[item.mediaSlug]),
  );
  const ERR_CODE_NAME_CONFLICT = 2004;

  function isVideoFile(file: File) {
    if (file.type.startsWith("video/")) return true;
    return /\.(mp4|mov|webm|mkv|avi|flv|wmv|ogv|ogg|mpeg|mpg|m4v)$/i.test(
      file.name,
    );
  }

  function isNameConflict(error: unknown) {
    return (
      error instanceof ApiError && error.errCode === ERR_CODE_NAME_CONFLICT
    );
  }

  function notifyMediaAdd(resp: AddToLibraryResponse) {
    if (resp.alreadyInLibrary) {
      toast.info(m.media_already_in_library());
    } else if (resp.transcodeReused) {
      toast.success(m.media_add_success());
    } else {
      toast.success(m.media_transcode_started());
    }
  }

  const upload = getContext<UploadManager>("upload");

  $effect(() => {
    upload.setAcceptFile(isVideoFile);
    upload.setGetCurrentSlug(async () => {
      const dir = await ensureMediaUploadDir();
      return dir.slug;
    });
    upload.setOnRejected((files) => {
      toast.error(m.media_upload_rejected({ count: files.length }));
    });
    upload.setOnDuplicateDetected(() => true);
    upload.setOnFileImported(async ({ fileSlug }) => {
      try {
        const resp = await addToLibrary(fileSlug);
        notifyMediaAdd(resp);
        scheduleRefresh();
      } catch (e) {
        toast.error(e instanceof Error ? e.message : m.media_add_failed());
        throw e;
      }
    });
    upload.setOnImportConflict(
      async ({ physicalFileSlug, fileName, source, error }) => {
        if (source !== "dedup" || !isNameConflict(error)) return false;

        const confirmed = await confirmAction(
          m.media_existing_file_title(),
          m.media_existing_file_message({ name: fileName }),
          m.media_readd_existing_btn(),
        );
        if (!confirmed) {
          throw new Error(m.upload_skipped_duplicate());
        }

        try {
          const resp = await readdExistingUploadToLibrary(
            physicalFileSlug,
            fileName,
          );
          notifyMediaAdd(resp);
          scheduleRefresh();
          return true;
        } catch (e) {
          toast.error(e instanceof Error ? e.message : m.media_add_failed());
          throw e;
        }
      },
    );
    upload.setOnCompleted(async () => {
      await refresh(false);
    });
  });

  function connectSSE() {
    if (es) return;
    const token = getAccessToken();
    if (!token) return;
    const url = new URL("/api/v1/media/events", window.location.origin);
    url.searchParams.set("access_token", token);
    es = new EventSource(url.toString());

    function updateItem(mediaSlug: string, update: Partial<MediaItem>) {
      const idx = items.findIndex((i) => i.mediaSlug === mediaSlug);
      if (idx !== -1) {
        items[idx] = { ...items[idx], ...update };
      }
    }

    es.addEventListener("processing", (e) => {
      const data = JSON.parse(e.data);
      updateItem(data.mediaSlug, {
        status: data.status,
        progress: data.progress,
      });
    });

    es.addEventListener("done", (e) => {
      const data = JSON.parse(e.data);
      updateItem(data.mediaSlug, { status: data.status, progress: 100 });
      void refreshItem(data.mediaSlug);
    });

    es.addEventListener("failed", (e) => {
      const data = JSON.parse(e.data);
      updateItem(data.mediaSlug, {
        status: data.status,
        progress: 0,
        errorMsg: data.errorMsg ?? null,
      });
    });

    es.onerror = () => {
      // EventSource auto-reconnects
    };
  }

  function disconnectSSE() {
    if (es) {
      es.close();
      es = undefined;
    }
  }

  async function refreshItem(mediaSlug: string) {
    try {
      const updated = await getMediaItem(mediaSlug);
      const idx = items.findIndex((i) => i.mediaSlug === mediaSlug);
      if (idx !== -1) {
        items[idx] = updated;
      }
    } catch {
      // ignore
    }
  }

  function scheduleRefresh() {
    clearTimeout(refreshTimer);
    refreshTimer = setTimeout(() => {
      disconnectSSE();
      void refresh(false);
    }, 250);
  }

  async function refresh(showLoading = true) {
    if (!$user) return;
    if (showLoading) loading = true;
    try {
      const data = await listMedia();
      items = data.items;
      total = data.total;
      const itemSlugs = new Set(items.map((item) => item.mediaSlug));
      const next: Record<string, boolean> = {};
      for (const slug of Object.keys(selected)) {
        if (itemSlugs.has(slug)) next[slug] = true;
      }
      selected = next;
      connectSSE();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.media_load_failed());
    } finally {
      if (showLoading) loading = false;
    }
  }

  onDestroy(() => {
    disconnectSSE();
    clearTimeout(refreshTimer);
    upload.setAcceptFile();
    upload.setOnRejected();
    upload.setOnFileImported();
    upload.setOnImportConflict();
    upload.setOnDuplicateDetected();
  });

  function toggleSelect(mediaSlug: string) {
    if (selected[mediaSlug]) {
      const { [mediaSlug]: _, ...rest } = selected;
      selected = rest;
    } else {
      selected = { ...selected, [mediaSlug]: true };
    }
  }

  function toggleSelectAll() {
    if (allSelected) {
      selected = {};
    } else {
      selected = Object.fromEntries(items.map((item) => [item.mediaSlug, true]));
    }
  }

  function clearSelection() {
    selected = {};
  }

  async function remove(slug: string, name: string) {
    if (!(await confirmDelete(m.confirm_remove_media({ name })))) return;
    try {
      await removeFromLibrary(slug);
      items = items.filter((i) => i.mediaSlug !== slug);
      total--;
      const { [slug]: _, ...rest } = selected;
      selected = rest;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.media_remove_failed());
    }
  }

  async function batchRemove() {
    const targets = selectedItems;
    if (targets.length === 0) return;
    const names = targets.map((item) => item.fileName);
    if (
      !(await confirmDelete(
        m.confirm_delete_multiple({
          count: String(targets.length),
          names: names.join("\n"),
        }),
      ))
    )
      return;
    try {
      const slugs = targets.map((item) => item.mediaSlug);
      await batchRemoveFromLibrary(slugs);
      items = items.filter((item) => !selected[item.mediaSlug]);
      total = Math.max(0, total - targets.length);
      clearSelection();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.media_remove_failed());
    }
  }

  async function rename(slug: string, currentName: string) {
    const newName = await promptInput(
      m.rename(),
      m.enter_new_name(),
      currentName,
      512,
    );
    const trimmed = newName?.trim();
    if (!trimmed || trimmed === currentName) return;
    try {
      const updated = await renameMediaItem(slug, trimmed);
      const idx = items.findIndex((item) => item.mediaSlug === slug);
      if (idx !== -1) items[idx] = updated;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.rename_failed());
    }
  }

  function onFabEnter() {
    clearTimeout(fabTimer);
    showFabMenu = true;
  }

  function onFabLeave() {
    fabTimer = setTimeout(() => {
      showFabMenu = false;
    }, 150);
  }

  onMount(() => {
    void refresh();
  });
</script>

{#if $authReady && $user}
  <div class="space-y-4 relative px-6 pt-4 pb-6">
    <!-- Toolbar -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        {#if items.length > 0}
          <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
          <span
            role="checkbox"
            aria-checked={allSelected}
            tabindex="-1"
            class="flex h-4 w-4 shrink-0 cursor-pointer items-center justify-center rounded-full border transition-colors {allSelected
              ? 'border-primary bg-primary text-white'
              : 'border-line hover:border-primary'}"
            onclick={toggleSelectAll}
            onkeydown={(e) => {
              if (e.key === " " || e.key === "Enter") toggleSelectAll();
            }}
          >
            {#if allSelected}
              <Check size={10} />
            {/if}
          </span>
        {/if}
        <span class="text-xs text-ink-4">{m.total_items({ total })}</span>
      </div>
    </div>

    <input
      bind:this={videoInput}
      type="file"
      accept="video/*,.mkv,.avi,.flv,.wmv,.ogv,.ogg,.mpeg,.mpg,.m4v"
      multiple
      class="hidden"
      onchange={upload.onPick}
    />

    <PasteUploadProvider
      targetLabel={m.media_title()}
      acceptFile={isVideoFile}
      onUpload={(files) => upload.enqueueFiles(files)}
    />

    {#if loading}
      <div class="flex items-center justify-center py-24">
        <LoaderCircle size={24} class="animate-spin text-ink-4" />
      </div>
    {:else if items.length === 0}
      <div
        class="flex flex-col items-center justify-center py-24 text-center"
      >
        <img src={noFilesSvg} class="mb-2 w-32 h-32" alt="" />
        <p class="text-sm text-ink-4">{m.media_empty()}</p>
        <p class="mt-1 text-xs text-ink-4">{m.media_help()}</p>
      </div>
    {:else}
      <div class="grid gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {#each items as item (item.mediaSlug)}
          {@const isSelected = !!selected[item.mediaSlug]}
          <div class="group relative">
            <!-- Checkbox -->
            {#if !isSelected}
              <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
              <span
                role="checkbox"
                aria-checked={isSelected}
                tabindex="-1"
                class="absolute left-2 top-2 z-10 flex h-5 w-5 cursor-pointer items-center justify-center rounded-full border border-white/80 bg-black/30 text-white opacity-0 backdrop-blur transition-opacity hover:bg-black/50 group-hover:opacity-100"
                onclick={(e) => {
                  e.stopPropagation();
                  toggleSelect(item.mediaSlug);
                }}
                onkeydown={(e) => {
                  if (e.key === " " || e.key === "Enter") {
                    e.stopPropagation();
                    toggleSelect(item.mediaSlug);
                  }
                }}
              ></span>
            {:else}
              <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
              <span
                role="checkbox"
                aria-checked={isSelected}
                tabindex="-1"
                class="absolute left-2 top-2 z-10 flex h-5 w-5 cursor-pointer items-center justify-center rounded-full border border-primary bg-primary text-white opacity-100"
                onclick={(e) => {
                  e.stopPropagation();
                  toggleSelect(item.mediaSlug);
                }}
                onkeydown={(e) => {
                  if (e.key === " " || e.key === "Enter") {
                    e.stopPropagation();
                    toggleSelect(item.mediaSlug);
                  }
                }}
              >
                <Check size={10} strokeWidth={3} />
              </span>
            {/if}

            <!-- Thumbnail -->
            <div class="relative aspect-video overflow-hidden rounded-xl bg-surface-sunken {item.status === 'done' ? '' : 'cursor-default'}">
              {#if item.status === "done"}
                <a href="/media/{item.mediaSlug}" class="block h-full">
                  {#if item.posterUrl}
                    <img
                      src={authedUrl(item.posterUrl)}
                      alt={item.fileName}
                      class="h-full w-full object-cover transition group-hover:scale-105"
                      loading="lazy"
                    />
                    <div
                      class="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/0 transition group-hover:bg-black/30"
                    >
                      <Play
                        size={40}
                        class="text-white opacity-0 transition group-hover:opacity-100"
                        fill="currentColor"
                      />
                    </div>
                  {:else}
                    <div class="flex h-full items-center justify-center">
                      <div
                        class="flex h-12 w-12 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur transition-transform group-hover:scale-110"
                      >
                        <Play size={20} fill="currentColor" />
                      </div>
                    </div>
                  {/if}
                </a>
              {:else}
                <div
                  class="flex h-full flex-col items-center justify-center gap-2"
                >
                  {#if item.status === "processing"}
                    <LoaderCircle
                      size={24}
                      class="animate-spin text-primary"
                    />
                    <span class="text-xs text-primary">{item.progress}%</span>
                  {:else if item.status === "pending"}
                    <Clock size={24} class="text-ink-4" />
                    <span class="text-xs text-ink-4">{m.queued()}</span>
                  {:else if item.status === "failed"}
                    <CircleAlert size={24} class="text-danger" />
                    <span class="text-xs text-danger">{m.failed()}</span>
                  {/if}
                </div>
              {/if}

              <!-- Duration badge -->
              {#if item.durationSec}
                <div
                  class="absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white"
                >
                  {fmtDurationText(item.durationSec)}
                </div>
              {/if}
            </div>

            <!-- Info -->
            <div class="mt-2 flex items-start gap-2">
              <div class="min-w-0 flex-1">
                <p
                  class="truncate text-sm font-medium text-ink"
                  title={item.fileName}
                >
                  {item.fileName}
                </p>
                <p class="mt-0.5 text-xs text-ink-4">
                  {new Date(item.createdAt).toLocaleDateString()}
                </p>
              </div>
              <div
                class="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100 {isSelected
                  ? '!opacity-100'
                  : ''}"
              >
                <button
                  type="button"
                  onclick={() => rename(item.mediaSlug, item.fileName)}
                  class="rounded-md p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-primary"
                  title={m.rename()}
                >
                  <Pencil size={14} />
                </button>
                <button
                  type="button"
                  onclick={() => remove(item.mediaSlug, item.fileName)}
                  class="rounded-md p-1 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-danger"
                  title={m.remove()}
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
            {#if item.status === "failed" && item.errorMsg}
              <p
                class="mt-1 truncate text-xs text-danger"
                title={item.errorMsg}
              >
                {item.errorMsg}
              </p>
            {/if}
          </div>
        {/each}
      </div>
    {/if}

    {#if hasSelection}
      <div
        class="fixed bottom-6 left-1/2 z-50 max-w-[calc(100vw-1rem)] -translate-x-1/2"
      >
        <div
          class="flex items-center gap-2 overflow-x-auto rounded-full border border-line-soft bg-white/95 px-3 py-2 shadow-[0_12px_36px_rgba(15,23,42,0.16)] backdrop-blur"
          transition:fly={{ y: 16, duration: 180, opacity: 0 }}
        >
          <span class="shrink-0 px-3 text-sm font-medium text-ink-2"
            >{m.selected_count({
              count: String(Object.keys(selected).length),
            })}</span
          >
          <div class="h-7 w-px shrink-0 bg-surface-sunken"></div>
          <button
            type="button"
            onclick={batchRemove}
            class="flex h-8 w-8 items-center justify-center rounded-full text-ink-3 transition-colors hover:bg-danger-soft hover:text-danger"
            title={m.delete_label()}
          >
            <Trash2 size={16} />
          </button>
          <div class="mx-1 h-7 w-px bg-surface-sunken"></div>
          <button
            type="button"
            onclick={clearSelection}
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
            title={m.close()}
          >
            <X size={16} />
          </button>
        </div>
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

        <div
          role="region"
          onmouseenter={onFabEnter}
          onmouseleave={onFabLeave}
        >
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-primary-soft hover:text-primary"
            onclick={() => {
              showFabMenu = false;
              videoInput?.click();
            }}
          >
            <Upload size={15} class="text-primary" />
            {m.upload_video()}
          </button>
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-ink-2 outline-none transition-colors duration-150 select-none cursor-pointer hover:bg-primary-soft hover:text-primary"
            onclick={() => {
              showFabMenu = false;
              showAddDialog = true;
            }}
          >
            <Plus size={15} class="text-primary" />
            {m.add_to_media_library()}
          </button>
        </div>
      </Popover>
    </div>

    <AddMediaDialog
      open={showAddDialog}
      onClose={() => (showAddDialog = false)}
      onDone={refresh}
    />
  </div>
{/if}
