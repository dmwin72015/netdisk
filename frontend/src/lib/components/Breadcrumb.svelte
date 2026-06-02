<script lang="ts">
	import { ChevronRight, House } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	export type Crumb = {
		id: string;
		name: string;
	};

	let {
		items,
		showHome = true,
		collapseThreshold = 4,
		onNavigate,
		onHome
	}: {
		items: Crumb[];
		showHome?: boolean;
		collapseThreshold?: number;
		onNavigate: (id: string) => void;
		onHome?: () => void;
	} = $props();

	let expanded = $state(false);

	export function collapse() {
		expanded = false;
	}

	let needsCollapse = $derived(items.length > collapseThreshold);

	function handleClick(id: string) {
		expanded = false;
		onNavigate(id);
	}
</script>

{#if items.length > 0 || showHome}
<div class="flex items-center gap-1 overflow-hidden text-sm">
	{#if showHome}
		<button
			type="button"
			onclick={() => { expanded = false; onHome?.(); }}
			class="shrink-0 rounded p-1 text-gray-500 transition-colors hover:text-gray-900"
			title={m.all_files()}
		>
			<House size={16} />
		</button>
	{/if}

	{#each items as crumb, i}
		{#if !expanded && needsCollapse && i > 0 && i < items.length - 1}
			{#if i === 1}
				<ChevronRight size={14} class="shrink-0 text-gray-300" />
				<button
					type="button"
					onclick={() => (expanded = true)}
					class="shrink-0 rounded px-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
					title={m.show_full_path()}
				>...</button>
			{/if}
		{:else}
			{#if i > 0 || showHome}
				<ChevronRight size={14} class="shrink-0 text-gray-300" />
			{/if}
			{#if i === items.length - 1}
				<span class="max-w-48 truncate font-medium text-gray-900 sm:max-w-64 md:max-w-80" title={crumb.name}>{crumb.name}</span>
			{:else}
				<button
					type="button"
					onclick={() => handleClick(crumb.id)}
					class="max-w-32 shrink truncate rounded px-1 text-gray-500 transition-colors hover:text-gray-900 sm:max-w-40"
					title={crumb.name}
				>{crumb.name}</button>
			{/if}
		{/if}
	{/each}
</div>
{/if}
