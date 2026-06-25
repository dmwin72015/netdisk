<script lang="ts">
  import { getContext } from "svelte";
  import { afterNavigate, goto } from "$app/navigation";
  import { page } from "$app/state";
  import { browser } from "$app/environment";
  import { user, authReady } from "$lib/stores/auth";
  import { downloadUrl } from "$lib/api/files";
  import { toast } from "svelte-sonner";
  import DrivePreview from "$lib/components/DrivePreview.svelte";
  import FileBrowserContent from "$lib/components/files/FileBrowserContent.svelte";
  import ShareDialog from "$lib/components/files/ShareDialog.svelte";
  import FolderUploadDialog from "$lib/components/files/FolderUploadDialog.svelte";
  import RemoteUploadDialog from "$lib/components/files/RemoteUploadDialog.svelte";
  import TextUploadDialog from "$lib/components/files/TextUploadDialog.svelte";
  import ConflictDialog from "$lib/components/files/ConflictDialog.svelte";
  import PasteUploadProvider from "$lib/components/files/PasteUploadProvider.svelte";
  import type { createUploadManager as UploadMgrFn } from "$lib/upload-manager.svelte";
  import { ConflictManager } from "$lib/stores/conflict-manager.svelte";
  import { fileManager } from "$lib/services/fileManager.svelte";
  import { previewManager } from "$lib/services/previewManager.svelte";
  import type { NormalizedFile } from "$lib/types/file";

  type UploadManager = ReturnType<typeof UploadMgrFn>;

  let { children } = $props();

  const upload = getContext<UploadManager>("upload");

  // --- Local dialog state ---
  let remoteUploadOpen = $state(false);
  let textUploadOpen = $state(false);
  let shareOpen = $state(false);
  let shareFiles = $state<NormalizedFile[]>([]);

  // --- Conflict resolution ---
  const conflicts = new ConflictManager();

  // --- Upload integration ---
  $effect(() => {
    upload.updateMaxConcurrent(1);
    upload.setGetCurrentSlug(() => fileManager.currentSlug);
    upload.setOnCompleted(() => fileManager.refresh(false, true));
    upload.setOnNameConflicts((c) => conflicts.onUploadConflicts(c));
  });

  if (browser) {
    $effect(() => {
      const handler = () => {
        const input = document.querySelector<HTMLInputElement>(
          'input[type="file"][multiple]',
        );
        input?.click();
      };
      window.addEventListener("nd:open-file-upload", handler);
      return () => window.removeEventListener("nd:open-file-upload", handler);
    });
  }

  // --- Navigation sync ---
  afterNavigate(() => {
    const slug = page.params.slug ?? "";
    fileManager.setSlug(slug);
  });

  // --- Share handlers ---
  function onShare(file: NormalizedFile) {
    shareFiles = [file];
    shareOpen = true;
  }

  function onBatchShare(files: NormalizedFile[]) {
    shareFiles = files;
    shareOpen = true;
  }

  async function navigateToDir(slug: string) {
    const status = await fileManager.navigateToDir(slug);
    if (status === "navigated") {
      await goto("/files/all/" + slug);
    }
  }
</script>

{#if $authReady && $user}
  <div class="space-y-4 rounded-xl border border-line bg-white p-4 relative">
    <FileBrowserContent
      {onBatchShare}
      onUploadFiles={() => {
        const input = document.querySelector<HTMLInputElement>(
          'input[type="file"][multiple]',
        );
        input?.click();
      }}
      onUploadFolder={() => {
        const input = document.querySelector<HTMLInputElement>(
          'input[type="file"][webkitdirectory]',
        );
        input?.click();
      }}
      onUploadFromURL={() => {
        remoteUploadOpen = true;
      }}
      onUploadText={() => {
        textUploadOpen = true;
      }}
    />
  </div>
{/if}

<ShareDialog bind:open={shareOpen} files={shareFiles} />

{#if previewManager.previewFile}
  <DrivePreview
    id={previewManager.previewFile.slug}
    name={previewManager.previewFile.name}
    mimeType={previewManager.previewFile.mimeType}
    size={previewManager.previewFile.size}
    open={true}
    onOpenChangeComplete={(open) => {
      if (!open) previewManager.close();
    }}
  />
{/if}

<FolderUploadDialog
  files={upload.folderDialogFiles}
  open={upload.folderDialogOpen}
  loading={upload.folderDialogLoading}
  onConfirm={upload.onFolderConfirm}
  onCancel={() => {
    upload.folderDialogOpen = false;
  }}
/>

<RemoteUploadDialog
  bind:open={remoteUploadOpen}
  parentSlug={fileManager.currentSlug}
/>

<TextUploadDialog
  bind:open={textUploadOpen}
  targetLabel={fileManager.currentDirLabel}
  onConfirm={async (file) => {
    await upload.enqueueFiles([file]);
  }}
  onCancel={() => {
    textUploadOpen = false;
  }}
/>

<ConflictDialog
  bind:open={conflicts.open}
  conflicts={conflicts.conflicts}
  onResolve={(results) => conflicts.finish(results)}
  onCancel={() => conflicts.cancel()}
/>

<PasteUploadProvider
  targetLabel={fileManager.currentDirLabel}
  onUpload={(files) => upload.enqueueFiles(files)}
/>

{@render children()}
