<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		delayDuration = 700,
		skipDelayDuration = 300,
		disableHoverableContent = false,
		disableCloseOnTriggerClick = false,
		disabled = false,
		ignoreNonKeyboardFocus = false,
	}: {
		children: Snippet;
		delayDuration?: number;
		skipDelayDuration?: number;
		disableHoverableContent?: boolean;
		disableCloseOnTriggerClick?: boolean;
		disabled?: boolean;
		ignoreNonKeyboardFocus?: boolean;
	} = $props();
</script>

<Tooltip.Provider
	{delayDuration}
	{skipDelayDuration}
	{disableHoverableContent}
	{disableCloseOnTriggerClick}
	{disabled}
	{ignoreNonKeyboardFocus}
>
	{@render children()}
</Tooltip.Provider>
