<script lang="ts">
	import { Popover } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		onOpenChange,
		children,
		...restProps
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		children: Snippet;
	} = $props();
</script>

<Popover.Root bind:open {onOpenChange} {...restProps}>
	{@render children()}
</Popover.Root>
