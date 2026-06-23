<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as ContextMenuBase from './base';
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
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		trigger: Snippet;
		children: Snippet;
		triggerClass?: string;
		contentClass?: string;
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
	} = $props();
</script>

<ContextMenuBase.Root bind:open {onOpenChange}>
	<ContextMenuBase.Trigger class={cn(triggerClass)}>
		{@render trigger()}
	</ContextMenuBase.Trigger>

	<ContextMenuBase.Content class={cn(contentClass)} {sideOffset} {align}>
		{@render children()}
	</ContextMenuBase.Content>
</ContextMenuBase.Root>
