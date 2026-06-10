<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';

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
		class="z-50 min-w-40 rounded-xl border border-gray-100 bg-white p-1.5 shadow-lg
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			[&[data-side=bottom]]:origin-top [&[data-side=top]]:origin-bottom
			[&[data-side=left]]:origin-right [&[data-side=right]]:origin-left
			{className}"
		{...restProps}
	>
		{@render children()}
	</DropdownMenu.Content>
</DropdownMenu.Portal>
