<script lang="ts">
  import { Copy } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import { fmtSize, fmtTime, copyToClipboard } from "$lib/utils/format";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";
  import MimeIcon from "$lib/components/MimeIcon.svelte";
  import { Dialog } from "$lib/ui/dialog";
  import { isImageFile, isVideoFile, authedFileUrl } from "$lib/utils/file-helpers";

  let {
    open = $bindable(false),
    file,
    currentPath = [],
    downloadUrlFn,
    onOpenChangeComplete,
    summary = null,
    summaryLoading = false,
  }: {
    open?: boolean;
    file: NormalizedFile | null;
    currentPath?: { id: string; name: string }[];
    downloadUrlFn: (id: string) => string;
    onOpenChangeComplete?: (open: boolean) => void;
    summary?: { fileCount: number; folderCount: number; size: number } | null;
    summaryLoading?: boolean;
  } = $props();

  function describeFolder(): string {
    if (!file) return "";
    const size = fmtSize(summary?.size ?? file.size);
    if (summary) {
      return m.folder_detail_size({
        size,
        files: String(summary.fileCount),
        folders: String(summary.folderCount),
      });
    }
    return summaryLoading ? m.folder_detail_size_loading({ size }) : size;
  }

  function describeDetailFile(): string {
    if (!file) return "";
    return file.isDir ? describeFolder() : fmtSize(file.size);
  }

  function detailPathParts(): string[] {
    return [m.all_files(), ...currentPath.map((item) => item.name)];
  }

  async function copySlug(slug: string) {
    if (await copyToClipboard(slug)) {
      toast.success(m.copied());
    } else {
      toast.error(m.copy_failed());
    }
  }

  function categoryLabel(cat: string): string {
    const map: Record<string, string> = {
      video: m.category_video(),
      audio: m.category_audio(),
      image: m.category_image(),
      document: m.category_document(),
      archive: m.category_archive(),
      other: m.category_other(),
    };
    return map[cat] ?? cat;
  }
</script>

<Dialog
  bind:open
  {onOpenChangeComplete}
  title={file?.name}
  footer={false}
  class="max-w-130"
  bodyClass="px-5 py-5"
>
  {#if file}
    <div>
      <div class="mb-6 flex h-30 items-center justify-center">
        {#if file.isDir}
          <MimeIcon
            mimeType={file.mimeType}
            name={file.name}
            isDir={true}
            category={file.fileCategory}
            size={88}
          />
        {:else if isImageFile(file)}
          <img
            src={authedFileUrl(file, downloadUrlFn)}
            alt={file.name}
            loading="lazy"
            class="max-h-[112px] max-w-[220px] rounded-lg border border-line object-contain"
          />
        {:else if isVideoFile(file)}
          <video
            src={authedFileUrl(file, downloadUrlFn)}
            class="max-h-[112px] max-w-[220px] rounded-lg border border-line object-cover"
            muted
            preload="metadata"
          ></video>
        {:else}
          <MimeIcon
            mimeType={file.mimeType}
            name={file.name}
            isDir={file.isDir}
            category={file.fileCategory}
            size={88}
          />
        {/if}
      </div>

      <section>
        <h3 class="text-sm font-medium leading-5 text-ink-2">
          {m.detail_info()}
        </h3>
        <div class="mt-4 space-y-3">
          <div>
            <p class="text-xs leading-5 text-ink-3">{file.name}</p>
            <p class="mt-1 text-sm leading-5 text-ink-2">
              {describeDetailFile()}
            </p>
          </div>

          <div>
            <p class="text-xs leading-5 text-ink-3">{m.file_location()}</p>
            <p class="mt-1 text-sm leading-5 text-ink-2">
              {#each detailPathParts() as part, index}
                {#if index > 0}
                  <span class="px-1 text-ink-2">›</span>
                {/if}
                <span class="underline underline-offset-4">{part}</span>
              {/each}
            </p>
          </div>

          <div>
            <p class="text-xs leading-5 text-ink-3">{m.file_slug()}</p>
            <div class="mt-1 flex items-center gap-2">
              <p class="text-sm leading-5 text-ink-2 font-mono">
                {file.slug}
              </p>
              <button
                onclick={() => copySlug(file.slug)}
                class="rounded p-1 text-ink-4 hover:text-ink hover:bg-surface-sunken transition-colors"
              >
                <Copy size={14} />
              </button>
            </div>
          </div>

          <div>
            <p class="text-xs leading-5 text-ink-3">
              {m.cloud_created_time()}
            </p>
            <p class="mt-1 text-sm leading-5 text-ink-2">
              {fmtTime(file.createdAt)}
            </p>
          </div>

          <div>
            <p class="text-xs leading-5 text-ink-3">
              {m.last_modified_time()}
            </p>
            <p class="mt-1 text-sm leading-5 text-ink-2">
              {fmtTime(file.updatedAt)}
            </p>
          </div>
          {#if file.fileHash}
            <div>
              <p class="text-xs leading-5 text-ink-3">
                {m.file_hash()}
                {file.hashAlgo ?? "(sha256)"}
              </p>
              <div class="mt-1 flex items-center gap-2">
                <p class="text-sm leading-5 text-ink-2 break-all font-mono">
                  {file.fileHash}
                </p>

                <button
                  onclick={() => copySlug(file.fileHash!)}
                  class="rounded shrink-0 p-1 text-ink-4 hover:text-ink hover:bg-surface-sunken transition-colors"
                >
                  <Copy size={14} />
                </button>
              </div>
            </div>
          {/if}
        </div>
      </section>
    </div>
  {/if}
</Dialog>
