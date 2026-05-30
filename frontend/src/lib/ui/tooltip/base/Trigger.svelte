<script lang="ts">
	import { Tooltip, type Tooltip as TooltipTypes } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type TriggerProps = Omit<TooltipTypes.TriggerProps, 'class' | 'disabled' | 'children'>;

	let {
		children,
		class: className = '',
		disabled = false,
		type = 'button',
		...restProps
	}: {
		children: Snippet;
		class?: string;
		disabled?: boolean;
	} & TriggerProps = $props();
</script>

<Tooltip.Trigger
	{type}
	class="inline-flex items-center justify-center outline-none {className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</Tooltip.Trigger>
