<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		variant = 'default',
		icon,
		shortcut,
		disabled = false,
		onSelect,
	}: {
		children: Snippet;
		class?: string;
		variant?: 'default' | 'destructive';
		icon?: Snippet;
		shortcut?: string;
		disabled?: boolean;
		onSelect?: (event: Event) => void;
	} = $props();

	const variantClasses = $derived(
		variant === 'destructive'
			? 'text-red-600 hover:bg-red-50 focus:bg-red-50'
			: 'text-gray-700 hover:bg-gray-50 focus:bg-gray-50'
	);
</script>

<DropdownMenu.Item
	class="flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm outline-none transition-colors select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 {variantClasses} {className}"
	{disabled}
	{onSelect}
>
	{#if icon}
		<span class="shrink-0 text-gray-400">
			{@render icon()}
		</span>
	{/if}
	{@render children()}
	{#if shortcut}
		<span class="ml-auto text-xs tracking-widest text-gray-400">{shortcut}</span>
	{/if}
</DropdownMenu.Item>
