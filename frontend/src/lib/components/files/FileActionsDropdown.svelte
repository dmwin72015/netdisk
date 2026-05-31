<script lang="ts">
  import { DropdownBase } from "$lib/ui/dropdown";
  import {
    MoreHorizontal,
    Star,
    Eye,
    Download,
    Pencil,
    Trash2,
    Film,
    Info,
    Link,
    FolderInput,
  } from "@lucide/svelte";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";

  let {
    file,
    downloadUrlFn,
    onStar,
    onPreview,
    onRename,
    onDelete,
    onMove,
    onAddToMedia,
    onShowDetails,
    onCopyLink,
    triggerClass = "",
  }: {
    file: NormalizedFile;
    downloadUrlFn: (id: string) => string;
    onStar?: (id: string, starred: boolean) => void;
    onPreview: (file: NormalizedFile) => void;
    onRename: (id: string, name: string) => void;
    onDelete: (id: string, name: string) => void;
    onMove?: (file: NormalizedFile) => void;
    onAddToMedia?: (file: NormalizedFile) => void;
    onShowDetails?: (file: NormalizedFile) => void;
    onCopyLink?: (file: NormalizedFile) => void;
    triggerClass?: string;
  } = $props();

  let isVideo = $derived(file.mimeType?.startsWith("video/") ?? false);
</script>

<DropdownBase.Root>
  <DropdownBase.Trigger
    class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 {triggerClass}"
  >
    <MoreHorizontal size={16} />
  </DropdownBase.Trigger>
  <DropdownBase.Content sideOffset={4} align="end">
    {#if onStar && !file.isSystem}
      <DropdownBase.Item onSelect={() => onStar(file.id, file.isStarred)}>
        {#snippet icon()}
          <Star
            size={14}
            class={file.isStarred ? "text-amber-400" : "text-gray-400"}
            fill={file.isStarred ? "currentColor" : "none"}
          />
        {/snippet}
        {file.isStarred ? m.unstar_file() : m.star_file()}
      </DropdownBase.Item>
    {/if}
    {#if !file.isDir}
      <DropdownBase.Item onSelect={() => onPreview(file)}>
        {#snippet icon()}
          <Eye size={14} class="text-gray-400" />
        {/snippet}
        {m.preview()}
      </DropdownBase.Item>
      <DropdownBase.Item>
        {#snippet icon()}
          <Download size={14} class="text-gray-400" />
        {/snippet}
        <a
          href={downloadUrlFn(file.id)}
          download={file.name}
          class="flex items-center gap-2.5"
        >
          {m.download()}
        </a>
      </DropdownBase.Item>
      {#if onCopyLink}
        <DropdownBase.Item onSelect={() => onCopyLink(file)}>
          {#snippet icon()}
            <Link size={14} class="text-gray-400" />
          {/snippet}
          {m.copy_url()}
        </DropdownBase.Item>
      {/if}
    {/if}
    {#if onShowDetails}
      <DropdownBase.Item onSelect={() => onShowDetails(file)}>
        {#snippet icon()}
          <Info size={14} class="text-gray-400" />
        {/snippet}
        {m.view_details()}
      </DropdownBase.Item>
    {/if}
    {#if isVideo && onAddToMedia}
      <DropdownBase.Item onSelect={() => onAddToMedia(file)}>
        {#snippet icon()}
          <Film size={14} class="text-gray-400" />
        {/snippet}
        {m.add_to_media_library()}
      </DropdownBase.Item>
    {/if}
    <DropdownBase.Separator />
    {#if !file.isSystem}
      <DropdownBase.Item onSelect={() => onRename(file.id, file.name)}>
        {#snippet icon()}
          <Pencil size={14} class="text-gray-400" />
        {/snippet}
        {m.rename()}
      </DropdownBase.Item>
    {/if}
    {#if onMove && !file.isSystem}
      <DropdownBase.Item onSelect={() => onMove(file)}>
        {#snippet icon()}
          <FolderInput size={14} class="text-gray-400" />
        {/snippet}
        {m.move_to()}
      </DropdownBase.Item>
    {/if}
    {#if !file.isSystem}
      <DropdownBase.Item
        variant="destructive"
        onSelect={() => onDelete(file.id, file.name)}
      >
        {#snippet icon()}
          <Trash2 size={14} />
        {/snippet}
        {m.delete_label()}
      </DropdownBase.Item>
    {/if}
  </DropdownBase.Content>
</DropdownBase.Root>
