<script lang="ts">
  import { onMount } from "svelte";
  import { authReady, user } from "$lib/stores/auth";
  import { Monitor, Smartphone, LoaderCircle } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";
  import { toast } from "svelte-sonner";
  import { api } from "$lib/api/client";

  interface SecurityLogItem {
    id: number;
    action: string;
    ip: string | null;
    ipRegion: string | null;
    os: string | null;
    browser: string | null;
    createdAt: string | null;
  }

  interface LoginDevice {
    ip: string;
    ipRegion: string;
    os: string;
    browser: string;
    userAgent: string;
    lastLogin: string;
    isCurrent: boolean;
  }

  interface SecurityLogResponse {
    items: SecurityLogItem[];
    total: number;
  }

  let devices = $state<LoginDevice[]>([]);
  let logs = $state<SecurityLogItem[]>([]);
  let logsTotal = $state(0);
  let logsPage = $state(1);
  let loadingDevices = $state(true);
  let loadingLogs = $state(true);

  const PAGE_SIZE = 10;

  const actionLabels: Record<string, string> = {
    "user.login": "账户登录",
    "user.register": "账户注册",
    "user.logout": "账户登出",
    "user.oauth_login": "第三方登录",
    "user.password_change": "密码修改",
  };

  function isMobile(ua: string): boolean {
    return /mobile|android|iphone|ipad|ipod/i.test(ua);
  }

  function fmtTime(time: string | null): string {
    if (!time) return "-";
    const d = new Date(time);
    return d.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  async function fetchDevices() {
    loadingDevices = true;
    try {
      const data = await api<LoginDevice[]>("/api/v1/user/login-devices");
      devices = data;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.load_failed());
    } finally {
      loadingDevices = false;
    }
  }

  async function fetchLogs(page = 1) {
    loadingLogs = true;
    try {
      const data = await api<SecurityLogResponse>(
        `/api/v1/user/security-logs?page=${page}&pageSize=${PAGE_SIZE}`,
      );
      logs = data.items;
      logsTotal = data.total;
      logsPage = page;
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.load_failed());
    } finally {
      loadingLogs = false;
    }
  }

  let totalPages = $derived(Math.ceil(logsTotal / PAGE_SIZE));

  onMount(() => {
    void fetchDevices();
    void fetchLogs();
  });
</script>

{#if $authReady && $user}
  <div class="space-y-8">
    <!-- Login Devices -->
    <section>
      <h2 class="mb-1 text-sm font-semibold text-ink">{m.security_login_devices()}</h2>
      <p class="mb-4 text-xs text-ink-4">{m.security_login_devices_desc()}</p>

      {#if loadingDevices}
        <div class="flex items-center justify-center py-12">
          <LoaderCircle size={20} class="animate-spin text-ink-4" />
        </div>
      {:else if devices.length === 0}
        <p class="py-8 text-center text-sm text-ink-4">{m.no_data()}</p>
      {:else}
        <div class="space-y-2">
          {#each devices as device}
            <div class="flex items-center gap-3 rounded-lg border border-line-soft bg-white px-4 py-3">
              <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-surface-sunken text-ink-3">
                {#if isMobile(device.userAgent)}
                  <Smartphone size={18} />
                {:else}
                  <Monitor size={18} />
                {/if}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm text-ink-2">
                    {device.ipRegion || device.ip}
                  </span>
                  {#if device.isCurrent}
                    <span class="rounded-full bg-primary-soft px-2 py-0.5 text-[10px] font-medium text-primary">
                      {m.security_current_device()}
                    </span>
                  {/if}
                </div>
                <span class="text-xs text-ink-4">
                  {device.browser}{device.os ? ` on ${device.os}` : ""} · {m.security_login_time()} {device.lastLogin}
                </span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <!-- Security Records -->
    <section>
      <h2 class="mb-4 text-sm font-semibold text-ink">{m.security_records()}</h2>

      {#if loadingLogs}
        <div class="flex items-center justify-center py-12">
          <LoaderCircle size={20} class="animate-spin text-ink-4" />
        </div>
      {:else if logs.length === 0}
        <p class="py-8 text-center text-sm text-ink-4">{m.no_data()}</p>
      {:else}
        <div class="overflow-hidden rounded-lg border border-line-soft">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-line-soft bg-surface-sunken text-left text-xs text-ink-4">
                <th class="px-4 py-2.5 font-medium">{m.security_detail()}</th>
                <th class="px-4 py-2.5 font-medium">{m.security_ip()}</th>
                <th class="px-4 py-2.5 font-medium">{m.security_time()}</th>
              </tr>
            </thead>
            <tbody>
              {#each logs as log}
                <tr class="border-b border-line-soft last:border-0">
                  <td class="px-4 py-2.5 text-ink-2">
                    {actionLabels[log.action] || log.action}
                  </td>
                  <td class="px-4 py-2.5 text-ink-3">
                    {log.ip || "-"}
                  </td>
                  <td class="px-4 py-2.5 text-ink-3">
                    {fmtTime(log.createdAt)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        {#if totalPages > 1}
          <div class="mt-4 flex items-center justify-center gap-1">
            <button
              type="button"
              disabled={logsPage <= 1}
              onclick={() => fetchLogs(logsPage - 1)}
              class="rounded-md px-2.5 py-1 text-xs text-ink-3 transition-colors hover:bg-surface-sunken disabled:opacity-40"
            >
              {m.prev()}
            </button>
            {#each Array.from({ length: Math.min(totalPages, 7) }, (_, i) => i + 1) as p}
              <button
                type="button"
                onclick={() => fetchLogs(p)}
                class="h-7 min-w-7 rounded-md px-1.5 text-xs transition-colors {p === logsPage
                  ? 'bg-primary-soft font-medium text-primary'
                  : 'text-ink-3 hover:bg-surface-sunken'}"
              >
                {p}
              </button>
            {/each}
            {#if totalPages > 7}
              <span class="px-1 text-xs text-ink-4">...</span>
              <button
                type="button"
                onclick={() => fetchLogs(totalPages)}
                class="h-7 min-w-7 rounded-md px-1.5 text-xs text-ink-3 transition-colors hover:bg-surface-sunken"
              >
                {totalPages}
              </button>
            {/if}
            <button
              type="button"
              disabled={logsPage >= totalPages}
              onclick={() => fetchLogs(logsPage + 1)}
              class="rounded-md px-2.5 py-1 text-xs text-ink-3 transition-colors hover:bg-surface-sunken disabled:opacity-40"
            >
              {m.next()}
            </button>
          </div>
        {/if}
      {/if}
    </section>
  </div>
{/if}
