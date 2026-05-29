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
	let hoveredCategory = $state<string | null>(null);

	let percent = $derived(quotaBytes > 0 ? Math.min((usedBytes / quotaBytes) * 100, 100) : 0);

	const categoryColors: Record<string, { bg: string; text: string; label: string }> = {
		video:    { bg: 'bg-purple-500', text: 'text-purple-600', label: '视频' },
		audio:    { bg: 'bg-blue-500',   text: 'text-blue-600',   label: '音频' },
		image:    { bg: 'bg-pink-500',   text: 'text-pink-600',   label: '图片' },
		document: { bg: 'bg-amber-500',  text: 'text-amber-600',  label: '文档' },
		archive:  { bg: 'bg-gray-500',   text: 'text-gray-600',   label: '压缩包' },
		other:    { bg: 'bg-gray-400',   text: 'text-gray-500',   label: '其他' },
	};

	function getCategoryInfo(cat: string) {
		return categoryColors[cat] || categoryColors.other;
	}

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

<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
	<h2 class="mb-4 flex items-center gap-2 text-sm font-medium text-gray-500">
		<HardDrive size={16} /> {m.drive_storage()}
	</h2>
	<div class="space-y-3">
		{#if loading}
			<p class="text-sm text-gray-400">{m.loading()}</p>
		{:else}
			<div>
				<div class="mb-2 flex items-baseline justify-between">
					<span class="text-2xl font-semibold text-gray-900">{fmtSize(usedBytes)}</span>
					<span class="text-sm text-gray-400">/ {fmtSize(quotaBytes)}</span>
				</div>

				<!-- Multi-segment progress bar -->
				<div class="relative h-3 w-full overflow-hidden rounded-full bg-gray-100">
					{#if categories.length > 0}
						{#each categories as cat, i (cat.category)}
							{@const info = getCategoryInfo(cat.category)}
							{@const width = quotaBytes > 0 ? (cat.bytes / quotaBytes) * 100 : 0}
							{@const offset = categories.slice(0, i).reduce((sum, c) => sum + (quotaBytes > 0 ? (c.bytes / quotaBytes) * 100 : 0), 0)}
							<div
								class="absolute top-0 h-full transition-all duration-300 {info.bg} {hoveredCategory === cat.category ? 'opacity-100 brightness-110' : hoveredCategory ? 'opacity-60' : 'opacity-100'}"
								style="left:{offset}%; width:{width}%"
								onmouseenter={() => hoveredCategory = cat.category}
								onmouseleave={() => hoveredCategory = null}
								role="presentation"
							></div>
						{/each}
					{:else}
						<div
							class="h-full rounded-full transition-all {percent > 90 ? 'bg-red-500' : percent > 70 ? 'bg-amber-500' : 'bg-blue-600'}"
							style="width:{percent}%"
						></div>
					{/if}
				</div>

				<p class="mt-1.5 text-right text-xs text-gray-400">{percent.toFixed(1)}%</p>
			</div>

			<!-- Category legend -->
			{#if categories.length > 0 && !loadingBreakdown}
				<div class="flex flex-wrap gap-x-4 gap-y-2">
					{#each categories as cat (cat.category)}
						{@const info = getCategoryInfo(cat.category)}
						<div
							class="group relative flex items-center gap-1.5 text-xs {info.text} cursor-default"
							onmouseenter={() => hoveredCategory = cat.category}
							onmouseleave={() => hoveredCategory = null}
							role="presentation"
						>
							<span class="inline-block h-2.5 w-2.5 rounded-sm {info.bg}"></span>
							<span>{info.label}</span>
							<span class="text-gray-400">{fmtSize(cat.bytes)}</span>

							<!-- Tooltip -->
							<div class="pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 -translate-x-1/2 whitespace-nowrap rounded-lg bg-gray-900 px-3 py-2 text-xs text-white opacity-0 shadow-lg transition-opacity group-hover:opacity-100">
								<p class="font-medium">{info.label}</p>
								<p>{fmtSize(cat.bytes)} · {cat.count} 个文件</p>
								<div class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
							</div>
						</div>
					{/each}
				</div>
			{/if}

			<p class="text-xs text-gray-400">{m.used()}: {fmtSize(usedBytes)} &middot; {m.drive_storage()}: {fmtSize(quotaBytes)}</p>
		{/if}
	</div>
</div>
