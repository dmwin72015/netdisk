<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		onOpenChange,
		delayDuration = 700,
		disableHoverableContent = false,
		disabled = false,
		children,
		...restProps
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		delayDuration?: number;
		disableHoverableContent?: boolean;
		disabled?: boolean;
		children: Snippet;
	} = $props();
</script>

<Tooltip.Root bind:open {onOpenChange} {delayDuration} {disableHoverableContent} {disabled} {...restProps}>
	{@render children()}
</Tooltip.Root>
