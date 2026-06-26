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

<div bind:this={ref} class={cn('flex justify-end gap-3 pt-4', className)}>
	{@render children()}
</div>
