<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as DropdownBase from './base';

	let {
		open = $bindable(false),
		onOpenChange,
		trigger,
		children,
		triggerClass = '',
		contentClass = '',
		sideOffset = 4,
		align = 'start',
		preventScroll = false,
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		trigger: Snippet;
		children: Snippet;
		triggerClass?: string;
		contentClass?: string;
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
		preventScroll?: boolean;
	} = $props();
</script>

<DropdownBase.Root bind:open {onOpenChange}>
	<DropdownBase.Trigger class={triggerClass}>
		{@render trigger()}
	</DropdownBase.Trigger>

	<DropdownBase.Content class={contentClass} {sideOffset} {align} {preventScroll}>
		{@render children()}
	</DropdownBase.Content>
</DropdownBase.Root>
