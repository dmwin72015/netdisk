<script lang="ts">
  import { fmtSize } from "$lib/utils/format";
  import { LoaderCircle, Check } from "@lucide/svelte";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { Dialog } from "$lib/ui/dialog";
  import * as m from "$lib/paraglide/messages";

  let {
    files,
    open = $bindable(false),
    loading = false,
    onConfirm,
    onCancel,
  }: {
    files: { file: File; relativePath: string }[];
    open?: boolean;
    loading?: boolean;
    onConfirm: (selected: { file: File; relativePath: string }[]) => void;
    onCancel: () => void;
  } = $props();

  let selected = $state<Set<number>>(new Set());

  // Initialize all files as selected when files change
  $effect(() => {
    if (files.length > 0) {
      selected = new Set(files.map((_, i) => i));
    }
  });

  let selectedCount = $derived(selected.size);
  let selectedSize = $derived(
    files.reduce((sum, f, i) => (selected.has(i) ? sum + f.file.size : sum), 0),
  );

  function toggle(index: number) {
    const next = new Set(selected);
    if (next.has(index)) {
      next.delete(index);
    } else {
      next.add(index);
    }
    selected = next;
  }

  function toggleAll() {
    if (selected.size === files.length) {
      selected = new Set();
    } else {
      selected = new Set(files.map((_, i) => i));
    }
  }

  function handleConfirm() {
    const picked = files.filter((_, i) => selected.has(i));
    onConfirm(picked);
  }

  function getCategoryFromName(name: string): string {
    const ext = name.split(".").pop()?.toLowerCase() || "";
    const map: Record<string, string> = {
      mp4: "video",
      mkv: "video",
      avi: "video",
      mov: "video",
      webm: "video",
      mp3: "audio",
      flac: "audio",
      wav: "audio",
      aac: "audio",
      ogg: "audio",
      jpg: "image",
      jpeg: "image",
      png: "image",
      gif: "image",
      webp: "image",
      svg: "image",
      pdf: "document",
      doc: "document",
      docx: "document",
      txt: "document",
      xls: "document",
      xlsx: "document",
      zip: "archive",
      rar: "archive",
      "7z": "archive",
      tar: "archive",
      gz: "archive",
    };
    return map[ext] || "other";
  }

  function getMimeTypeFromName(name: string): string | null {
    const ext = name.split(".").pop()?.toLowerCase() || "";
    const map: Record<string, string> = {
      mp4: "video/mp4",
      mkv: "video/x-matroska",
      avi: "video/x-msvideo",
      mp3: "audio/mpeg",
      flac: "audio/flac",
      wav: "audio/wav",
      jpg: "image/jpeg",
      png: "image/png",
      gif: "image/gif",
      webp: "image/webp",
      pdf: "application/pdf",
      txt: "text/plain",
    };
    return map[ext] || null;
  }

  function handleClose(v: boolean) {
    if (!v && !loading) onCancel();
  }
</script>

<Dialog
  bind:open
  onOpenChangeComplete={handleClose}
  title={loading ? m.loading() : m.select_files_to_upload()}
  footer={false}
  closable={!loading}
  size="md"
  bodyClass="p-0"
>
  {#if loading}
    <!-- Loading state -->
    <div class="flex flex-col items-center justify-center py-16">
      <LoaderCircle size={32} class="mb-4 animate-spin text-primary" />
      <p class="text-sm text-ink-3">{m.reading_files()}</p>
    </div>
  {:else}
    <!-- Stats bar -->
    <div
      class="flex items-center justify-between border-b border-line-soft px-5 py-2.5 text-xs text-ink-3"
    >
      <button
        type="button"
        onclick={toggleAll}
        class="flex items-center gap-1.5 text-sm text-ink-3 hover:text-ink"
      >
        <span
          class="flex h-4 w-4 items-center justify-center rounded border {selected.size ===
          files.length
            ? 'border-primary bg-primary text-primary-on'
            : 'border-line'}"
        >
          {#if selected.size === files.length}
            <Check size={12} />
          {/if}
        </span>
        {m.files_selected({ count: String(selectedCount) })}
      </button>
      <span>{m.total_size({ size: fmtSize(selectedSize) })}</span>
    </div>

    <!-- File list -->
    <div class="max-h-[50vh] overflow-y-auto px-2 py-2">
      {#each files as item, i (item.relativePath)}
        <button
          type="button"
          onclick={() => toggle(i)}
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition-colors hover:bg-surface-sunken"
        >
          <span
            class="flex h-4 w-4 shrink-0 items-center justify-center rounded border {selected.has(
              i,
            )
              ? 'border-primary bg-primary text-primary-on'
              : 'border-line'}"
          >
            {#if selected.has(i)}
              <Check size={12} />
            {/if}
          </span>
          <MimeIcon
            mimeType={getMimeTypeFromName(item.file.name)}
            name={item.file.name}
            isDir={false}
            category={getCategoryFromName(item.file.name)}
            size={18}
          />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm text-ink-2">{item.relativePath}</p>
          </div>
          <span class="shrink-0 text-xs text-ink-4"
            >{fmtSize(item.file.size)}</span
          >
        </button>
      {/each}
    </div>

    <!-- Footer -->
    <div
      class="flex items-center justify-end gap-2 border-t border-line-soft px-5 py-3"
    >
      <button
        type="button"
        onclick={onCancel}
        class="rounded-lg border border-line bg-surface px-4 py-2 text-sm text-ink-2 transition-colors hover:bg-surface-sunken"
      >
        {m.cancel()}
      </button>
      <button
        type="button"
        onclick={handleConfirm}
        disabled={selectedCount === 0}
        class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover disabled:opacity-50"
      >
        {m.upload_files()} ({selectedCount})
      </button>
    </div>
  {/if}
</Dialog>
