<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
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

<DropdownMenu.Trigger
	class={cn('inline-flex items-center outline-none', className)}
	{disabled}
	{...restProps}
>
	{@render children()}
</DropdownMenu.Trigger>
