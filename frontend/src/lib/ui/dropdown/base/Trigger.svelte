<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
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

<DropdownMenu.Trigger
	class="inline-flex items-center justify-center outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 {className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</DropdownMenu.Trigger>
