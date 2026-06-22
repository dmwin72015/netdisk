<script lang="ts">
	import { onMount } from 'svelte';
	import { HardDrive } from '@lucide/svelte';
	import { fmtSize } from '$lib/utils/format';
	import { getStorageBreakdown, type CategoryStat } from '$lib/api/profile';
	import * as m from '$lib/paraglide/messages';

	let {
		usedBytes,
		quotaBytes,
		loading = false
	}: {
		usedBytes: number;
		quotaBytes: number;
		loading?: boolean;
	} = $props();

	let categories = $state<CategoryStat[]>([]);
	let loadingBreakdown = $state(false);
	let percent = $derived(quotaBytes > 0 ? Math.min((usedBytes / quotaBytes) * 100, 100) : 0);
	let categoryBaseBytes = $derived(Math.max(usedBytes, 0));

	const categoryColors: Record<string, string> = {
		video: '#8b5cf6',
		audio: '#3b82f6',
		image: '#ec4899',
		document: '#f59e0b',
		archive: '#6b7280',
		other: '#9ca3af',
		trash: '#dc2626',
	};

	const categoryLabels: Record<string, string> = {
		video: m.category_video(),
		audio: m.category_audio(),
		image: m.category_image(),
		document: m.category_document(),
		archive: m.category_archive(),
		other: m.category_other(),
		trash: m.category_trash(),
	};

	function getColor(cat: string) {
		return categoryColors[cat] || categoryColors.other;
	}

	function getLabel(cat: string) {
		return categoryLabels[cat] || categoryLabels.other;
	}

	let barSegments = $derived.by(() => {
		if (categories.length === 0 || categoryBaseBytes <= 0) return [];
		return categories
			.filter((cat) => cat.bytes > 0)
			.map((cat) => ({
				cat,
				width: Math.max((cat.bytes / categoryBaseBytes) * 100, 0.5),
				color: getColor(cat.category),
			}));
	});

	onMount(async () => {
		loadingBreakdown = true;
		try {
			categories = await getStorageBreakdown();
		} catch {
			// ignore
		} finally {
			loadingBreakdown = false;
		}
	});
</script>

<div class="rounded-xl border border-line-soft bg-white p-6 ">
	<h2 class="mb-5 flex items-center gap-2 text-sm font-medium text-ink-3">
		<HardDrive size={16} /> {m.drive_storage()}
	</h2>

	{#if loading}
		<p class="text-sm text-ink-4">{m.loading()}</p>
	{:else}
		<!-- Block 1: Overall usage (used of total) -->
		<div class="mb-3 text-sm text-ink-3">
			<span class="font-semibold text-ink">{fmtSize(usedBytes)}</span>
			<span class="text-ink-4">{m.used()}</span>
			<span class="mx-1 text-ink-4">/</span>
			<span class="font-semibold text-ink">{fmtSize(quotaBytes)}</span>
		</div>

		<!-- Overall usage bar -->
		<div class="h-2 w-full overflow-hidden rounded-full bg-surface-sunken">
			<div
				class="h-full rounded-full transition-all {percent > 90 ? 'bg-danger' : percent > 70 ? 'bg-warning' : 'bg-primary'}"
				style="width:{percent}%"
			></div>
		</div>

		<!-- Block 2: Category breakdown (GitHub-style) -->
		{#if categories.length > 0 && !loadingBreakdown}
			<div class="mt-6">
				<!-- GitHub-style thin segmented bar -->
				<div class="mb-3 h-2 w-full overflow-hidden rounded-full bg-surface-sunken">
					{#if barSegments.length > 0}
						<div class="flex h-full">
							{#each barSegments as seg (seg.cat.category)}
								<div
									style="background-color:{seg.color}; width:{seg.width}%"
									class="h-full first:rounded-l-full last:rounded-r-full"
								></div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- GitHub-style compact legend -->
				<div class="flex flex-wrap gap-x-5 gap-y-1.5">
					{#each categories as cat (cat.category)}
						{@const pct = categoryBaseBytes > 0 ? ((cat.bytes / categoryBaseBytes) * 100).toFixed(2) : '0.00'}
						<div class="group relative flex items-center gap-1.5 text-sm">
							<span
								class="inline-block h-2.5 w-2.5 rounded-full shrink-0"
								style="background-color:{getColor(cat.category)}"
							></span>
							<span class="cursor-default text-ink underline decoration-dotted decoration-line underline-offset-2">{getLabel(cat.category)}</span>
							<span class="text-ink-3">{pct}%</span>
							<!-- Tooltip -->
							<div class="pointer-events-none absolute -top-1 left-0 z-10 -translate-y-full whitespace-nowrap rounded-md bg-ink px-2 py-1 text-xs text-white opacity-0 shadow-pop transition-opacity group-hover:opacity-100">
								{m.total_items({ total: cat.count })}
								<div class="absolute left-3 top-full h-0 w-0 border-4 border-transparent border-t-gray-900"></div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
