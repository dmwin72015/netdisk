<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		side = 'top' as 'top' | 'bottom' | 'left' | 'right',
		sideOffset = 4,
		...restProps
	}: {
		children: Snippet;
		class?: string;
		side?: 'top' | 'bottom' | 'left' | 'right';
		sideOffset?: number;
	} = $props();
</script>

<Tooltip.Portal>
	<Tooltip.Content
		{side}
		{sideOffset}
		class="z-50 rounded-md bg-gray-900 px-3 py-1.5 text-sm text-white shadow-md
			data-[state=open]:animate-in data-[state=closed]:animate-out
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0
			data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2
			data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2
			{className}"
		{...restProps}
	>
		{@render children()}
		<Tooltip.Arrow class="fill-gray-900" />
	</Tooltip.Content>
</Tooltip.Portal>
