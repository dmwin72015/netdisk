<script lang="ts">
	import { Popover } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		style = '',
		side = 'bottom',
		sideOffset = 4,
		align = 'center',
		...restProps
	}: {
		children: Snippet;
		class?: string;
		style?: string;
		side?: 'top' | 'right' | 'bottom' | 'left';
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
	} = $props();
</script>

<Popover.Portal>
	<Popover.Content
		{side}
		{sideOffset}
		{align}
		class="bg-surface text-ink border-line shadow-pop z-50 min-w-40 rounded-lg border
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2
			data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2
			duration-150
			{className}"
		{style}
		{...restProps}
	>
		{@render children()}
	</Popover.Content>
</Popover.Portal>
