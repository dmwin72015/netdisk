<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as PopoverBase from './base';
	import { cn } from '$lib/utils/cn';

	let {
		open = $bindable(false),
		onOpenChange,
		trigger,
		children,
		triggerClass = '',
		contentClass = '',
		contentStyle = '',
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
		contentStyle?: string;
		side?: 'top' | 'right' | 'bottom' | 'left';
		sideOffset?: number;
		align?: 'start' | 'center' | 'end';
	} = $props();
</script>

<PopoverBase.Root bind:open {onOpenChange}>
	<PopoverBase.Trigger class={cn(triggerClass)}>
		{@render trigger()}
	</PopoverBase.Trigger>

	<PopoverBase.Content class={cn(contentClass)} style={contentStyle} {side} {sideOffset} {align}>
		{@render children()}
	</PopoverBase.Content>
</PopoverBase.Root>
