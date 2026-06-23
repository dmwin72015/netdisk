<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<DropdownMenu.GroupHeading class={cn('px-3 py-1.5 text-xs font-medium text-ink-3', className)}>
	{@render children()}
</DropdownMenu.GroupHeading>
