<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as DropdownBase from './base';
	import { cn } from '$lib/utils/cn';

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
		disabled = false,
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
		disabled?: boolean;
	} = $props();
</script>

<DropdownBase.Root bind:open {onOpenChange}>
	<DropdownBase.Trigger class={cn(triggerClass)} {disabled}>
		{@render trigger()}
	</DropdownBase.Trigger>

	<DropdownBase.Content class={cn(contentClass)} {sideOffset} {align} {preventScroll}>
		{@render children()}
	</DropdownBase.Content>
</DropdownBase.Root>
