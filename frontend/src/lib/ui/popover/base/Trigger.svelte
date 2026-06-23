<script lang="ts">
	import { Popover } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

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

<Popover.Trigger
	class={cn('inline-flex items-center justify-center outline-none', className)}
	{disabled}
	{...restProps}
>
	{@render children()}
</Popover.Trigger>
