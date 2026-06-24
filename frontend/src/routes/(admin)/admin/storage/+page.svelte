<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { LoaderCircle, HardDrive, Trash2, Film, Music, Image, FileText, Archive, FileQuestion } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { adminDashboardStats, adminStorageStats, type AdminDashboardStats, type CategoryStat } from '$lib/api/admin';
	import { fmtSize } from '$lib/utils/format';
	import * as m from '$lib/paraglide/messages';

	let stats = $state<AdminDashboardStats | null>(null);
	let catStats = $state<CategoryStat[]>([]);
	let loading = $state(true);

	const categoryMeta: Record<string, { label: string; icon: typeof HardDrive; color: string }> = {
		video: { label: m.category_video(), icon: Film, color: 'text-danger' },
		audio: { label: m.category_audio(), icon: Music, color: 'text-success' },
		image: { label: m.category_image(), icon: Image, color: 'text-info' },
		document: { label: m.category_document(), icon: FileText, color: 'text-primary' },
		archive: { label: m.category_archive(), icon: Archive, color: 'text-warning' },
		other: { label: m.category_other(), icon: FileQuestion, color: 'text-ink-3' },
		trash: { label: m.category_trash(), icon: Trash2, color: 'text-danger' },
	};

	const sortedCatStats = $derived(
		catStats.filter((c) => c.category !== 'trash').sort((a, b) => b.bytes - a.bytes)
	);

	const trashStat = $derived(catStats.find((c) => c.category === 'trash'));

	onMount(() => {
		if (!browser) return;
		loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [s, c] = await Promise.all([adminDashboardStats(), adminStorageStats()]);
			stats = s;
			catStats = c;
		} catch {
			toast.error(m.admin_load_failed());
		} finally {
			loading = false;
		}
	}

	const pctUsed = $derived(stats && stats.totalStorage > 0 ? Math.round((stats.storageUsed / stats.totalStorage) * 100) : 0);
	const usedStr = $derived(stats ? fmtSize(stats.storageUsed) : '');
	const totalStr = $derived(stats ? fmtSize(stats.totalStorage) : '');

	function pct(bytes: number): number {
		if (!stats || stats.totalStorage === 0) return 0;
		return Math.round((bytes / stats.totalStorage) * 100);
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-ink">{m.admin_storage()}</h1>
		<p class="mt-1 text-sm text-ink-4">{m.admin_storage_breakdown_desc()}</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-16">
			<LoaderCircle size={24} class="animate-spin text-ink-4" />
		</div>
	{:else if stats}
		<!-- Overall storage -->
		<div class="rounded-xl border border-line bg-surface p-5">
			<h2 class="mb-4 text-sm font-semibold text-ink-2">{m.admin_storage_usage()}</h2>
			<div class="mb-2 flex items-center justify-between text-sm">
				<span class="text-ink-3">{m.admin_used_of({ used: usedStr, total: totalStr, pct: String(pctUsed) })}</span>
				<span class="text-xs text-ink-4">{m.admin_total_files()}: {stats.totalFiles}</span>
			</div>
			<div class="h-3 overflow-hidden rounded-full bg-surface-sunken">
				<div
					class="h-full rounded-full bg-gradient-to-r from-primary to-info transition-all duration-500"
					style="width: {pctUsed}%"
				></div>
			</div>
		</div>

		<!-- Category breakdown -->
		<div class="rounded-xl border border-line bg-surface p-5">
			<h2 class="mb-4 text-sm font-semibold text-ink-2">{m.admin_by_category()}</h2>
			<div class="space-y-3">
				{#each sortedCatStats as cat}
					{@const meta = categoryMeta[cat.category] || { label: cat.category, icon: FileQuestion, color: 'text-ink-3' }}
					{@const Icon = meta.icon}
					<div class="flex items-center gap-4">
						<div class="rounded-lg bg-surface-sunken p-2">
							<Icon size={18} class={meta.color} />
						</div>
						<div class="flex-1 min-w-0">
							<div class="flex items-center justify-between text-sm">
								<span class="font-medium text-ink">{meta.label}</span>
								<span class="text-xs text-ink-4">{cat.count} files</span>
							</div>
							<div class="mt-1 flex items-center justify-between text-xs text-ink-4">
								<span>{fmtSize(cat.bytes)}</span>
								<span>{pct(cat.bytes)}%</span>
							</div>
							<div class="mt-1 h-1.5 overflow-hidden rounded-full bg-surface-sunken">
								<div class="h-full rounded-full bg-primary" style="width: {pct(cat.bytes)}%"></div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Trash section -->
		{#if trashStat}
			<div class="rounded-xl border border-line bg-surface p-5">
				<div class="flex items-center gap-3">
					<div class="rounded-lg bg-danger-soft p-2">
						<Trash2 size={18} class="text-danger" />
					</div>
					<div class="flex-1">
						<div class="flex items-center justify-between text-sm">
							<span class="font-medium text-ink">{m.admin_trashed_files()}</span>
							<span class="text-xs text-ink-4">{trashStat.count} files</span>
						</div>
						<p class="text-xs text-ink-4">{fmtSize(trashStat.bytes)}</p>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
