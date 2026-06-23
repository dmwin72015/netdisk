<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

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
			? 'text-danger hover:bg-danger-soft focus:bg-danger-soft'
			: 'text-ink-2 hover:bg-surface-sunken focus:bg-surface-sunken hover:text-ink focus:text-ink'
	);
</script>

<DropdownMenu.Item
	class={cn(
		'flex cursor-pointer items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm outline-none transition-colors duration-150 select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
		variantClasses,
		className,
	)}
	{disabled}
	{onSelect}
>
	{#if icon}
		<span class="shrink-0 text-ink-4">
			{@render icon()}
		</span>
	{/if}
	{@render children()}
	{#if shortcut}
		<span class="ml-auto text-xs tracking-widest text-ink-4">{shortcut}</span>
	{/if}
</DropdownMenu.Item>
