<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { LoaderCircle, Users, FileText, HardDrive, UserPlus, FileUp, Monitor } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { adminDashboardStats, type AdminDashboardStats } from '$lib/api/admin';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let stats = $state<AdminDashboardStats | null>(null);
	let loading = $state(true);

	onMount(() => {
		if (!browser) return;
		loadStats();
	});

	async function loadStats() {
		loading = true;
		try {
			stats = await adminDashboardStats();
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}

	type StatCard = {
		label: string;
		value: string;
		icon: typeof Users;
		color: string;
		bg: string;
	};

	const cards = $derived<StatCard[]>(stats ? [
		{
			label: m.admin_total_users(),
			value: String(stats.totalUsers),
			icon: Users,
			color: 'text-primary',
			bg: 'bg-primary-soft',
		},
		{
			label: m.admin_total_files(),
			value: String(stats.totalFiles),
			icon: FileText,
			color: 'text-success',
			bg: 'bg-success-soft',
		},
		{
			label: m.admin_storage_used(),
			value: fmtSize(stats.storageUsed),
			icon: HardDrive,
			color: 'text-info',
			bg: 'bg-info-soft',
		},
		{
			label: m.admin_total_quota(),
			value: fmtSize(stats.totalStorage),
			icon: HardDrive,
			color: 'text-admin',
			bg: 'bg-warning-soft',
		},
		{
			label: m.admin_new_users_today(),
			value: String(stats.newTodayUsers),
			icon: UserPlus,
			color: 'text-danger',
			bg: 'bg-danger-soft',
		},
		{
			label: m.admin_new_files_today(),
			value: String(stats.newTodayFiles),
			icon: FileUp,
			color: 'text-ink-2',
			bg: 'bg-surface-sunken',
		},
	] : []);
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-ink">{m.admin_dashboard()}</h1>
		<p class="mt-1 text-sm text-ink-4">{m.admin_dashboard_desc()}</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<LoaderCircle size={24} class="animate-spin text-ink-4" />
		</div>
	{:else if cards.length > 0}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
			{#each cards as card}
				{@const Icon = card.icon}
				<div class="rounded-xl border border-line bg-surface p-4 transition-colors hover:border-ink-5">
					<div class="rounded-lg {card.bg} inline-flex p-2">
						<Icon size={20} class={card.color} />
					</div>
					<p class="mt-3 text-2xl font-bold text-ink">{card.value}</p>
					<p class="mt-0.5 text-xs text-ink-4">{card.label}</p>
				</div>
			{/each}
		</div>

		{@const usagePct = stats && stats.totalStorage > 0 ? Math.round((stats.storageUsed / stats.totalStorage) * 100) : 0}
		{@const usedStr = fmtSize(stats?.storageUsed ?? 0)}
		{@const totalStr = fmtSize(stats?.totalStorage ?? 0)}
		<div class="rounded-xl border border-line bg-surface p-5">
			<h2 class="mb-3 text-sm font-semibold text-ink-2">{m.admin_storage_usage()}</h2>
			<div class="mb-2 flex items-center justify-between text-sm">
				<span class="text-ink-3">{m.admin_used_of({ used: usedStr, total: totalStr, pct: String(usagePct) })}</span>
			</div>
			<div class="h-2.5 overflow-hidden rounded-full bg-surface-sunken">
				<div
					class="h-full rounded-full bg-gradient-to-r from-primary to-info transition-all duration-500"
					style="width: {usagePct}%"
				></div>
			</div>
		</div>

		{#if stats.diskTotal > 0}
			{@const diskPct = Math.round((stats.diskUsed / stats.diskTotal) * 100)}
			{@const diskUsedStr = fmtSize(stats.diskUsed)}
			{@const diskFreeStr = fmtSize(stats.diskFree)}
			<div class="rounded-xl border border-line bg-surface p-5">
				<div class="mb-3 flex items-center gap-2">
					<div class="rounded-lg bg-surface-sunken p-1.5">
						<Monitor size={16} class="text-ink-2" />
					</div>
					<h2 class="text-sm font-semibold text-ink-2">System Disk</h2>
				</div>
				<div class="mb-2 flex items-center justify-between text-sm">
					<span class="text-ink-3">{diskUsedStr} / {fmtSize(stats.diskTotal)}</span>
					<span class="text-xs text-ink-4">{diskFreeStr} free</span>
				</div>
				<div class="h-2.5 overflow-hidden rounded-full bg-surface-sunken">
					<div
						class="h-full rounded-full transition-all {diskPct > 90 ? 'bg-danger' : diskPct > 70 ? 'bg-warning' : 'bg-info'}"
						style="width: {diskPct}%"
					></div>
				</div>
			</div>
		{/if}
	{/if}
</div>
