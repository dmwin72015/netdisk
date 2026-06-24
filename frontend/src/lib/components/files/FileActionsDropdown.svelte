<script lang="ts">
  import { DropdownBase } from "$lib/ui/dropdown";
  import {
    Ellipsis,
    Star,
    Eye,
    Download,
    Pencil,
    Trash2,
    Film,
    Info,
    Link,
    Share2,
    FolderInput,
    Lock,
    LockOpen,
    Trash,
  } from "@lucide/svelte";
  import type { NormalizedFile } from "$lib/types/file";
  import * as m from "$lib/paraglide/messages";
  import { authedUrl } from "$lib/utils/format";

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
    onShare,
    onSetDirectoryLock,
    onClearDirectoryLock,
    onForceDeleteDir,
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
    onShare?: (file: NormalizedFile) => void;
    onSetDirectoryLock?: (file: NormalizedFile) => void;
    onClearDirectoryLock?: (file: NormalizedFile) => void;
    onForceDeleteDir?: (file: NormalizedFile) => void;
    triggerClass?: string;
  } = $props();

  let isVideo = $derived(file.mimeType?.startsWith("video/") ?? false);

  let hasAboveItems = $derived(
    (onStar && !file.isSystem) ||
      !file.isDir ||
      !!onShowDetails ||
      (isVideo && !!onAddToMedia)
  );

  let showSeparator = $derived(hasAboveItems && !file.isSystem);

  function downloadFile() {
    const url = authedUrl(downloadUrlFn(file.id));
    const a = document.createElement("a");
    a.href = url;
    a.download = file.name;
    a.click();
    a.remove();
  }
</script>

<DropdownBase.Root>
  <DropdownBase.Trigger
    class="rounded-md p-1.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-3 {triggerClass}"
  >
    <Ellipsis size={16} />
  </DropdownBase.Trigger>
  <DropdownBase.Content sideOffset={4} align="end">
    {#if onStar && !file.isSystem}
      <DropdownBase.Item onSelect={() => onStar(file.id, file.isStarred)}>
        {#snippet icon()}<Star
            size={14}
            class={file.isStarred ? "text-warning" : "text-ink-4"}
            fill={file.isStarred ? "currentColor" : "none"}
          />{/snippet}
        {#snippet children()}{file.isStarred
            ? m.unstar_file()
            : m.star_file()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if !file.isDir}
      <DropdownBase.Item onSelect={() => onPreview(file)}>
        {#snippet icon()}<Eye size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.preview()}{/snippet}
      </DropdownBase.Item>
      <DropdownBase.Item onSelect={downloadFile}>
        {#snippet icon()}<Download size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.download()}{/snippet}
      </DropdownBase.Item>
      {#if onCopyLink}
        <DropdownBase.Item onSelect={() => onCopyLink(file)}>
          {#snippet icon()}<Link size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}{m.copy_url()}{/snippet}
        </DropdownBase.Item>
      {/if}
      {#if onShare}
        <DropdownBase.Item onSelect={() => onShare(file)}>
          {#snippet icon()}<Share2 size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}分享{/snippet}
        </DropdownBase.Item>
      {/if}
    {/if}
    {#if onShowDetails}
      <DropdownBase.Item onSelect={() => onShowDetails(file)}>
        {#snippet icon()}<Info size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.view_details()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if isVideo && onAddToMedia}
      <DropdownBase.Item onSelect={() => onAddToMedia(file)}>
        {#snippet icon()}<Film size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.add_to_media_library()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if showSeparator}
      <DropdownBase.Separator />
    {/if}
    {#if !file.isSystem}
      <DropdownBase.Item onSelect={() => onRename(file.id, file.name)}>
        {#snippet icon()}<Pencil size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.rename()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if onMove && !file.isSystem}
      <DropdownBase.Item onSelect={() => onMove(file)}>
        {#snippet icon()}<FolderInput size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.move_to()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if file.isDir && !file.isSystem}
      {#if file.isLocked && onClearDirectoryLock}
        <DropdownBase.Item onSelect={() => onClearDirectoryLock(file)}>
          {#snippet icon()}<LockOpen size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}取消目录密码{/snippet}
        </DropdownBase.Item>
      {:else if onSetDirectoryLock}
        <DropdownBase.Item onSelect={() => onSetDirectoryLock(file)}>
          {#snippet icon()}<Lock size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}设置目录密码{/snippet}
        </DropdownBase.Item>
      {/if}
    {/if}
    {#if file.isDir && !file.isSystem && onForceDeleteDir}
      <DropdownBase.Item
        variant="destructive"
        onSelect={() => onForceDeleteDir(file)}
      >
        {#snippet icon()}<Trash size={14} class="text-danger" />{/snippet}
        {#snippet children()}强制删除{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if !file.isSystem}
      <DropdownBase.Item
        variant="destructive"
        onSelect={() => onDelete(file.id, file.name)}
      >
        {#snippet icon()}<Trash2 size={14} class="text-danger" />{/snippet}
        {#snippet children()}{m.delete_label()}{/snippet}
      </DropdownBase.Item>
    {/if}
  </DropdownBase.Content>
</DropdownBase.Root>
