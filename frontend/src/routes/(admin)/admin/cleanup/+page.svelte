<script lang="ts">
  import { onMount } from "svelte";
  import {
    Search,
    Trash2,
    LoaderCircle,
    AlertTriangle,
    HardDrive,
    Upload,
    User,
    FileText,
    X,
  } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import { adminCleanupFile, type CleanupFileResult } from "$lib/api/admin";
  import { fmtSize } from "$lib/utils/format";
  import Dialog from "$lib/ui/dialog/Dialog.svelte";
  import * as m from "$lib/paraglide/messages";

  let slug = $state("");
  let querying = $state(false);
  let result = $state<CleanupFileResult | null>(null);
  let showConfirm = $state(false);
  let cleaning = $state(false);

  async function handleQuery() {
    if (!slug.trim()) return;
    querying = true;
    result = null;
    try {
      result = await adminCleanupFile(slug.trim());
    } catch (e) {
      toast.error(m.admin_load_failed());
    } finally {
      querying = false;
    }
  }

  async function handleCleanup() {
    if (!result) return;
    cleaning = true;
    try {
      result = await adminCleanupFile(result.slug, true);
      toast.success(m.admin_cleanup_success());
    } catch (e) {
      toast.error(m.admin_cleanup_failed());
    } finally {
      cleaning = false;
      showConfirm = false;
    }
  }

  $inspect(result);
</script>

<div class="mx-auto max-w-3xl">
  <div class="mb-6">
    <h1 class="text-xl font-semibold text-ink">{m.admin_cleanup_title()}</h1>
    <p class="mt-1 text-sm text-ink-3">{m.admin_cleanup_desc()}</p>
  </div>

  <!-- Query form -->
  <div class="rounded-xl border border-line bg-surface p-5">
    <div class="flex gap-3">
      <div class="relative flex-1">
        <Search
          size={16}
          class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-4"
        />
        <input
          type="text"
          bind:value={slug}
          placeholder={m.admin_cleanup_placeholder()}
          class="w-full rounded-lg border border-line bg-surface-sunken pl-9 pr-4 py-2.5 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none"
          onkeydown={(e) => e.key === "Enter" && handleQuery()}
        />
      </div>
      <button
        onclick={handleQuery}
        disabled={querying || !slug.trim()}
        class="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-primary/90 disabled:opacity-50"
      >
        {#if querying}
          <LoaderCircle size={16} class="animate-spin" />
        {:else}
          <Search size={16} />
        {/if}
        {m.admin_cleanup_query()}
      </button>
    </div>
  </div>

  <!-- Results -->
  {#if result}
    <div class="mt-6 space-y-4">
      {#if !result || ((result.uploadTasks?.length ?? 0) === 0 && (result.userFiles?.length ?? 0) === 0 && (result.physicalFiles?.length ?? 0) === 0)}
        <div class="rounded-xl border border-line bg-surface p-8 text-center">
          <p class="text-sm text-ink-3">{m.admin_cleanup_no_data()}</p>
        </div>
      {:else}
        <!-- Upload Tasks -->
        {#if (result.uploadTasks?.length ?? 0) > 0}
          <div class="rounded-xl border border-line bg-surface overflow-hidden">
            <div class="flex items-center gap-2 border-b border-line px-5 py-3">
              <Upload size={16} class="text-ink-4" />
              <h2 class="text-sm font-medium text-ink">
                {m.admin_cleanup_upload_tasks()}
              </h2>
              <span class="text-xs text-ink-4"
                >({result.uploadTasks?.length ?? 0})</span
              >
            </div>
            <div class="divide-y divide-line">
              {#each result.uploadTasks as task}
                <div class="px-5 py-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm font-medium text-ink"
                      >{task.originalName || task.id}</span
                    >
                    <span class="text-xs text-ink-4">{task.status}</span>
                  </div>
                  <div class="mt-1 flex gap-4 text-xs text-ink-3">
                    <span class="flex items-center gap-1"
                      ><User size={12} />{task.username ||
                        "ID " + task.ownerUserId}</span
                    >
                    <span>Size: {fmtSize(task.fileSize)}</span>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- User Files -->
        {#if (result.userFiles?.length ?? 0) > 0}
          <div class="rounded-xl border border-line bg-surface overflow-hidden">
            <div class="flex items-center gap-2 border-b border-line px-5 py-3">
              <User size={16} class="text-ink-4" />
              <h2 class="text-sm font-medium text-ink">
                {m.admin_cleanup_user_files()}
              </h2>
              <span class="text-xs text-ink-4"
                >({result.userFiles?.length ?? 0})</span
              >
            </div>
            <div class="divide-y divide-line">
              {#each result.userFiles as uf}
                <div class="px-5 py-3">
                  <div class="flex items-center justify-between">
                    <span class="text-sm font-medium text-ink"
                      >{uf.fileName}</span
                    >
                    <span class="text-xs text-ink-4">by {uf.username}</span>
                  </div>
                  <div class="mt-1 flex gap-4 text-xs text-ink-3">
                    <span>User ID: {uf.userId}</span>
                    <span>Size: {fmtSize(uf.fileSize)}</span>
                    {#if uf.physicalFileId}
                      <span>Physical: #{uf.physicalFileId}</span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Physical Files -->
        {#if (result.physicalFiles?.length ?? 0) > 0}
          <div class="rounded-xl border border-line bg-surface overflow-hidden">
            <div class="flex items-center gap-2 border-b border-line px-5 py-3">
              <HardDrive size={16} class="text-ink-4" />
              <h2 class="text-sm font-medium text-ink">
                {m.admin_cleanup_physical_files()}
              </h2>
              <span class="text-xs text-ink-4"
                >({result.physicalFiles?.length ?? 0})</span
              >
            </div>
            <div class="divide-y divide-line">
              {#each result.physicalFiles as pf}
                <div class="px-5 py-3">
                  <div class="flex items-center justify-between">
                    <span class="font-mono text-xs text-ink-2 break-all"
                      >{pf.fileHash}</span
                    >
                    {#if pf.refCount == null || pf.refCount === 0}
                      <span class="text-xs text-success">
                        {m.admin_cleanup_orphaned()}
                      </span>
                    {:else}
                      <span
                        class="text-xs text-warning"
                        data-count={pf.refCount}
                      >
                        {m.admin_cleanup_in_use({ n: pf.refCount })}
                      </span>
                    {/if}
                  </div>
                  <div class="mt-1 flex gap-4 text-xs text-ink-3">
                    <span>Size: {fmtSize(pf.fileSize)}</span>
                    <span>Refs: {pf.refCount ?? 0}</span>
                  </div>
                  <div class="mt-1 font-mono text-[10px] text-ink-4 break-all">
                    {pf.storagePath}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Action buttons -->
        <div
          class="flex items-center gap-3 rounded-xl border border-warning/30 bg-warning-soft/50 p-4"
        >
          <AlertTriangle size={20} class="text-warning shrink-0" />
          <div class="flex-1">
            <p class="text-sm text-ink-2">{m.admin_cleanup_warning()}</p>
          </div>
          <button
            onclick={() => (showConfirm = true)}
            disabled={cleaning}
            class="inline-flex items-center gap-2 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-danger/90 disabled:opacity-50"
          >
            {#if cleaning}
              <LoaderCircle size={16} class="animate-spin" />
            {:else}
              <Trash2 size={16} />
            {/if}
            {m.admin_cleanup_btn()}
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Confirm dialog -->
<Dialog open={showConfirm} onOpenChange={(o) => (showConfirm = o)}>
  <div class="p-6">
    <h2 class="text-lg font-semibold text-ink">
      {m.admin_cleanup_confirm_title()}
    </h2>
    <p class="mt-2 text-sm text-ink-3">{m.admin_cleanup_confirm_desc()}</p>
    <div class="mt-4 flex justify-end gap-3">
      <button
        onclick={() => (showConfirm = false)}
        class="rounded-lg border border-line px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
      >
        {m.admin_cleanup_cancel()}
      </button>
      <button
        onclick={handleCleanup}
        disabled={cleaning}
        class="inline-flex items-center gap-2 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-danger/90 disabled:opacity-50"
      >
        {#if cleaning}
          <LoaderCircle size={16} class="animate-spin" />
        {:else}
          <Trash2 size={16} />
        {/if}
        {m.admin_cleanup_confirm_btn()}
      </button>
    </div>
  </div>
</Dialog>
