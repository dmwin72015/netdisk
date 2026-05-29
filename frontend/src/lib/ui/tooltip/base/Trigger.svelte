<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
		disabled = false,
		...restProps
	}: {
		children: Snippet;
		class?: string;
		disabled?: boolean;
	} = $props();
</script>

<Tooltip.Trigger
	class="inline-flex items-center justify-center outline-none {className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</Tooltip.Trigger>
