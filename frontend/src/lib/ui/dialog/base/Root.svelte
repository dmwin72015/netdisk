<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		onOpenChange,
		onOpenChangeComplete,
		children,
		...restProps
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		onOpenChangeComplete?: (open: boolean) => void;
		children: Snippet;
	} = $props();
	</script>

	<Dialog.Root bind:open {onOpenChange} {onOpenChangeComplete} {...restProps}>
		{@render children()}
	</Dialog.Root>
