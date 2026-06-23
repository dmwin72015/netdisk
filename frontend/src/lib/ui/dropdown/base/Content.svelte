<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
		sideOffset = 4,
		align = 'start',
		preventScroll = false,
		...restProps
	}: {
		children: Snippet;
		class?: string;
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
		preventScroll?: boolean;
	} = $props();
</script>

<DropdownMenu.Portal>
	<DropdownMenu.Content
		{sideOffset}
		{align}
		{preventScroll}
		class={cn(
			'bg-surface text-ink border-line shadow-pop z-50 min-w-40 rounded-lg border p-1.5',
			'data-[state=open]:animate-in data-[state=closed]:animate-out',
			'data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0',
			'data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95',
			'duration-150',
			'[&[data-side=bottom]]:origin-top [&[data-side=top]]:origin-bottom',
			'[&[data-side=left]]:origin-right [&[data-side=right]]:origin-left',
			className,
		)}
		{...restProps}
	>
		{@render children()}
	</DropdownMenu.Content>
</DropdownMenu.Portal>
