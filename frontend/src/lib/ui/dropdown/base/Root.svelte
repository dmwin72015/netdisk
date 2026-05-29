<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
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

<DropdownMenu.Root bind:open {onOpenChange} {...restProps}>
	{@render children()}
</DropdownMenu.Root>
