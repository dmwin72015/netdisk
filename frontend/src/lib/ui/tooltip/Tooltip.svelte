<script lang="ts">
	import type { Snippet } from 'svelte';
	import { type Tooltip as TooltipTypes } from 'bits-ui';
	import * as TooltipBase from './base';

	type TriggerProps = Omit<TooltipTypes.TriggerProps, 'class' | 'disabled' | 'children'>;

	let {
		open = $bindable(false),
		onOpenChange,
		content,
		children,
		side = 'top',
		sideOffset = 4,
		delayDuration = 700,
		skipDelayDuration = 300,
		disableHoverableContent = false,
		disabled = false,
		triggerClass = '',
		contentClass = '',
		triggerProps = {},
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		content: string;
		children: Snippet;
		side?: 'top' | 'bottom' | 'left' | 'right';
		sideOffset?: number;
		delayDuration?: number;
		skipDelayDuration?: number;
		disableHoverableContent?: boolean;
		disabled?: boolean;
		triggerClass?: string;
		contentClass?: string;
		triggerProps?: TriggerProps;
	} = $props();
</script>

<TooltipBase.Provider {delayDuration} {skipDelayDuration} {disableHoverableContent} {disabled}>
	<TooltipBase.Root bind:open {onOpenChange} {delayDuration} {disableHoverableContent} {disabled}>
		<TooltipBase.Trigger class={triggerClass} {...triggerProps}>
			{@render children()}
		</TooltipBase.Trigger>

		<TooltipBase.Content class={contentClass} {side} {sideOffset}>
			{content}
		</TooltipBase.Content>
	</TooltipBase.Root>
</TooltipBase.Provider>
