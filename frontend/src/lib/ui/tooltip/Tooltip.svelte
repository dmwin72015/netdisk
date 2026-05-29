<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as TooltipBase from './base';

	let {
		open = $bindable(false),
		onOpenChange,
		content,
		children,
		side = 'top',
		sideOffset = 4,
		delayDuration = 700,
		disableHoverableContent = false,
		disabled = false,
		triggerClass = '',
		contentClass = '',
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		content: string;
		children: Snippet;
		side?: 'top' | 'bottom' | 'left' | 'right';
		sideOffset?: number;
		delayDuration?: number;
		disableHoverableContent?: boolean;
		disabled?: boolean;
		triggerClass?: string;
		contentClass?: string;
	} = $props();
</script>

<TooltipBase.Root bind:open {onOpenChange} {delayDuration} {disableHoverableContent} {disabled}>
	<TooltipBase.Trigger class={triggerClass}>
		{@render children()}
	</TooltipBase.Trigger>

	<TooltipBase.Content class={contentClass} {side} {sideOffset}>
		{content}
	</TooltipBase.Content>
</TooltipBase.Root>
