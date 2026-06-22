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
	class="inline-flex items-center outline-none {className}"
	{disabled}
	{...restProps}
>
	{@render children()}
</DropdownMenu.Trigger>
