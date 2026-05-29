<script lang="ts">
	import { Dialog } from 'bits-ui';
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

<Dialog.Root bind:open {onOpenChange} {...restProps}>
	{@render children()}
</Dialog.Root>
