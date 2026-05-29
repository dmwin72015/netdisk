<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<DropdownMenu.GroupHeading class="px-3 py-1.5 text-xs font-medium text-gray-500 {className}">
	{@render children()}
</DropdownMenu.GroupHeading>
