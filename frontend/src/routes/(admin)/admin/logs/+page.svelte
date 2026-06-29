<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { getLocale } from "$lib/paraglide/runtime";
  import { goto } from "$app/navigation";
  import {
    Search,
    X,
    ChevronLeft,
    ChevronRight,
    LoaderCircle,
    ChevronDown,
    ExternalLink,
    Eye,
  } from "@lucide/svelte";
  import { Select } from "bits-ui";
  import { toast } from "svelte-sonner";
  import DateRangePicker from "$lib/ui/date-range-picker/DateRangePicker.svelte";
  import Dialog from "$lib/ui/dialog/Dialog.svelte";
  import {
    adminListActivityLogs,
    adminListActivityLogActions,
    type AdminActivityLog,
  } from "$lib/api/admin";
  import * as m from "$lib/paraglide/messages";

  const PAGE_SIZE = 20;

  let logs = $state<AdminActivityLog[]>([]);
  let total = $state(0);
  let offset = $state(0);
  let loading = $state(true);
  let actions = $state<{action: string; label: string}[]>([]);
  let actionFilter = $state("");
  let userIdFilter = $state("");
  let ipFilter = $state("");
  let dateRange = $state<{ start: Date | null; end: Date | null }>({
    start: null,
    end: null,
  });

  let detailOpen = $state(false);
  let detailLog = $state<AdminActivityLog | null>(null);

  let currentPage = $derived(Math.floor(offset / PAGE_SIZE) + 1);
  let totalPages = $derived(Math.ceil(total / PAGE_SIZE));

  onMount(() => {
    if (!browser) return;
    loadActions();
    loadLogs();
  });

  async function loadActions() {
    try {
      actions = await adminListActivityLogActions();
    } catch {
      // non-critical
    }
  }

  async function loadLogs() {
    loading = true;
    try {
      const res = await adminListActivityLogs(
        PAGE_SIZE,
        offset,
        actionFilter || undefined,
        userIdFilter || undefined,
        ipFilter || undefined,
        dateRange.start?.toISOString(),
        dateRange.end?.toISOString(),
      );
      logs = res.items;
      total = res.total;
    } catch {
      toast.error("Failed to load activity logs");
    } finally {
      loading = false;
    }
  }

  function goPage(page: number) {
    offset = (page - 1) * PAGE_SIZE;
    loadLogs();
  }

  function handleFilter() {
    offset = 0;
    loadLogs();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") handleFilter();
  }

  function openDetail(log: AdminActivityLog) {
    detailLog = log;
    detailOpen = true;
  }

  function fmtTime(ts: string): string {
    return new Date(ts).toLocaleString();
  }

  function handleDetailUserClick() {
    const id = detailLog?.userId;
    if (id != null) {
      detailOpen = false;
      goto(`/admin/users/${id}`);
    }
  }

  function fmtExtra(item: AdminActivityLog): string {
    if (!item.extra) return "";
    try {
      return JSON.stringify(item.extra, null, 2);
    } catch {
      return "";
    }
  }
</script>

<div class="space-y-5">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-xl font-bold text-ink">{m.admin_logs_title()}</h1>
      <p class="mt-0.5 text-sm text-ink-4">{m.admin_logs_subtitle()}</p>
    </div>
  </div>

  <!-- Filters -->
  <div class="flex flex-wrap items-center gap-3">
    <div class="relative min-w-[120px] max-w-[160px]">
      <input
        bind:value={userIdFilter}
        onkeydown={handleKeydown}
        placeholder={m.admin_logs_user_placeholder()}
        class="w-full rounded-lg border border-line bg-surface py-2 pl-3 pr-8 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
      {#if userIdFilter}
        <button
          type="button"
          onclick={() => (userIdFilter = "")}
          class="absolute right-2 top-1/2 flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
        >
          <X size={12} />
        </button>
      {/if}
    </div>
    <Select.Root type="single" bind:value={actionFilter}>
      <Select.Trigger
        class="flex items-center justify-between gap-2 rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink-3 min-w-[150px] data-[placeholder]:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      >
        <Select.Value placeholder={m.admin_logs_all_actions()} />
        {#if actionFilter}
          <button
            type="button"
            onclick={(e) => { e.stopPropagation(); actionFilter = ""; }}
            class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
          >
            <X size={12} />
          </button>
        {/if}
        <ChevronDown size={14} class="text-ink-4" />
      </Select.Trigger>
      <Select.Content
        class="z-50 overflow-hidden rounded-lg border border-line bg-surface p-1 shadow-pop"
        sideOffset={4}
        align="start"
      >
        {#each actions as act}
          <Select.Item
            value={act.action}
            class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm text-ink outline-none transition-colors hover:bg-primary/5 data-[highlighted]:bg-primary-soft data-[highlighted]:text-primary data-[state=checked]:bg-primary-soft data-[state=checked]:text-primary data-[state=checked]:font-semibold"
          >
            {act.label}
          </Select.Item>
        {/each}
      </Select.Content>
    </Select.Root>
    <div class="relative min-w-[140px] max-w-[180px]">
      <input
        bind:value={ipFilter}
        onkeydown={handleKeydown}
        placeholder={m.admin_logs_ip_placeholder()}
        class="w-full rounded-lg border border-line bg-surface py-2 pl-3 pr-8 text-sm text-ink placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
      {#if ipFilter}
        <button
          type="button"
          onclick={() => (ipFilter = "")}
          class="absolute right-2 top-1/2 flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full text-ink-4 transition-colors hover:bg-surface-sunken hover:text-ink-2"
        >
          <X size={12} />
        </button>
      {/if}
    </div>
    <DateRangePicker
      value={dateRange}
      onValueChange={(range) => {
        dateRange = range;
      }}
      onClear={() => {
        dateRange = { start: null, end: null };
      }}
      locale={getLocale()}
      class="min-w-60"
    />
    <button
      onclick={handleFilter}
      class="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-primary-hover"
    >
      <Search size={14} />
      {m.admin_search()}
    </button>
    <span class="ml-auto text-sm text-ink-4"
      >{m.total_items({ total: String(total) })}</span
    >
  </div>

  <!-- Logs table -->
  <div class="overflow-hidden rounded-xl border border-line bg-surface">
    <table class="w-full text-left text-sm">
      <colgroup>
        <col style="width: 60px" />
        <col style="width: 130px" />
        <col style="width: 140px" />
        <col style="width: 100px" />
        <col style="width: 100px" />
        <col style="width: 80px" />
        <col style="width: 80px" />
        <col style="width: 150px" />
        <col style="width: 70px" />
      </colgroup>
      <thead class="border-b border-line bg-surface-sunken text-xs text-ink-3">
        <tr>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_id()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_user()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_action()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_resource()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_ip()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_os()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_browser()}</th>
          <th class="px-4 py-3 font-medium">{m.admin_logs_col_time()}</th>
          <th class="px-4 py-3 font-medium"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-line-soft">
        {#if loading}
          <tr>
            <td colspan="9" class="px-4 py-12 text-center text-ink-4">
              <LoaderCircle size={20} class="mx-auto animate-spin" />
            </td>
          </tr>
        {:else if logs.length === 0}
          <tr>
            <td colspan="9" class="px-4 py-12 text-center text-ink-4"
              >{m.admin_logs_no_logs()}</td
            >
          </tr>
        {:else}
          {#each logs as log (log.id)}
            <tr class="transition-colors hover:bg-surface-sunken">
              <td class="px-4 py-3 text-xs text-ink-4">{log.id}</td>
              <td class="px-4 py-3">
                <button
                  onclick={() => goto(`/admin/users/${log.userId}`)}
                  class="flex items-center gap-1.5 rounded px-1 -mx-1 py-0.5 text-ink transition-colors hover:text-primary"
                >
                  <span class="text-xs text-ink-4">#{log.userId}</span>
                  <span class="font-medium truncate max-w-[80px]">{log.username || '—'}</span>
                  <ExternalLink size={12} class="shrink-0 text-ink-4 opacity-0 group-hover:opacity-100" />
                </button>
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-block max-w-[180px] truncate rounded bg-primary-soft px-2 py-0.5 text-xs font-medium text-primary"
                  title={log.action}
                >{log.actionLabel}</span>
              </td>
              <td class="truncate px-4 py-3 text-xs text-ink-3 max-w-[100px]" title={log.resourceName || log.resourceType || ''}>
                {log.resourceName || log.resourceType || '—'}
              </td>
              <td class="px-4 py-3 text-xs text-ink-3 font-mono">{log.ip || '—'}</td>
              <td class="px-4 py-3 text-xs text-ink-4">{log.os || '—'}</td>
              <td class="px-4 py-3 text-xs text-ink-4">{log.browser || '—'}</td>
              <td class="px-4 py-3 text-xs text-ink-4 whitespace-nowrap"
                >{fmtTime(log.createdAt)}</td
              >
              <td class="px-4 py-3">
                <button
                  onclick={() => openDetail(log)}
                  class="flex items-center gap-1 rounded px-1.5 py-1 text-xs text-ink-4 transition-colors hover:bg-primary-soft hover:text-primary"
                >
                  <Eye size={14} />
                </button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if totalPages > 1}
    <div class="flex items-center justify-center gap-2">
      <button
        class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-2 transition-colors hover:border-ink-2 hover:bg-surface hover:text-ink disabled:opacity-40"
        disabled={currentPage <= 1}
        onclick={() => goPage(currentPage - 1)}
      >
        <ChevronLeft size={14} />
      </button>
      <span class="text-sm text-ink-4">{currentPage} / {totalPages}</span>
      <button
        class="rounded-lg border border-line px-3 py-1.5 text-sm text-ink-2 transition-colors hover:border-ink-2 hover:bg-surface hover:text-ink disabled:opacity-40"
        disabled={currentPage >= totalPages}
        onclick={() => goPage(currentPage + 1)}
      >
        <ChevronRight size={14} />
      </button>
    </div>
  {/if}
</div>

<!-- Detail Dialog -->
<Dialog
  bind:open={detailOpen}
  title="Log Detail"
  size="md"
  footer={false}
  closable={true}
>
  {#if detailLog}
    <div class="space-y-4">
      <div class="grid grid-cols-[100px_1fr] gap-x-4 gap-y-2 text-sm">
        <span class="text-ink-4">ID</span>
        <span class="text-ink font-mono">{detailLog.id}</span>

        <span class="text-ink-4">{m.admin_logs_col_user()}</span>
        <span class="text-ink">
          <button
            onclick={handleDetailUserClick}
            class="inline-flex items-center gap-1 text-primary hover:underline"
          >
            #{detailLog.userId} {detailLog.username || ''}
            <ExternalLink size={12} />
          </button>
        </span>

        <span class="text-ink-4">{m.admin_logs_col_action()}</span>
        <span class="text-ink font-medium text-sm">{detailLog.actionLabel}</span>
        <span class="col-start-2 text-ink-4 text-xs font-mono">{detailLog.action}</span>

        <span class="text-ink-4">Resource Type</span>
        <span class="text-ink-3">{detailLog.resourceType || '—'}</span>

        <span class="text-ink-4">Resource Name</span>
        <span class="text-ink-3 break-all">{detailLog.resourceName || '—'}</span>

        <span class="text-ink-4">{m.admin_logs_col_ip()}</span>
        <span class="text-ink font-mono text-xs">{detailLog.ip || '—'}</span>

        <span class="text-ink-4">IP Region</span>
        <span class="text-ink-3">{detailLog.ipRegion || '—'}</span>

        <span class="text-ink-4">User Agent</span>
        <span class="text-ink-3 text-xs break-all">{detailLog.userAgent || '—'}</span>

        <span class="text-ink-4">{m.admin_logs_col_os()}</span>
        <span class="text-ink-3">{detailLog.os || '—'}</span>

        <span class="text-ink-4">{m.admin_logs_col_browser()}</span>
        <span class="text-ink-3">{detailLog.browser || '—'}</span>

        <span class="text-ink-4">{m.admin_logs_col_time()}</span>
        <span class="text-ink-3">{fmtTime(detailLog.createdAt)}</span>
      </div>

      {#if detailLog.extra}
        <div>
          <h3 class="mb-2 text-sm font-medium text-ink">Extra</h3>
          <pre class="max-h-64 overflow-auto rounded-lg bg-surface-muted p-3 text-xs text-ink-2 font-mono whitespace-pre-wrap break-all">{fmtExtra(detailLog)}</pre>
        </div>
      {/if}
    </div>
  {/if}
</Dialog>
