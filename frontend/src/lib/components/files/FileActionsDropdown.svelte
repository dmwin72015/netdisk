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
    FingerprintPattern,
  } from "@lucide/svelte";
  import type { NormalizedFile } from "$lib/types/file";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { lockManager } from "$lib/services/lockManager.svelte";
  import { previewManager } from "$lib/services/previewManager.svelte";
  import * as m from "$lib/paraglide/messages";
  import { authedUrl } from "$lib/utils/format";

  let {
    file,
    onMove,
    onShowDetails,
    onCopyLink,
    onCopyHash,
    onShare,
    triggerClass = "",
  }: {
    file: NormalizedFile;
    onMove?: (file: NormalizedFile) => void;
    onShowDetails?: (file: NormalizedFile) => void;
    onCopyLink?: (file: NormalizedFile) => void;
    onCopyHash?: (file: NormalizedFile) => void;
    onShare?: (file: NormalizedFile) => void;
    triggerClass?: string;
  } = $props();

  let isVideo = $derived(file.mimeType?.startsWith("video/") ?? false);

  let hasAboveItems = $derived(
    !file.isSystem || !file.isDir || !!onShowDetails || (isVideo && true),
  );

  let showSeparator = $derived(hasAboveItems && !file.isSystem);

  function downloadFile() {
    const url = authedUrl(fileManager.getDownloadUrl(file.id));
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
    {#if !file.isSystem}
      <DropdownBase.Item
        onSelect={() => fileManager.toggleStar(file.id, file.isStarred)}
      >
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
      <DropdownBase.Item onSelect={() => previewManager.open(file)}>
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
      {#if onCopyHash && file.fileHash}
        <DropdownBase.Item onSelect={() => onCopyHash!(file)}>
          {#snippet icon()}
            <FingerprintPattern size={14} class="text-ink-4" />
          {/snippet}
          {#snippet children()}{m.copy_hash()}{/snippet}
        </DropdownBase.Item>
      {/if}
      {#if onShare}
        <DropdownBase.Item onSelect={() => onShare(file)}>
          {#snippet icon()}<Share2 size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}{m.share_file()}{/snippet}
        </DropdownBase.Item>
      {/if}
    {/if}
    {#if onShowDetails && !lockManager.isEffectivelyLocked(file)}
      <DropdownBase.Item onSelect={() => onShowDetails(file)}>
        {#snippet icon()}<Info size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.view_details()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if isVideo}
      <DropdownBase.Item onSelect={() => fileManager.addToMedia(file)}>
        {#snippet icon()}<Film size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.add_to_media_library()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if showSeparator}
      <DropdownBase.Separator />
    {/if}
    {#if !file.isSystem && !lockManager.isEffectivelyLocked(file)}
      <DropdownBase.Item
        onSelect={() => fileManager.rename(file.id, file.name)}
      >
        {#snippet icon()}<Pencil size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.rename()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if onMove && !file.isSystem && !lockManager.isEffectivelyLocked(file)}
      <DropdownBase.Item onSelect={() => onMove(file)}>
        {#snippet icon()}<FolderInput size={14} class="text-ink-4" />{/snippet}
        {#snippet children()}{m.move_to()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if file.isDir && !file.isSystem}
      {#if file.hasPassword}
        <DropdownBase.Item onSelect={() => lockManager.clearLock(file)}>
          {#snippet icon()}<LockOpen size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}{m.clear_dir_password()}{/snippet}
        </DropdownBase.Item>
      {:else}
        <DropdownBase.Item onSelect={() => lockManager.lock(file)}>
          {#snippet icon()}<Lock size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}{m.set_dir_password()}{/snippet}
        </DropdownBase.Item>
      {/if}
    {/if}
    {#if file.isDir && !file.isSystem && !lockManager.isEffectivelyLocked(file)}
      <DropdownBase.Item
        variant="destructive"
        onSelect={() => fileManager.forceRemoveDir(file)}
      >
        {#snippet icon()}<Trash size={14} class="text-danger" />{/snippet}
        {#snippet children()}{m.force_delete()}{/snippet}
      </DropdownBase.Item>
    {/if}
    {#if !file.isSystem && !lockManager.isEffectivelyLocked(file)}
      <DropdownBase.Item
        variant="destructive"
        onSelect={() => fileManager.remove(file.id, file.name)}
      >
        {#snippet icon()}<Trash2 size={14} class="text-danger" />{/snippet}
        {#snippet children()}{m.delete_label()}{/snippet}
      </DropdownBase.Item>
    {/if}
  </DropdownBase.Content>
</DropdownBase.Root>
