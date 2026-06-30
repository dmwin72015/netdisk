<script lang="ts">
  import {
    Search,
    Trash2,
    LoaderCircle,
    HardDrive,
    FileText,
    Hash,
    Link,
    ShieldAlert,
    User,
    Mail,
    Calendar,
    HardDrive as HardDriveIcon,
    Clock,
    X,
  } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import {
    adminCleanupQuery,
    adminCleanupDeleteUserFile,
    adminCleanupDeletePhysicalFile,
    adminGetUser,
    type CleanupQueryResult,
    type CleanupQueryUserFile,
    type AdminUser,
  } from "$lib/api/admin";
  import { fmtSize } from "$lib/utils/format";
  import Dialog from "$lib/ui/dialog/Dialog.svelte";
  import { Popover } from "$lib/ui/popover";
  import * as m from "$lib/paraglide/messages";
  import { searchHistory } from "$lib/stores/search-history.svelte";

  let mode = $state<"slug" | "hash">("slug");
  let slug = $state("");
  let hash = $state("");
  let querying = $state(false);
  let result = $state<CleanupQueryResult | null>(null);
  let error = $state<string | null>(null);

  // Search history
  let showHistory = $state(false);

  let deletingUserFileId = $state<number | null>(null);
  let deletingPhysicalFileId = $state<number | null>(null);
  let deleting = $state(false);
  let confirmDeleteAll = $state(false);
  let pendingDeleteUserFile = $state<CleanupQueryUserFile | null>(null);

  // User detail popover
  let userDetailLoading = $state<Record<number, boolean>>({});
  let userDetailCache = $state<Record<number, AdminUser>>({});

  function getCurrentTerm(): string {
    return mode === "slug" ? slug.trim() : hash.trim();
  }

  function setCurrentTerm(term: string) {
    if (mode === "slug") slug = term;
    else hash = term;
  }

  async function handleQuery() {
    const term = getCurrentTerm();
    if (!term) return;
    searchHistory.addSearch(mode, term);
    showHistory = false;
    querying = true;
    error = null;
    result = null;
    try {
      result = await adminCleanupQuery(
        mode === "slug" ? term : "",
        mode === "hash" ? term : "",
      );
      if (result.userFiles.length === 0 && !result.physicalFile) {
        error = m.admin_cleanup_no_data_for_mode({ mode });
      }
    } catch (e) {
      error = e instanceof Error ? e.message : m.admin_cleanup_query_failed();
      toast.error(error);
    } finally {
      querying = false;
    }
  }

  function handleHistoryClick(term: string) {
    setCurrentTerm(term);
    showHistory = false;
    handleQuery();
  }

  function handleHistoryRemove(e: Event, term: string) {
    e.stopPropagation();
    searchHistory.removeEntry(mode, term);
  }

  function handleHistoryClear() {
    searchHistory.clearHistory(mode);
    showHistory = false;
  }

  function handleInputFocus() {
    showHistory = true;
  }

  function handleInputBlur() {
    setTimeout(() => {
      showHistory = false;
    }, 200);
  }

  async function handleDeleteUserFile(userFileId: number) {
    deletingUserFileId = userFileId;
    deleting = true;
    try {
      const res = await adminCleanupDeleteUserFile(userFileId);
      toast.success(res.message);
      result!.userFiles = result!.userFiles.filter((f) => f.id !== userFileId);
      result!.totalUploads--;
      pendingDeleteUserFile = null;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.admin_cleanup_delete_failed());
    } finally {
      deleting = false;
      deletingUserFileId = null;
    }
  }

  function confirmDeleteUserFile() {
    if (!pendingDeleteUserFile) return;
    handleDeleteUserFile(pendingDeleteUserFile.id);
  }

  async function fetchUserDetail(userId: number) {
    if (userDetailCache[userId]) return userDetailCache[userId];

    userDetailLoading[userId] = true;
    try {
      const user = await adminGetUser(String(userId));
      userDetailCache[userId] = user;
      return user;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.admin_cleanup_query_failed());
      return null;
    } finally {
      userDetailLoading[userId] = false;
    }
  }

  function handleUserPopoverOpen(userId: number) {
    if (!userDetailCache[userId]) {
      void fetchUserDetail(userId);
    }
  }

  async function handleDeletePhysicalFile() {
    if (!result?.physicalFile) return;
    deletingPhysicalFileId = result.physicalFile.id;
    deleting = true;
    try {
      const res = await adminCleanupDeletePhysicalFile(result.physicalFile.id);
      toast.success(res.message);
      confirmDeleteAll = false;
      result = null;
      slug = "";
      hash = "";
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.admin_cleanup_delete_failed());
    } finally {
      deleting = false;
      deletingPhysicalFileId = null;
    }
  }
</script>

<div class="mx-auto max-w-7xl">
  <div class="mb-6">
    <h1 class="text-xl font-semibold text-ink">{m.admin_cleanup_title()}</h1>
    <p class="mt-1 text-sm text-ink-3">
      {m.admin_cleanup_subtitle()}
    </p>
  </div>

  <!-- Mode toggle + Query form -->
  <div class="rounded-xl border border-line bg-surface p-5">
    <div class="mb-4 flex gap-2">
      <button
        onclick={() => (mode = "slug")}
        class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {mode === 'slug' ? 'text-primary bg-primary/10' : 'text-ink-3 hover:bg-surface-sunken'}"
      >
        <Link size={14} />
        {m.admin_cleanup_by_slug()}
      </button>
      <button
        onclick={() => (mode = "hash")}
        class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {mode === 'hash' ? 'text-primary bg-primary/10' : 'text-ink-3 hover:bg-surface-sunken'}"
      >
        <Hash size={14} />
        {m.admin_cleanup_by_hash()}
      </button>
    </div>

    <div class="flex gap-3">
      <div class="relative flex-1">
        <Search
          size={16}
          class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-4 pointer-events-none"
        />
        {#if mode === "slug"}
          <input
            type="text"
            bind:value={slug}
            placeholder={m.admin_cleanup_slug_placeholder()}
            class="w-full rounded-lg border border-line bg-surface-sunken pl-9 pr-4 py-2.5 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            onfocus={handleInputFocus}
            onblur={handleInputBlur}
            onkeydown={(e) => e.key === "Enter" && handleQuery()}
          />
        {:else}
          <input
            type="text"
            bind:value={hash}
            placeholder={m.admin_cleanup_hash_placeholder()}
            class="w-full rounded-lg border border-line bg-surface-sunken pl-9 pr-4 py-2.5 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
            onfocus={handleInputFocus}
            onblur={handleInputBlur}
            onkeydown={(e) => e.key === "Enter" && handleQuery()}
          />
        {/if}

        <!-- Search history dropdown -->
        {#if showHistory && searchHistory.getHistory(mode).length > 0}
          <div
            class="absolute left-0 right-0 top-full z-50 mt-1 overflow-hidden rounded-lg border border-line bg-surface shadow-pop"
          >
            <div class="px-3 py-2 text-xs font-medium text-ink-4 flex items-center gap-1.5 border-b border-line">
              <Clock size={12} />
              {m.admin_cleanup_recent_searches()}
            </div>
            <div class="max-h-64 overflow-y-auto py-1">
              {#each searchHistory.getHistory(mode) as term}
                <button
                  type="button"
                  class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-ink transition-colors hover:bg-surface-sunken"
                  onmousedown={() => handleHistoryClick(term)}
                >
                  <Clock size={13} class="shrink-0 text-ink-4" />
                  <span class="min-w-0 flex-1 truncate">{term}</span>
                  <span
                    role="button"
                    tabindex="-1"
                    class="shrink-0 rounded p-0.5 text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
                    onmousedown={(e) => handleHistoryRemove(e, term)}
                  >
                    <X size={12} />
                  </span>
                </button>
              {/each}
            </div>
            <button
              type="button"
              class="flex w-full items-center gap-2 border-t border-line px-3 py-2 text-xs text-ink-4 transition-colors hover:text-ink-3"
              onmousedown={handleHistoryClear}
            >
              <Trash2 size={12} />
              {m.admin_cleanup_clear_history()}
            </button>
          </div>
        {/if}
      </div>
      <button
        onclick={handleQuery}
        disabled={
          querying ||
          (mode === "slug" && !slug.trim()) ||
          (mode === "hash" && !hash.trim())
        }
        class="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-on transition-colors hover:bg-primary/90 disabled:opacity-50"
      >
        {#if querying}
          <LoaderCircle size={16} class="animate-spin" />
        {:else}
          <Search size={16} />
        {/if}
        {m.admin_cleanup_query_btn()}
      </button>
    </div>
  </div>

  <!-- Error state -->
  {#if error && !result}
    <div class="mt-6 rounded-xl border border-line bg-surface p-8 text-center">
      <ShieldAlert size={32} class="mx-auto text-ink-4" />
      <p class="mt-2 text-sm text-ink-3">{error}</p>
    </div>
  {/if}

  <!-- Results -->
  {#if result}
    <!-- Summary -->
    <div class="mt-6 grid grid-cols-3 gap-4">
      <div class="rounded-xl border border-line bg-surface p-4">
        <p class="text-xs text-ink-4">{m.admin_cleanup_total_uploads()}</p>
        <p class="mt-1 text-2xl font-semibold text-ink">{result.totalUploads}</p>
      </div>
      <div class="rounded-xl border border-line bg-surface p-4">
        <p class="text-xs text-ink-4">{m.admin_cleanup_unique_users()}</p>
        <p class="mt-1 text-2xl font-semibold text-ink">{result.uniqueUsers}</p>
      </div>
      <div class="rounded-xl border border-line bg-surface p-4">
        <p class="text-xs text-ink-4">{m.admin_cleanup_total_size()}</p>
        <p class="mt-1 text-2xl font-semibold text-ink">
          {fmtSize(
            result.userFiles.reduce((a, f) => a + f.fileSize, 0) +
              (result.physicalFile?.fileSize ?? 0),
          )}
        </p>
      </div>
    </div>

    <!-- Physical File Card -->
    {#if result.physicalFile}
      {@const pf = result.physicalFile}
      <div class="mt-6 rounded-xl border border-line bg-surface overflow-hidden">
        <div class="flex items-center gap-2 border-b border-line px-5 py-3">
          <HardDrive size={16} class="text-ink-4" />
          <h2 class="text-sm font-medium text-ink">{m.admin_cleanup_physical_file()}</h2>
          <span class="text-xs text-ink-4">#{pf.id}</span>
          {#if pf.fileExists}
            <span class="ml-auto text-xs text-success">{m.admin_cleanup_file_on_disk()}</span>
          {:else}
            <span class="ml-auto text-xs text-danger">{m.admin_cleanup_missing_on_disk()}</span>
          {/if}
        </div>
        <div class="px-5 py-4 space-y-3">
          <div class="flex items-center justify-between gap-4">
            <span class="shrink-0 text-xs text-ink-4">{m.admin_cleanup_field_hash()}</span>
            <span
              class="min-w-0 flex-1 truncate text-right font-mono text-xs text-ink-2"
              title={pf.fileHash}
            >{pf.fileHash}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="shrink-0 text-xs text-ink-4">{m.admin_cleanup_field_size()}</span>
            <span class="text-sm text-ink">{fmtSize(pf.fileSize)}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="shrink-0 text-xs text-ink-4">{m.admin_cleanup_field_mime()}</span>
            <span class="text-sm text-ink">{pf.mimeType}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="shrink-0 text-xs text-ink-4">{m.admin_cleanup_field_storage_path()}</span>
            <span
              class="min-w-0 flex-1 truncate text-right font-mono text-xs text-ink-3"
              title={pf.storagePath}
            >{pf.storagePath}</span>
          </div>
        </div>
      </div>
    {/if}

    <!-- User Files Table -->
    <div class="mt-6 rounded-xl border border-line bg-surface overflow-hidden">
      <div class="flex items-center gap-2 border-b border-line px-5 py-3">
        <FileText size={16} class="text-ink-4" />
        <h2 class="text-sm font-medium text-ink">{m.admin_cleanup_user_files()}</h2>
        <span class="text-xs text-ink-4">({result.userFiles.length})</span>
        {#if result.physicalFile}
          <button
            onclick={() => (confirmDeleteAll = true)}
            disabled={deleting}
            class="ml-auto inline-flex items-center gap-1.5 rounded-lg bg-danger px-3 py-1.5 text-xs font-medium text-primary-on transition-colors hover:bg-danger/90 disabled:opacity-50"
          >
            <Trash2 size={13} />
            {m.admin_cleanup_delete_all()}
          </button>
        {/if}
      </div>
      {#if result.userFiles.length === 0}
        <div class="p-8 text-center">
          <p class="text-sm text-ink-3">{m.admin_cleanup_no_user_files()}</p>
        </div>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-line bg-surface-sunken/50 text-left text-xs text-ink-4">
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_id()}</th>
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_slug()}</th>
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_filename()}</th>
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_user()}</th>
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_size()}</th>
                <th class="px-5 py-3 font-medium">{m.admin_cleanup_col_created()}</th>
                <th class="px-5 py-3 font-medium w-28 whitespace-nowrap">{m.admin_cleanup_col_actions()}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-line">
              {#each result.userFiles as uf}
                <tr class="transition-colors hover:bg-surface-sunken/30">
                  <td class="px-5 py-3 font-mono text-xs text-ink-3">{uf.id}</td>
                  <td class="px-5 py-3 font-mono text-xs text-primary max-w-[120px] truncate">{uf.slug}</td>
                  <td class="px-5 py-3 text-ink">{uf.fileName}</td>
                  <td class="px-5 py-3 text-ink-3">
                    <Popover
                      onOpenChange={(o) => o && handleUserPopoverOpen(uf.userId)}
                      side="right"
                      sideOffset={8}
                      align="end"
                      triggerClass="cursor-pointer hover:text-ink transition-colors"
                      contentClass="w-72 rounded-xl border border-line bg-surface shadow-dialog p-4"
                    >
                      {#snippet trigger()}
                        <span class="inline-flex items-center gap-1.5">
                          <User size={13} class="text-ink-4" />
                          {uf.username}
                        </span>
                      {/snippet}
                      {#if userDetailLoading[uf.userId]}
                        <div class="flex items-center justify-center py-4">
                          <LoaderCircle size={16} class="animate-spin text-ink-4" />
                        </div>
                      {:else if userDetailCache[uf.userId]}
                        {@const user = userDetailCache[uf.userId]}
                        <div class="space-y-3">
                          <div class="flex items-center gap-2 border-b border-line pb-3">
                            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary-soft text-primary">
                              <User size={20} />
                            </div>
                            <div class="min-w-0">
                              <p class="text-sm font-medium text-ink truncate">{user.username}</p>
                              {#if user.profile?.displayName}
                                <p class="text-xs text-ink-3 truncate">{user.profile.displayName}</p>
                              {/if}
                            </div>
                          </div>
                          <div class="space-y-2 text-xs">
                            <div class="flex items-center gap-2 text-ink-3">
                              <Mail size={12} class="shrink-0 text-ink-4" />
                              <span class="truncate">{user.email}</span>
                            </div>
                            <div class="flex items-center gap-2 text-ink-3">
                              <Calendar size={12} class="shrink-0 text-ink-4" />
                              <span>{new Date(user.createdAt * 1000).toLocaleString()}</span>
                            </div>
                            <div class="flex items-center gap-2 text-ink-3">
                              <HardDriveIcon size={12} class="shrink-0 text-ink-4" />
                              <span>{fmtSize(user.usedBytes)} / {fmtSize(user.totalBytes)}</span>
                            </div>
                            <div class="flex items-center gap-2">
                              <span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-medium {user.role === 'admin' ? 'bg-danger/10 text-danger' : 'bg-primary/10 text-primary'}">
                                {user.role === 'admin' ? m.admin_role_admin() : m.admin_role_user()}
                              </span>
                              <span class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-medium {user.status === 1 ? 'bg-success/10 text-success' : 'bg-danger/10 text-danger'}">
                                {user.status === 1 ? m.admin_status_active() : m.admin_status_inactive()}
                              </span>
                            </div>
                          </div>
                        </div>
                      {/if}
                    </Popover>
                  </td>
                  <td class="px-5 py-3 text-ink-2">{fmtSize(uf.fileSize)}</td>
                  <td class="px-5 py-3 text-ink-3 text-xs whitespace-nowrap">
                    {new Date(uf.createdAt * 1000).toISOString().slice(0, 19).replace('T', ' ')}
                  </td>
                  <td class="px-5 py-3 whitespace-nowrap">
                    <button
                      onclick={() => (pendingDeleteUserFile = uf)}
                      disabled={deleting && deletingUserFileId === uf.id}
                      class="inline-flex items-center gap-1 rounded-md bg-danger/10 px-2 py-1 text-xs font-medium text-danger whitespace-nowrap transition-colors hover:bg-danger/20 disabled:opacity-40"
                    >
                      {#if deleting && deletingUserFileId === uf.id}
                        <LoaderCircle size={12} class="shrink-0 animate-spin" />
                      {:else}
                        <Trash2 size={12} class="shrink-0" />
                      {/if}
                      {m.admin_cleanup_delete()}
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Delete All confirm dialog -->
<Dialog
  open={confirmDeleteAll}
  onOpenChange={(o) => (confirmDeleteAll = o)}
  footer={false}
  title={m.admin_cleanup_delete_all_title()}
  bodyClass="p-0"
>
  <p class="px-5 pb-4 pt-5 text-sm text-ink-3">
    {m.admin_cleanup_delete_all_desc({ n: result?.userFiles.length ?? 0 })}
  </p>
  <div class="flex justify-end gap-3 px-5 pb-5">
    <button
      onclick={() => (confirmDeleteAll = false)}
      class="rounded-lg border border-line px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
    >
      {m.admin_cleanup_cancel()}
    </button>
    <button
      onclick={handleDeletePhysicalFile}
      disabled={deleting}
      class="inline-flex items-center gap-2 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-danger/90 disabled:opacity-50"
    >
      {#if deleting}
        <LoaderCircle size={16} class="animate-spin" />
      {:else}
        <Trash2 size={16} class="shrink-0" />
      {/if}
      {m.admin_cleanup_delete_all()}
    </button>
  </div>
</Dialog>

<!-- Delete single user file confirm dialog -->
<Dialog
  open={pendingDeleteUserFile !== null}
  onOpenChange={(o) => {
    if (!o && !deleting) pendingDeleteUserFile = null;
  }}
  footer={false}
  title={m.admin_cleanup_delete_user_file_title()}
  bodyClass="p-0"
>
  <p class="px-5 pb-4 pt-5 text-sm text-ink-3">
    {m.admin_cleanup_delete_user_file_desc({
      username: pendingDeleteUserFile?.username ?? "",
      fileName: pendingDeleteUserFile?.fileName ?? "",
    })}
  </p>
  <div class="flex justify-end gap-3 px-5 pb-5">
    <button
      onclick={() => (pendingDeleteUserFile = null)}
      disabled={deleting}
      class="rounded-lg border border-line px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken disabled:opacity-50"
    >
      {m.admin_cleanup_cancel()}
    </button>
    <button
      onclick={confirmDeleteUserFile}
      disabled={deleting}
      class="inline-flex items-center gap-2 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-danger/90 disabled:opacity-50"
    >
      {#if deleting}
        <LoaderCircle size={16} class="animate-spin" />
      {:else}
        <Trash2 size={16} class="shrink-0" />
      {/if}
      {m.admin_cleanup_confirm_delete()}
    </button>
  </div>
</Dialog>
