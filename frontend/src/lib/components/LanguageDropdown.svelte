<script lang="ts">
	import { ChevronDown, Globe } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import { getLocale, locales, setLocale } from '$lib/paraglide/runtime';

	let {
		triggerClass = 'flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink-2 data-[state=open]:bg-surface-sunken data-[state=open]:text-ink-2',
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
		<ChevronDown size={10} class="text-ink-4" />
	{/snippet}

	{#each locales as locale}
		<DropdownBase.Item
			class={locale === getLocale() ? 'bg-primary-soft text-primary font-medium' : ''}
			onSelect={() => setLocale(locale)}
		>
			{localeLabels[locale] ?? locale}
		</DropdownBase.Item>
	{/each}
</Dropdown>
