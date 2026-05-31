<script lang="ts">
	import { ChevronDown, Globe } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { getLocale, locales, setLocale } from '$lib/paraglide/runtime';

	let {
		triggerClass = 'flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 data-[state=open]:bg-gray-100 data-[state=open]:text-gray-700',
		contentClass = 'min-w-[120px]',
		sideOffset = 8,
		align = 'end',
	}: {
		triggerClass?: string;
		contentClass?: string;
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
	} = $props();

	const localeLabels: Record<string, string> = { zh: '中文', en: 'English' };
	let open = $state(false);
</script>

<Dropdown bind:open {triggerClass} {contentClass} {sideOffset} {align}>
	{#snippet trigger()}
		<Globe size={14} />
		<span>{localeLabels[getLocale()] ?? getLocale()}</span>
		<ChevronDown size={10} class="text-gray-400" />
	{/snippet}

	{#each locales as locale}
		<DropdownBase.Item
			class={locale === getLocale() ? 'bg-blue-50 text-blue-600 font-medium' : ''}
			onSelect={() => setLocale(locale)}
		>
			{localeLabels[locale] ?? locale}
		</DropdownBase.Item>
	{/each}
</Dropdown>
