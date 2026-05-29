<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as PopoverBase from './base';

	let {
		open = $bindable(false),
		onOpenChange,
		trigger,
		children,
		triggerClass = '',
		contentClass = '',
		side = 'bottom',
		sideOffset = 4,
		align = 'center',
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		trigger: Snippet;
		children: Snippet;
		triggerClass?: string;
		contentClass?: string;
		side?: 'top' | 'right' | 'bottom' | 'left';
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
	} = $props();
</script>

<PopoverBase.Root bind:open {onOpenChange}>
	<PopoverBase.Trigger class={triggerClass}>
		{@render trigger()}
	</PopoverBase.Trigger>

	<PopoverBase.Content class={contentClass} {side} {sideOffset} {align}>
		{@render children()}
	</PopoverBase.Content>
</PopoverBase.Root>
