<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	let {
		ref = $bindable(null),
		children,
		class: className = '',
	}: {
		ref?: HTMLDivElement | null;
		children: Snippet;
		class?: string;
	} = $props();
</script>

<div bind:this={ref} class={cn('flex flex-col space-y-1.5', className)}>
	{@render children()}
</div>
